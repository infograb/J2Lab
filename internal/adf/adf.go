package adf

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
