package adf

import (
	"fmt"
	"strings"

	"github.com/xanzy/go-gitlab"
)

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

// Jira Account ID -> GitLab
type UserMap map[string]*gitlab.User

type Media struct {
	Markdown  string
	CreatedAt string
}

func AdfToGitLabMarkdown(blocks []*ADFBlock, mediaMarkdown []*Media, userMap UserMap, isProject bool) (string, error) {
	var md strings.Builder

	mediaCount := 0

	for _, block := range blocks {
		switch block.Type {
		case "blockquote":
			md.WriteString("> " + block.Content[0].Content[0].Text + "\n")
		case "bulletList":
			md.WriteString(handleBulletList(block))
		case "codeBlock":
			md.WriteString("```" + block.Attrs["language"].(string) + "\n" + block.Content[0].Text + "\n```\n")
		case "heading":
			level := int(block.Attrs["level"].(float64))
			headingText := ""
			if len(block.Content) > 0 {
				headingText = block.Content[0].Text
			}
			md.WriteString(strings.Repeat("#", level) + " " + headingText + "\n")
		case "mediaGroup", "mediaSingle":
			if isProject {
				for range block.Content {
					md.WriteString(mediaMarkdown[mediaCount].Markdown + "\n")
					mediaCount++
				}
			} else {
				md.WriteString("Currently GitLab API v4 doesn't support group media.\n\n")
			}
		case "orderedList":
			for i, item := range block.Content {
				md.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Text))
			}
		case "panel":
			md.WriteString("> " + block.Text + "\n")
		case "paragraph":
			md.WriteString(handleParagraph(block, userMap))
		case "rule":
			md.WriteString("---\n")
		case "table":
			md.WriteString(handleTable(block))
		}
	}

	return md.String(), nil
}
