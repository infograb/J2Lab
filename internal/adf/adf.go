package adf

type ADF struct {
	Version int         `json:"version,omitempty" structs:"version,omitempty"`
	Type    string      `json:"type,omitempty" structs:"type,omitempty"`
	Content []*ADFBlock `json:"content,omitempty" structs:"content,omitempty"`
}

type ADFBlock struct {
	Type    string                 `json:"type,omitempty" structs:"type,omitempty"`
	Text    string                 `json:"text,omitempty" structs:"text,omitempty"`
	Content []*ADFBlock            `json:"content,omitempty" structs:"content,omitempty"`
	Attrs   map[string]interface{} `json:"attrs,omitempty" structs:"attrs,omitempty"`
	Marks   []*ADFMark             `json:"marks,omitempty" structs:"marks,omitempty"`
}

type ADFMark struct {
	Type  string                 `json:"type,omitempty" structs:"type,omitempty"`
	Attrs map[string]interface{} `json:"attrs,omitempty" structs:"attrs,omitempty"`
}
