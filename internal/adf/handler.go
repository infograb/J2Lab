package adf

import (
	"strings"
)

func handleBulletList(block *ADFBlock) string {
	var md strings.Builder
	for _, listItem := range block.Content {
		if listItem.Type == "listItem" && len(listItem.Content) > 0 {
			paragraph := listItem.Content[0]
			if paragraph.Type == "paragraph" && len(paragraph.Content) > 0 {
				textBlock := paragraph.Content[0]
				if textBlock.Type == "text" {
					md.WriteString("- " + textBlock.Text + "\n")
				}
			}
		}
	}
	return md.String()
}

func handleParagraph(block *ADFBlock, userMap UserMap) string {
	var md strings.Builder
	for _, content := range block.Content {
		switch content.Type {
		case "mention":
			username := userMap[content.Attrs["id"].(string)].Username
			md.WriteString("@" + username)
		case "text":
			if len(content.Marks) > 0 {
				for _, mark := range content.Marks {
					if mark.Type == "link" {
						md.WriteString("[" + content.Text + "](" + mark.Attrs["href"].(string) + ")")
					}
				}
			} else {
				md.WriteString(content.Text)
			}
		case "hardBreak":
			md.WriteString("\n")
		}
	}
	md.WriteString("\n\n")
	return md.String()
}

func handleTable(block *ADFBlock) string {
	var md strings.Builder
	// Your code to handle tables
	md.WriteString("|")
	for _, row := range block.Content {
		if row.Type == "tableRow" {
			for _, cell := range row.Content {
				if cell.Type == "tableHeader" || cell.Type == "tableCell" {
					if len(cell.Content) > 0 && cell.Content[0].Type == "paragraph" {
						if len(cell.Content[0].Content) > 0 && cell.Content[0].Content[0].Type == "text" {
							md.WriteString(" " + cell.Content[0].Content[0].Text + " |")
						}
					}
				}
			}
			md.WriteString("\n")
			if row.Content[0].Type == "tableHeader" {
				md.WriteString("|")
				for range row.Content {
					md.WriteString(" --- |")
				}
				md.WriteString("\n")
			}
		}
	}
	return md.String()
}
