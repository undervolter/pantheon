package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"pantheon/core/logx"
	"pantheon/core/textutil"
)

const (
	defaultAPIBase = "https://api.telegram.org"

	// Long polling holds a request open for pollTimeout seconds; its client
	// must outlive that. Sends get their own client and pool so a stalled
	// poll connection cannot hold up an alert, and the other way round.
	pollTimeout       = 25
	pollClientTimeout = 40 * time.Second
	// A rich alert makes Telegram fetch up to ten photos before it answers.
	sendClientTimeout = 60 * time.Second

	defaultRetryWait = 2 * time.Second
	// A menu edit that hits a short flood limit waits it out once; a longer
	// limit is the caller's problem.
	maxUIRetryWait = 3 * time.Second
)

// APIError is a response the Bot API rejected.
type APIError struct {
	Method      string
	Code        int
	Description string
	RetryAfter  time.Duration
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s HTTP %d: %s", e.Method, e.Code, e.Description)
}

// Client talks to one bot.
type Client struct {
	token   string
	apiBase string
	send    *http.Client
	poll    *http.Client

	// plainPages is set once the server turned out to have no rich messages.
	plainPages atomic.Bool
}

// New returns a client for the given bot token. The token is registered as a
// secret, so it is stripped from every log line and transport error.
func New(token string) *Client {
	token = strings.TrimSpace(token)
	logx.AddSecret(token)
	return &Client{
		token:   token,
		apiBase: defaultAPIBase,
		send:    &http.Client{Timeout: sendClientTimeout},
		poll:    &http.Client{Timeout: pollClientTimeout},
	}
}

// WithBaseURL points the client at another Bot API server (tests, a local
// Bot API instance).
func (c *Client) WithBaseURL(base string) *Client {
	c.apiBase = strings.TrimRight(base, "/")
	return c
}

// Enabled reports whether a token is configured.
func (c *Client) Enabled() bool { return c != nil && c.token != "" }

type apiResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	Description string          `json:"description"`
}

func parseRetryAfter(body []byte) time.Duration {
	var payload struct {
		RetryAfter int `json:"retry_after"`
		Parameters struct {
			RetryAfter int `json:"retry_after"`
		} `json:"parameters"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		n := payload.Parameters.RetryAfter
		if n == 0 {
			n = payload.RetryAfter
		}
		if n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return defaultRetryWait
}

func (c *Client) do(req *http.Request, method string, out any) error {
	client := c.send
	if method == "getUpdates" {
		client = c.poll
	}
	resp, err := client.Do(req)
	if err != nil {
		// A transport error carries the request URL, and the URL carries the token.
		return fmt.Errorf("%s: %w", method, logx.RedactErr(err))
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))

	var parsed apiResponse
	_ = json.Unmarshal(raw, &parsed)
	if resp.StatusCode != http.StatusOK || !parsed.OK {
		apiErr := &APIError{Method: method, Code: resp.StatusCode, Description: parsed.Description}
		if apiErr.Description == "" {
			apiErr.Description = textutil.TruncateRunes(strings.TrimSpace(string(raw)), 300)
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			apiErr.RetryAfter = parseRetryAfter(raw)
		}
		return apiErr
	}
	if out != nil && len(parsed.Result) > 0 {
		return json.Unmarshal(parsed.Result, out)
	}
	return nil
}

// Call performs one Bot API method. It applies no retries and no rate limit.
func (c *Client) Call(ctx context.Context, method string, payload any, out any) error {
	if !c.Enabled() {
		return errors.New("telegram bot token is empty")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/bot%s/%s", c.apiBase, c.token, method), bytes.NewReader(body))
	if err != nil {
		return logx.RedactErr(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, method, out)
}

// callUI is Call for menu traffic: a short flood limit is waited out once.
func (c *Client) callUI(ctx context.Context, method string, payload any, out any) error {
	err := c.Call(ctx, method, payload, out)
	wait, limited := IsRateLimited(err)
	if !limited || wait > maxUIRetryWait {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(wait):
	}
	return c.Call(ctx, method, payload, out)
}

// ---------- error classification ----------

// IsRateLimited reports a 429 and how long the server asked to wait.
func IsRateLimited(err error) (time.Duration, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.Code == http.StatusTooManyRequests {
		d := apiErr.RetryAfter
		if d <= 0 {
			d = defaultRetryWait
		}
		return d, true
	}
	return 0, false
}

// IsDeadChat reports that the chat can no longer be written to: the user
// blocked the bot, deleted the account, or the chat never existed.
func IsDeadChat(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "user is deactivated") ||
		strings.Contains(s, "bot was blocked by the user") ||
		strings.Contains(s, "chat not found") ||
		strings.Contains(s, "bot can't initiate conversation") ||
		strings.Contains(s, "bot was kicked") ||
		strings.Contains(s, "peer_id_invalid")
}

// IsNotModified reports an edit that changed nothing, which is not a failure.
func IsNotModified(err error) bool {
	return err != nil && strings.Contains(err.Error(), "message is not modified")
}

// IsTimeout reports that the request ran out of time. Whether Telegram acted
// on it is unknown.
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// IsUnauthorized reports a rejected bot token.
func IsUnauthorized(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && (apiErr.Code == http.StatusUnauthorized || apiErr.Code == http.StatusNotFound && apiErr.Method == "getMe")
}

// IsConflict reports that another instance is polling, or a webhook is set.
func IsConflict(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Code == http.StatusConflict
}

// IsRetryable reports whether sending the same request again can succeed:
// network trouble, a flood limit or a server-side error. A request the API
// understood and refused (400, 403) is final.
func IsRetryable(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == http.StatusTooManyRequests || apiErr.Code >= 500
	}
	return true
}

func isParseError(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != http.StatusBadRequest {
		return false
	}
	s := strings.ToLower(apiErr.Description)
	return strings.Contains(s, "can't parse") || strings.Contains(s, "unsupported start tag") ||
		strings.Contains(s, "unexpected end tag") || strings.Contains(s, "can't find end tag")
}

// isPhotoRelated reports that Telegram could not use the photos of a
// message. The same card without photos is worth one more try.
func isPhotoRelated(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != http.StatusBadRequest {
		return false
	}
	s := strings.ToLower(apiErr.Description)
	return strings.Contains(s, "photo") || strings.Contains(s, "webpage_curl_failed") ||
		strings.Contains(s, "failed to get http url content") || strings.Contains(s, "wrong file identifier") ||
		strings.Contains(s, "wrong type of the web page content") || strings.Contains(s, "media")
}

// ---------- plain messages ----------

func (c *Client) sendText(ctx context.Context, chatID, text, parseMode string, km *InlineKeyboardMarkup, silent bool) (int, error) {
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}
	if silent {
		payload["disable_notification"] = true
	}
	if parseMode != "" {
		payload["parse_mode"] = parseMode
	}
	if km != nil {
		payload["reply_markup"] = km
	}
	var msg TGMessage
	if err := c.callUI(ctx, "sendMessage", payload, &msg); err != nil {
		return 0, err
	}
	return msg.MessageID, nil
}

// SendHTML sends a message with HTML markup. Only when Telegram rejects the
// markup itself is the text resent without it; a timeout or a flood limit is
// returned as is, because resending then would duplicate the message.
func (c *Client) SendHTML(ctx context.Context, chatID, text string, km *InlineKeyboardMarkup, silent bool) (int, error) {
	id, err := c.sendText(ctx, chatID, text, "HTML", km, silent)
	if err == nil || !isParseError(err) {
		return id, err
	}
	logx.Warnf("[TG] HTML rejected, sent as plain text: %v", err)
	return c.sendText(ctx, chatID, textutil.StripHTMLTags(text), "", km, silent)
}

// EditHTML replaces the text of a message. An edit that changes nothing
// counts as success.
func (c *Client) EditHTML(ctx context.Context, chatID string, messageID int, text string, km *InlineKeyboardMarkup) error {
	payload := map[string]any{
		"chat_id":                  chatID,
		"message_id":               messageID,
		"text":                     text,
		"parse_mode":               "HTML",
		"disable_web_page_preview": true,
	}
	if km != nil {
		payload["reply_markup"] = km
	}
	err := c.callUI(ctx, "editMessageText", payload, nil)
	if err == nil || IsNotModified(err) {
		return nil
	}
	if !isParseError(err) {
		return err
	}
	delete(payload, "parse_mode")
	payload["text"] = textutil.StripHTMLTags(text)
	if err := c.callUI(ctx, "editMessageText", payload, nil); err != nil && !IsNotModified(err) {
		return err
	}
	return nil
}

// AnswerCallback closes the spinner of a pressed button, optionally with a
// toast or, when alert is set, a modal.
func (c *Client) AnswerCallback(ctx context.Context, callbackID, text string, alert bool) error {
	payload := map[string]any{"callback_query_id": callbackID}
	if text != "" {
		payload["text"] = textutil.TruncateRunes(text, 190)
	}
	if alert {
		payload["show_alert"] = true
	}
	return c.Call(ctx, "answerCallbackQuery", payload, nil)
}

// ---------- bot set-up and polling ----------

func (c *Client) GetMe(ctx context.Context) (TGUser, error) {
	var u TGUser
	err := c.Call(ctx, "getMe", map[string]any{}, &u)
	return u, err
}

// DeleteWebhook makes sure long polling is allowed. Pending updates are kept.
func (c *Client) DeleteWebhook(ctx context.Context) error {
	return c.Call(ctx, "deleteWebhook", map[string]any{"drop_pending_updates": false}, nil)
}

// GetUpdates long-polls for updates after offset.
func (c *Client) GetUpdates(ctx context.Context, offset int) ([]TGUpdate, error) {
	var updates []TGUpdate
	err := c.Call(ctx, "getUpdates", map[string]any{
		"offset":          offset,
		"timeout":         pollTimeout,
		"allowed_updates": []string{"message", "callback_query"},
	}, &updates)
	return updates, err
}

// BotCommand is one entry of the command menu.
type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

// SetCommands registers the command menu; languageCode "" sets the default.
func (c *Client) SetCommands(ctx context.Context, commands []BotCommand, languageCode string) error {
	payload := map[string]any{"commands": commands}
	if languageCode != "" {
		payload["language_code"] = languageCode
	}
	return c.Call(ctx, "setMyCommands", payload, nil)
}

// ShowCommandsButton makes the chat menu button open the command list.
func (c *Client) ShowCommandsButton(ctx context.Context) error {
	return c.Call(ctx, "setChatMenuButton", map[string]any{"menu_button": map[string]string{"type": "commands"}}, nil)
}

// ---------- files ----------

// SendDocument uploads a file as a document.
func (c *Client) SendDocument(ctx context.Context, chatID, fileName string, data []byte, caption string) error {
	if !c.Enabled() {
		return errors.New("telegram bot token is empty")
	}
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("chat_id", chatID)
	if caption != "" {
		_ = w.WriteField("caption", caption)
		_ = w.WriteField("parse_mode", "HTML")
	}
	part, err := w.CreateFormFile("document", fileName)
	if err != nil {
		return err
	}
	if _, err := part.Write(data); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/bot%s/sendDocument", c.apiBase, c.token), &body)
	if err != nil {
		return logx.RedactErr(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return c.do(req, "sendDocument", nil)
}

// DownloadFile fetches a file a user sent to the bot, refusing anything
// larger than maxBytes.
func (c *Client) DownloadFile(ctx context.Context, fileID string, maxBytes int64) ([]byte, error) {
	var file struct {
		FilePath string `json:"file_path"`
		FileSize int64  `json:"file_size"`
	}
	if err := c.Call(ctx, "getFile", map[string]any{"file_id": fileID}, &file); err != nil {
		return nil, err
	}
	if file.FilePath == "" {
		return nil, errors.New("getFile: empty file_path")
	}
	if maxBytes > 0 && file.FileSize > maxBytes {
		return nil, fmt.Errorf("file is too large (%d bytes)", file.FileSize)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/file/bot%s/%s", c.apiBase, c.token, file.FilePath), nil)
	if err != nil {
		return nil, logx.RedactErr(err)
	}
	resp, err := c.send.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download: %w", logx.RedactErr(err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}
	limit := maxBytes
	if limit <= 0 {
		limit = 20 << 20
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("file is too large")
	}
	return data, nil
}
