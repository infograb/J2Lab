package j2g

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/config"
)

// comment -> comments : GitLab 작성자는 API owner이지만, 텍스트로 Jira 작성자를 표현
func convertToGitLabComment(issueKey string, jiraComment *jira.Comment) *gitlab.CreateIssueNoteOptions {
	created, err := time.Parse("2006-01-02T15:04:05.000-0700", jiraComment.Created)
	if err != nil {
		log.Fatalf("Error parsing time: %s", err)
	}

	cfg := config.GetConfig()

	commentLink := fmt.Sprintf("%s/browse/%s?focusedCommentId=%s", cfg.Jira.Host, issueKey, jiraComment.ID)
	dateFormat := fmt.Sprintf("%s at %s", created.Format("January 02, 2006"), created.Format("3:04 PM"))
	body := fmt.Sprintf("%s\n\n%s by %s [[Original](%s)]",
		jiraComment.Body, dateFormat, jiraComment.Author.DisplayName, commentLink)

	return &gitlab.CreateIssueNoteOptions{
		Body:      &body,
		CreatedAt: &created,
	}
}

// TODO: Jira ADF -> GitLab Markdown
func formatDescription(issueKey string, temp string) *string {
	cfg := config.GetConfig()
	description, err := getIssueDescriptionADF(issueKey)
	if err != nil {
		log.Fatalf("Error getting issue description: %s", err)
	}
	markdownDescription := adfToMarkdown(description.Content)
	result := fmt.Sprintf("%s\n\nImported from Jira [%s](%s/browse/%s)", markdownDescription, issueKey, cfg.Jira.Host, issueKey)
	return &result
}

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

func getIssueDescriptionADF(issueKey string) (*ADFBlock, error) {
	cfg := config.GetConfig()
	client := &http.Client{}

	url := fmt.Sprintf("%s/rest/api/3/issue/%s?fields=description", cfg.Jira.Host, issueKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	auth := cfg.Jira.Email + ":" + cfg.Jira.Token
	basicToken := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+basicToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	fields, ok := result["fields"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Missing 'fields' in the response")
	}

	descriptionData, ok := fields["description"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Missing 'description' in the response")
	}
	descriptionJSON, err := json.Marshal(descriptionData)
	if err != nil {
		return nil, err
	}
	var adfBlock ADFBlock
	err = json.Unmarshal(descriptionJSON, &adfBlock)
	if err != nil {
		return nil, err
	}
	return &adfBlock, nil
}

func adfToMarkdown(blocks []ADFBlock) string {

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
			md.WriteString(handleParagraph(block))
		case "rule":
			md.WriteString("---\n")
		case "table":
			md.WriteString(handleTable(block))
		}
	}

	return md.String()
}

func handleBulletList(block ADFBlock) string {
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

func handleParagraph(block ADFBlock) string {
	var md strings.Builder
	for _, content := range block.Content {
		switch content.Type {
		case "mention":
			// TODO: Jira Username -> GitLab Username 으로 변경 필요, 현재는 Jira Username으로 진행됨
			md.WriteString(fmt.Sprintf("%s", content.Attrs["text"].(string)))
		case "text":
			if len(content.Marks) > 0 {
				for _, mark := range content.Marks {
					if mark.Type == "link" {
						md.WriteString(fmt.Sprintf("[%s](%s)", content.Text, mark.Attrs["href"].(string)))
					}
				}
			} else {
				md.WriteString(content.Text)
			}
		case "hardBreak":
			md.WriteString("\n")
		}
	}
	md.WriteString("\n")
	return md.String()
}

func handleTable(block ADFBlock) string {
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
