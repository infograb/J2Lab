package j2g

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/adf"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/config"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/jirax"
)

// comment -> comments : GitLab 작성자는 API owner이지만, 텍스트로 Jira 작성자를 표현
func convertToIssueNoteOptions(issueKey string, jiraComment *jirax.Comment, mediaMarkdown []*adf.Media, userMap UserMap, isProject bool) (*gitlab.CreateIssueNoteOptions, error) {
	created, err := time.Parse("2006-01-02T15:04:05.000-0700", jiraComment.Created)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing time")
	}

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting config")
	}

	commentLink := fmt.Sprintf("%s/browse/%s?focusedCommentId=%s", cfg.Jira.Host, issueKey, jiraComment.ID)
	dateFormat := fmt.Sprintf("%s at %s", created.Format("January 02, 2006"), created.Format("3:04 PM"))
	formatedBody, err := formatDescription(issueKey, jiraComment.Body, mediaMarkdown, userMap, isProject)
	if err != nil {
		return nil, errors.Wrap(err, "Error formatting description")
	}

	body := fmt.Sprintf("%s\n\n%s by %s [[Original](%s)]",
		*formatedBody, dateFormat, jiraComment.Author.DisplayName, commentLink)

	return &gitlab.CreateIssueNoteOptions{
		Body:      &body,
		CreatedAt: &created,
	}, nil
}

func formatDescription(issueKey string, content *adf.ADF, mediaMarkdown []*adf.Media, userMap UserMap, isProject bool) (*string, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting config")
	}

	adfBlock := content.Content
	markdownDescription, err := adf.AdfToGitLabMarkdown(adfBlock, mediaMarkdown, adf.UserMap(userMap), isProject)
	if err != nil {
		return nil, errors.Wrap(err, "Error converting ADF to GitLab Markdown")
	}
	result := fmt.Sprintf("%s\n\nImported from Jira [%s](%s/browse/%s)", markdownDescription, issueKey, cfg.Jira.Host, issueKey)
	return &result, nil
}
