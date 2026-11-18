// Package tgtest is a fake Bot API server for tests. It records every call
// and answers the way Telegram does, unless a test scripts something else.
package tgtest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// Token is the bot token tests should build their client with.
const Token = "123456789:AATESTTOKEN-not-a-real-one"

// Call is one request the server received.
type Call struct {
	Method  string
	Payload map[string]any
	Raw     string
}

// Text returns the message text or the rich page of a send or edit call.
func (c Call) Text() string {
	if rich, ok := c.Payload["rich_message"].(map[string]any); ok {
		s, _ := rich["html"].(string)
		return s
	}
	s, _ := c.Payload["text"].(string)
	return s
}

// ChatID returns chat_id as a string, however the caller encoded it.
func (c Call) ChatID() string {
	return fmt.Sprint(c.Payload["chat_id"])
}

// Buttons returns the callback data (or URL) of every inline button, row by row.
func (c Call) Buttons() []string {
	var out []string
	markup, _ := c.Payload["reply_markup"].(map[string]any)
	rows, _ := markup["inline_keyboard"].([]any)
	for _, r := range rows {
		cells, _ := r.([]any)
		for _, cell := range cells {
			b, _ := cell.(map[string]any)
			if data, ok := b["callback_data"].(string); ok && data != "" {
				out = append(out, data)
			} else if url, ok := b["url"].(string); ok {
				out = append(out, url)
			}
		}
	}
	return out
}

// ButtonTexts returns the labels of every inline button, row by row.
func (c Call) ButtonTexts() []string {
	var out []string
	markup, _ := c.Payload["reply_markup"].(map[string]any)
	rows, _ := markup["inline_keyboard"].([]any)
	for _, r := range rows {
		cells, _ := r.([]any)
		for _, cell := range cells {
			b, _ := cell.(map[string]any)
			text, _ := b["text"].(string)
			out = append(out, text)
		}
	}
	return out
}

// Reply is a scripted answer: an HTTP status and a JSON body.
type Reply struct {
	Status int
	Body   string
}

// Fail builds the error body Telegram returns.
func Fail(status int, description string) Reply {
	body, _ := json.Marshal(map[string]any{"ok": false, "error_code": status, "description": description})
	return Reply{Status: status, Body: string(body)}
}

// FloodWait builds a 429 with retry_after.
func FloodWait(seconds int) Reply {
	return Reply{Status: http.StatusTooManyRequests, Body: fmt.Sprintf(
		`{"ok":false,"error_code":429,"description":"Too Many Requests: retry after %d","parameters":{"retry_after":%d}}`, seconds, seconds)}
}

type Server struct {
	*httptest.Server

	mu       sync.Mutex
	calls    []Call
	scripted map[string][]Reply
	always   map[string]func(Call) *Reply
	nextMsg  int
	files    map[string][]byte
}

// New starts a server that is shut down with the test.
func New(t testing.TB) *Server {
	t.Helper()
	s := &Server{
		scripted: map[string][]Reply{},
		always:   map[string]func(Call) *Reply{},
		nextMsg:  100,
		files:    map[string][]byte{},
	}
	s.Server = httptest.NewServer(http.HandlerFunc(s.handle))
	t.Cleanup(s.Close)
	return s
}

// Queue makes the next calls of method get these replies, in order. Once the
// queue is empty the default answer is back.
func (s *Server) Queue(method string, replies ...Reply) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scripted[method] = append(s.scripted[method], replies...)
}

// On installs a handler for every call of method; returning nil falls
// through to the default answer.
func (s *Server) On(method string, fn func(Call) *Reply) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.always[method] = fn
}

// AddFile makes a file downloadable through getFile and returns its file_id.
func (s *Server) AddFile(name string, data []byte) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[name] = data
	return name
}

// Calls returns the recorded calls of a method, oldest first; an empty
// method returns everything.
func (s *Server) Calls(method string) []Call {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Call
	for _, c := range s.calls {
		if method == "" || c.Method == method {
			out = append(out, c)
		}
	}
	return out
}

// Sent returns every call that put text in front of a user: new messages,
// rich pages and edits.
func (s *Server) Sent() []Call {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Call
	for _, c := range s.calls {
		switch c.Method {
		case "sendMessage", "sendRichMessage", "editMessageText":
			out = append(out, c)
		}
	}
	return out
}

// Last returns the most recent call of a method.
func (s *Server) Last(method string) (Call, bool) {
	calls := s.Calls(method)
	if len(calls) == 0 {
		return Call{}, false
	}
	return calls[len(calls)-1], true
}

// Reset forgets the recorded calls.
func (s *Server) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = nil
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasPrefix(path, "/file/bot") {
		name := path[strings.LastIndex(path, "/")+1:]
		s.mu.Lock()
		data, ok := s.files[name]
		s.mu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(data)
		return
	}

	method := path[strings.LastIndex(path, "/")+1:]
	raw, _ := io.ReadAll(r.Body)
	call := Call{Method: method, Raw: string(raw), Payload: map[string]any{}}
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/") {
		r.Body = io.NopCloser(strings.NewReader(string(raw)))
		if err := r.ParseMultipartForm(8 << 20); err == nil {
			for k, v := range r.MultipartForm.Value {
				if len(v) > 0 {
					call.Payload[k] = v[0]
				}
			}
			for k, files := range r.MultipartForm.File {
				if len(files) == 0 {
					continue
				}
				f, _ := files[0].Open()
				data, _ := io.ReadAll(f)
				f.Close()
				call.Payload[k] = string(data)
				call.Payload[k+"_name"] = files[0].Filename
			}
		}
	} else {
		_ = json.Unmarshal(raw, &call.Payload)
	}

	s.mu.Lock()
	s.calls = append(s.calls, call)
	var reply *Reply
	if q := s.scripted[method]; len(q) > 0 {
		reply = &q[0]
		s.scripted[method] = q[1:]
	}
	handler := s.always[method]
	s.mu.Unlock()

	if reply == nil && handler != nil {
		reply = handler(call)
	}
	if reply == nil {
		reply = s.defaultReply(call)
	}
	status := reply.Status
	if status == 0 {
		status = http.StatusOK
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, reply.Body)
}

func (s *Server) defaultReply(call Call) *Reply {
	switch call.Method {
	case "getUpdates":
		return &Reply{Body: `{"ok":true,"result":[]}`}
	case "getMe":
		return &Reply{Body: `{"ok":true,"result":{"id":42,"first_name":"Pantheon","username":"pantheon_test_bot"}}`}
	case "getFile":
		id, _ := call.Payload["file_id"].(string)
		s.mu.Lock()
		size := len(s.files[id])
		s.mu.Unlock()
		return &Reply{Body: fmt.Sprintf(`{"ok":true,"result":{"file_id":%q,"file_path":%q,"file_size":%d}}`, id, id, size)}
	case "sendMessage", "sendRichMessage", "sendPhoto", "sendDocument":
		s.mu.Lock()
		s.nextMsg++
		id := s.nextMsg
		s.mu.Unlock()
		return &Reply{Body: fmt.Sprintf(`{"ok":true,"result":{"message_id":%d}}`, id)}
	}
	return &Reply{Body: `{"ok":true,"result":true}`}
}
