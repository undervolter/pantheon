// Package telegram is the Bot API transport: one typed call, the rich pages
// the control panel is drawn with, and the alert card.
//
// It carries no delivery policy. Rate limiting between alerts lives in the
// alert sender (see Limiter), so a burst of alerts cannot hold up a menu edit.
package telegram

// Button styles understood by the Bot API.
const (
	StylePrimary = "primary"
	StyleSuccess = "success"
	StyleDanger  = "danger"
)

// NoopData is the callback data of buttons that only display something.
const NoopData = "noop"

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
	Style        string `json:"style,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type TGUser struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code,omitempty"`
}

type TGChat struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type TGDocument struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	FileSize int64  `json:"file_size"`
}

type TGMessage struct {
	MessageID int         `json:"message_id"`
	From      TGUser      `json:"from"`
	Chat      TGChat      `json:"chat"`
	Text      string      `json:"text"`
	Document  *TGDocument `json:"document,omitempty"`
}

type CallbackQuery struct {
	ID      string     `json:"id"`
	From    TGUser     `json:"from"`
	Message *TGMessage `json:"message"`
	Data    string     `json:"data"`
}

type TGUpdate struct {
	UpdateID      int            `json:"update_id"`
	Message       *TGMessage     `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

// Btn is a plain callback button.
func Btn(text, data string) InlineKeyboardButton {
	return InlineKeyboardButton{Text: text, CallbackData: data}
}

// URLBtn opens a link.
func URLBtn(text, url string) InlineKeyboardButton {
	return InlineKeyboardButton{Text: text, URL: url}
}

func (b InlineKeyboardButton) Styled(style string) InlineKeyboardButton {
	b.Style = style
	return b
}

// Primary marks the main action of a screen.
func Primary(text, data string) InlineKeyboardButton { return Btn(text, data).Styled(StylePrimary) }

// Success marks a confirming action or a switch that is on.
func Success(text, data string) InlineKeyboardButton { return Btn(text, data).Styled(StyleSuccess) }

// Danger marks a destructive action.
func Danger(text, data string) InlineKeyboardButton { return Btn(text, data).Styled(StyleDanger) }

// Label is a button that only shows text, such as the "2 / 5" of a pager.
func Label(text string) InlineKeyboardButton { return Btn(text, NoopData) }

func Row(buttons ...InlineKeyboardButton) []InlineKeyboardButton { return buttons }

func Markup(rows ...[]InlineKeyboardButton) *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{InlineKeyboard: rows}
}
