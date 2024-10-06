package mvc

// UserLang defines an interface for translating text with arguments.
type UserLang interface {
	Lang(text string, args ...any) string
	LangCode() string
}

// ModelMessage represents a message related to validation or other errors.
type ModelMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
