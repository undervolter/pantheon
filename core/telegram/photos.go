package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"pantheon/core/domain"
	"pantheon/core/logx"
)

var errNoDownloadedPhotos = errors.New("no listing photos could be downloaded")

type photoUpload struct {
	name string
	data []byte
}

// Bound download time and memory for an entire card, including broken CDN links.
const maxPhotoBytes = 10 << 20
const maxCardPhotoBytes = 25 << 20

func (c *Client) downloadPhoto(ctx context.Context, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "image/jpeg,image/png")
	req.Header.Set("Referer", "https://www.olx.ua/")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := c.send.Do(req)
	if err != nil {
		return nil, logx.RedactErr(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("photo download HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > maxPhotoBytes {
		return nil, errors.New("photo exceeds 10 MB")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxPhotoBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxPhotoBytes {
		return nil, errors.New("photo exceeds 10 MB")
	}
	switch http.DetectContentType(data) {
	case "image/jpeg", "image/png":
		return data, nil
	default:
		return nil, errors.New("photo response is not JPEG or PNG")
	}
}

func (c *Client) sendRichUploaded(ctx context.Context, chatID string, o domain.Offer, opt CardOptions, urls []string, km *InlineKeyboardMarkup) (int, error) {
	downloadCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	var files []photoUpload
	var refs []string
	total := 0
	for i, u := range urls {
		data, err := c.downloadPhoto(downloadCtx, u)
		if err != nil {
			logx.Warnf("[TG] download of photo %d of listing %d failed: %v", i+1, o.ID, err)
			if downloadCtx.Err() != nil {
				break
			}
			continue
		}
		if total+len(data) > maxCardPhotoBytes {
			break
		}
		total += len(data)
		name := "upload_" + strconv.Itoa(i)
		files = append(files, photoUpload{name: name, data: data})
		refs = append(refs, "attach://"+name)
	}
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	if len(files) == 0 {
		return 0, errNoDownloadedPhotos
	}
	return c.sendRich(ctx, chatID, o, opt, refs, km, files)
}

func (c *Client) callPhotoUpload(ctx context.Context, payload map[string]any, files []photoUpload, out any) error {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for key, value := range payload {
		field, ok := value.(string)
		if !ok {
			encoded, err := json.Marshal(value)
			if err != nil {
				return err
			}
			field = string(encoded)
		}
		if err := w.WriteField(key, field); err != nil {
			return err
		}
	}
	for _, file := range files {
		ext := ".jpg"
		if http.DetectContentType(file.data) == "image/png" {
			ext = ".png"
		}
		part, err := w.CreateFormFile(file.name, file.name+ext)
		if err != nil {
			return err
		}
		if _, err := part.Write(file.data); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/bot%s/sendRichMessage", c.apiBase, c.token), &body)
	if err != nil {
		return logx.RedactErr(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return c.do(req, "sendRichMessage", out)
}
