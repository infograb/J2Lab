package j2g

import (
	"fmt"
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

	// https://ig-cave.atlassian.net/browse/SSP-25?focusedCommentId=10040
	cfg := config.GetConfig()

	commentLink := fmt.Sprintf("%s/browse/%s?focusedCommentId=%s", cfg.Jira.Host, issueKey, jiraComment.ID)
	body := fmt.Sprintf("%s\n\nauthored by %s at %s [original](%s)",
		jiraComment.Body, jiraComment.Author.DisplayName, created.Format("2006-01-02 15:04:05"), commentLink)

	return &gitlab.CreateIssueNoteOptions{
		Body:      &body,
		CreatedAt: &created,
	}
}
