package telegram

import (
	"bytes"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pantheon/core/domain"
	"pantheon/core/i18n"
	"pantheon/core/telegram/tgtest"
)

func photoCDN(t *testing.T) (*httptest.Server, []byte) {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil); err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Referer") != "https://www.olx.ua/" || !strings.Contains(r.Header.Get("Accept"), "image/jpeg") {
			t.Error("photo download must request an image with the OLX referer")
		}
		switch r.URL.Path {
		case "/missing":
			http.NotFound(w, r)
		case "/html":
			_, _ = w.Write([]byte("<!doctype html><html>blocked</html>"))
		case "/huge":
			w.Header().Set("Content-Length", "10485761")
		default:
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write(data)
		}
	}))
	t.Cleanup(s.Close)
	return s, data
}

func TestSendOfferUploadsPhotosWhenTelegramCannotFetchURLs(t *testing.T) {
	c, srv := newClient(t)
	cdn, imageData := photoCDN(t)
	srv.Queue("sendRichMessage", tgtest.Fail(400, "Bad Request: failed to get HTTP URL content"))
	o := sampleOffer(t)
	// A broken CDN image must not discard the other photos of the listing.
	o.Photos = []domain.Photo{{Link: cdn.URL + "/a"}, {Link: cdn.URL + "/missing"}, {Link: cdn.URL + "/b"}}
	opt := CardOptions{Lang: i18n.EN, Rich: true, MaxPhotos: 3, Silent: true}
	for attempt := 0; attempt < 2; attempt++ {
		id, err := c.SendOffer(ctx, "-100500", o, opt)
		if err != nil || id == 0 {
			t.Fatalf("SendOffer = %d, %v", id, err)
		}
		call, _ := srv.Last("sendRichMessage")
		if call.ChatID() != "-100500" || call.Payload["disable_notification"] != "true" || len(call.Buttons()) != 1 {
			t.Fatal("multipart upload lost chat, silent mode or keyboard")
		}
		if call.Payload["upload_0"] != string(imageData) || call.Payload["upload_2"] != string(imageData) {
			t.Fatal("multipart request must contain the actual image bytes")
		}
		rich := call.Payload["rich_message"].(map[string]any)
		media := rich["media"].([]any)
		if len(media) != 2 || strings.Count(call.Text(), "<img ") != 2 || !strings.Contains(call.Text(), "<tg-collage>") {
			t.Fatal("card must keep both downloadable photos in its collage")
		}
		for _, entry := range media {
			item := entry.(map[string]any)
			photo := item["media"].(map[string]any)
			attachment := strings.TrimPrefix(photo["media"].(string), "attach://")
			if call.Payload[attachment] != string(imageData) || !strings.Contains(call.Text(), "tg://photo?id="+item["id"].(string)) {
				t.Fatal("HTML, media and multipart attachments must reference the same photos")
			}
		}
	}
	if len(srv.Calls("sendRichMessage")) != 3 || len(srv.Calls("sendMessage")) != 0 {
		t.Fatal("after the first successful upload, subsequent cards must upload directly")
	}
}

func TestPhotoUploadDoesNotRetryUnknownOutcomes(t *testing.T) {
	cdn, _ := photoCDN(t)
	for _, status := range []int{429, 502, 403} {
		c, srv := newClient(t)
		srv.Queue("sendRichMessage", tgtest.Fail(400, "failed to get HTTP URL content"), tgtest.Fail(status, "send failed"))
		o := sampleOffer(t)
		o.Photos = []domain.Photo{{Link: cdn.URL + "/a"}}
		if _, err := c.SendOffer(ctx, "1", o, CardOptions{Rich: true, MaxPhotos: 1}); err == nil {
			t.Fatalf("upload HTTP %d must be returned to the caller", status)
		}
		if len(srv.Calls("")) != 2 {
			t.Fatalf("upload HTTP %d must not trigger another card", status)
		}
	}
}

func TestRejectedPhotoUploadFallsBackWithoutPhotos(t *testing.T) {
	c, srv := newClient(t)
	cdn, _ := photoCDN(t)
	srv.Queue("sendRichMessage",
		tgtest.Fail(400, "failed to get HTTP URL content"),
		tgtest.Fail(400, "Bad Request: PHOTO_INVALID_DIMENSIONS"))
	o := sampleOffer(t)
	o.Photos = []domain.Photo{{Link: cdn.URL + "/a"}}
	if _, err := c.SendOffer(ctx, "1", o, CardOptions{Rich: true, MaxPhotos: 1}); err != nil {
		t.Fatal(err)
	}
	calls := srv.Calls("sendRichMessage")
	if len(calls) != 3 || strings.Contains(calls[2].Text(), "<img") || c.uploadPhotos.Load() {
		t.Fatal("rejected uploads must fall back to a photo-free card without enabling direct uploads")
	}
}

func TestDownloadPhotoRejectsNonImagesAndOversizeResponses(t *testing.T) {
	c, _ := newClient(t)
	cdn, _ := photoCDN(t)
	for _, path := range []string{"/missing", "/html", "/huge"} {
		if _, err := c.downloadPhoto(ctx, cdn.URL+path); err == nil {
			t.Errorf("download %s must be rejected", path)
		}
	}
}
