// models/types.go
package models

type SaveRequest struct {
	URL     string `json:"url"`
	Content string `json:"content"`
	UserID  string `json:"user_id"`
}
