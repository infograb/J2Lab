package adf

import (
	"fmt"
	"strings"

	"github.com/xanzy/go-gitlab"
)

// Jira Account ID -> GitLab
type UserMap map[string]*gitlab.User

func AdfToMarkdown(blocks []*ADFBlock, userMap UserMap) string {
	var md strings.Builder

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
			// TODO: media.go 에서 함수 제작 후 추가
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

	return md.String()
}
