package j2g

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/config"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/utils"
)

type CreateEpicOptions struct {
	Title            *string         `url:"title,omitempty" json:"title,omitempty"`
	Description      *string         `url:"description,omitempty" json:"description,omitempty"`
	Labels           *gitlab.Labels  `url:"labels,comma,omitempty" json:"labels,omitempty"`
	StartDateIsFixed *bool           `url:"start_date_is_fixed,omitempty" json:"start_date_is_fixed,omitempty"`
	StartDateFixed   *gitlab.ISOTime `url:"start_date_fixed,omitempty" json:"start_date_fixed,omitempty"`
	DueDateIsFixed   *bool           `url:"due_date_is_fixed,omitempty" json:"due_date_is_fixed,omitempty"`
	DueDateFixed     *gitlab.ISOTime `url:"due_date_fixed,omitempty" json:"due_date_fixed,omitempty"`

	//* 라이브러리에서 지원하지 않는 추가 옵션
	Color        *string    `url:"color,omitempty" json:"color,omitempty"`
	Confidential *bool      `url:"confidential,omitempty" json:"confidential,omitempty"`
	CreatedAt    *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
	// ParentID ...
}

func parseID(id interface{}) (string, error) {
	switch v := id.(type) {
	case int:
		return strconv.Itoa(v), nil
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("invalid ID type %#v, the ID must be an int or a string", id)
	}
}

func createEpic(gl *gitlab.Client, gid interface{}, opt *CreateEpicOptions) (*gitlab.Epic, *gitlab.Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/epics", gitlab.PathEscape(group))

	req, err := gl.NewRequest(http.MethodPost, u, opt, nil)
	if err != nil {
		return nil, nil, err
	}

	e := new(gitlab.Epic)
	resp, err := gl.Do(req, e)
	if err != nil {
		return nil, resp, err
	}

	return e, resp, nil
}

func ConvertJiraIssueToGitLabEpic(gl *gitlab.Client, jr *jira.Client, jiraIssue *jira.Issue) *gitlab.Epic {
	cfg := config.GetConfig()
	gid := cfg.Project.GitLab.Epic

	gitlabCreateEpicOptions := CreateEpicOptions{
		Title:        gitlab.String(jiraIssue.Fields.Summary),
		Description:  gitlab.String(jiraIssue.Fields.Description),
		Color:        utils.RandomColor(),
		CreatedAt:    (*time.Time)(&jiraIssue.Fields.Created),
		Labels:       convertJiraToGitLabLabels(gl, jr, gid, jiraIssue, true),
		DueDateFixed: (*gitlab.ISOTime)(&jiraIssue.Fields.Duedate),
	}

	//* StartDate
	if cfg.Project.Jira.CustomField.EpicStartDate != "" {
		startDateStr := jiraIssue.Fields.Unknowns[cfg.Project.Jira.CustomField.EpicStartDate].(string)
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			log.Fatalf("Error parsing time: %s", err)
		}

		gitlabCreateEpicOptions.StartDateIsFixed = gitlab.Bool(true)
		gitlabCreateEpicOptions.StartDateFixed = (*gitlab.ISOTime)(&startDate)
	}

	//* DueDate
	// TODO

	//* 에픽을 생성합니다.
	gitlabEpic, _, err := createEpic(gl, cfg.Project.GitLab.Epic, &gitlabCreateEpicOptions)
	if err != nil {
		log.Fatalf("Error creating GitLab epic: %s", err)
	}

	//* Comment -> Comment
	// TODO : Jira ADF -> GitLab Markdown
	for _, jiraComment := range jiraIssue.Fields.Comments.Comments {
		createIssueNoteOptions := convertToGitLabComment(jiraComment)
		createEpicNoteOptions := gitlab.CreateEpicNoteOptions{
			Body: createIssueNoteOptions.Body,
		}

		_, _, err := gl.Notes.CreateEpicNote(gid, gitlabEpic.ID, &createEpicNoteOptions)
		if err != nil {
			log.Fatalf("Error creating GitLab comment with gid %s, epic ID %d: %s", gid, gitlabEpic.ID, err)
		}
	}

	//* attachment -> comments의 attachment
	// TODO: 그룹에서 attachement를 붙이는 API 없다!
	// for _, jiraAttachment := range jiraIssue.Fields.Attachments {
	// 	markdown := convertJiraAttachementToMarkdown(gl, jr, gid, jiraAttachment)
	// 	createdAt, err := time.Parse("2006-01-02T15:04:05.000-0700", jiraAttachment.Created)
	// 	if err != nil {
	// 		log.Fatalf("Error parsing time: %s", err)
	// 	}

	// 	_, _, err = gl.Notes.CreateIssueNote(gid, gitlabEpic.IID, &gitlab.CreateIssueNoteOptions{
	// 		Body:      &markdown,
	// 		CreatedAt: &createdAt,
	// 	})
	// 	if err != nil {
	// 		log.Fatalf("Error creating GitLab comment attachement: %s", err)
	// 	}
	// }

	//* Resolution -> Close issue (CloseAt)
	if jiraIssue.Fields.Resolution != nil {
		log.Infof("Closing issue: %d", gitlabEpic.IID)
		gl.Epics.UpdateEpic(gid, gitlabEpic.IID, &gitlab.UpdateEpicOptions{
			StateEvent: gitlab.String("close"),
		})
	}

	return gitlabEpic
}
