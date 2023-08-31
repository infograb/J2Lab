package adf

import (
	"encoding/base64"
)

type ADFBlock struct {
	Type    string                 `json:"type"`
	Text    string                 `json:"text,omitempty"`
	Content []ADFBlock             `json:"content,omitempty"`
	Attrs   map[string]interface{} `json:"attrs,omitempty"`
	Marks   []ADFMark              `json:"marks"`
}
type ADFMark struct {
	Type  string                 `json:"type"`
	Attrs map[string]interface{} `json:"attrs"`
}

func generateBasicToken(email string, token string) string {
	auth := email + ":" + token
	basicToken := base64.StdEncoding.EncodeToString([]byte(auth))

	return basicToken
}
