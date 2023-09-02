package j2g

import (
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/config"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/gitlabx"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/utils"
)

func ConvertJiraIssueToGitLabEpic(gl *gitlab.Client, jr *jira.Client, jiraIssue *jira.Issue) *gitlab.Epic {
	cfg := config.GetConfig()
	gid := cfg.Project.GitLab.Epic

	gitlabCreateEpicOptions := gitlabx.CreateEpicOptions{
		Title:        gitlab.String(jiraIssue.Fields.Summary),
		Description:  formatDescription(jr, jiraIssue.Key, jiraIssue.Fields.Description),
		Color:        utils.RandomColor(),
		CreatedAt:    (*time.Time)(&jiraIssue.Fields.Created),
		Labels:       convertJiraToGitLabLabels(gl, jr, gid, jiraIssue, true),
		DueDateFixed: (*gitlab.ISOTime)(&jiraIssue.Fields.Duedate),
	}

	//* StartDate
	if cfg.Project.Jira.CustomField.EpicStartDate != "" {
		startDateStr, ok := jiraIssue.Fields.Unknowns[cfg.Project.Jira.CustomField.EpicStartDate].(string)
		if ok {
			startDate, err := time.Parse("2006-01-02", startDateStr)
			if err != nil {
				log.Fatalf("Error parsing time: %s", err)
			}

			gitlabCreateEpicOptions.StartDateIsFixed = gitlab.Bool(true)
			gitlabCreateEpicOptions.StartDateFixed = (*gitlab.ISOTime)(&startDate)
		} else {
			log.Warnf("Unable to convert epic start date from Jira issue %s to GitLab start date", jiraIssue.Key)
		}
	}

	//* DueDate
	// TODO DueDate

	//* 에픽을 생성합니다.
	gitlabEpic, _, err := gitlabx.CreateEpic(gl, cfg.Project.GitLab.Epic, &gitlabCreateEpicOptions)
	if err != nil {
		log.Fatalf("Error creating GitLab epic: %s", err)
	}
	log.Debugf("Created GitLab epic: %d from Jira issue: %s", gitlabEpic.IID, jiraIssue.Key)

	//* Comment -> Comment
	for _, jiraComment := range jiraIssue.Fields.Comments.Comments {
		createIssueNoteOptions := convertToGitLabComment(jiraIssue.Key, jiraComment)
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
		gl.Epics.UpdateEpic(gid, gitlabEpic.IID, &gitlab.UpdateEpicOptions{
			StateEvent: gitlab.String("close"),
		})
		log.Debugf("Closed GitLab epic: %d", gitlabEpic.IID)
	}

	return gitlabEpic
}
