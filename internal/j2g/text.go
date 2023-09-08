package j2g

import (
	"fmt"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/pkg/errors"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/config"
)

func textToGitLabMarkdown(text string, userMap UserMap, attachments AttachmentMap, isProject bool) (string, []string, error) {
	result, usedAttachments, err := JiraToMD(text, attachments, userMap)
	if err != nil {
		return "", nil, errors.Wrap(err, "Error converting Jira to GitLab Markdown")
	}

	return result, usedAttachments, nil
}

// comment -> comments : GitLab 작성자는 API owner이지만, 텍스트로 Jira 작성자를 표현
func formatNote(issueKey string, jiraComment *jira.Comment, userMap UserMap, attachments AttachmentMap, isProject bool) (*string, *time.Time, []string, error) {
	created, err := time.Parse("2006-01-02T15:04:05.000-0700", jiraComment.Created)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Error parsing time")
	}

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Error getting config")
	}

	commentLink := fmt.Sprintf("%s/browse/%s?focusedCommentId=%s", cfg.Jira.Host, issueKey, jiraComment.ID)
	dateFormat := fmt.Sprintf("%s at %s", created.Format("January 02, 2006"), created.Format("3:04 PM"))

	markdownBody, usedAttachments, err := textToGitLabMarkdown(jiraComment.Body, userMap, attachments, isProject)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Error converting Text to GitLab Markdown")
	}

	result := fmt.Sprintf("%s\n\n%s by %s [[Original](%s)]",
		markdownBody, dateFormat, jiraComment.Author.DisplayName, commentLink)
	return &result, &created, usedAttachments, nil
}

func formatDescription(issue *jira.Issue, userMap UserMap, attachments AttachmentMap, isProject bool) (*string, []string, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error getting config")
	}

	markdownDescription, usedAttachments, err := textToGitLabMarkdown(issue.Fields.Description, userMap, attachments, isProject)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error converting Text to GitLab Markdown")
	}
	result := fmt.Sprintf("%s\n\nImported from Jira [%s](%s/browse/%s)", markdownDescription, issue.Key, cfg.Jira.Host, issue.Key)
	return &result, usedAttachments, nil
}
