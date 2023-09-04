package j2g

import (
	"fmt"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/adf"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/config"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/gitlabx"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/jirax"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/utils"
)

func ConvertJiraIssueToGitLabEpic(gl *gitlab.Client, jr *jira.Client, jiraIssue *jirax.Issue, userMap UserMap) (*gitlab.Epic, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting config")
	}

	gid := cfg.Project.GitLab.Epic

	description, err := formatDescription(jiraIssue.Key, jiraIssue.Fields.Description, []*adf.Media{}, userMap, false)
	if err != nil {
		return nil, errors.Wrap(err, "Error formatting description")
	}

	labels, err := convertJiraToGitLabLabels(gl, jr, gid, jiraIssue, true)
	if err != nil {
		return nil, errors.Wrap(err, "Error converting Jira labels to GitLab labels")
	}

	gitlabCreateEpicOptions := gitlabx.CreateEpicOptions{
		Title:        gitlab.String(jiraIssue.Fields.Summary),
		Description:  description,
		Color:        utils.RandomColor(),
		CreatedAt:    (*time.Time)(&jiraIssue.Fields.Created),
		Labels:       labels,
		DueDateFixed: (*gitlab.ISOTime)(&jiraIssue.Fields.Duedate),
	}

	//* StartDate
	if cfg.Project.Jira.CustomField.EpicStartDate != "" {
		startDateStr, ok := jiraIssue.Fields.Unknowns[cfg.Project.Jira.CustomField.EpicStartDate].(string)
		if ok {
			startDate, err := time.Parse("2006-01-02", startDateStr)
			if err != nil {
				return nil, errors.Wrap(err, "Error parsing time")
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
		return nil, errors.Wrap(err, "Error creating GitLab epic")
	}
	log.Debugf("Created GitLab epic: %d from Jira issue: %s", gitlabEpic.IID, jiraIssue.Key)

	//* Comment -> Comment
	for _, jiraComment := range jiraIssue.Fields.Comments.Comments {
		options, err := convertToIssueNoteOptions(jiraIssue.Key, jiraComment, []*adf.Media{}, userMap, false)
		if err != nil {
			return nil, errors.Wrap(err, "Error converting Jira comment to GitLab comment")
		}

		createEpicNoteOptions := gitlab.CreateEpicNoteOptions{
			Body: options.Body,
		}

		_, _, err = gl.Notes.CreateEpicNote(gid, gitlabEpic.ID, &createEpicNoteOptions)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error creating GitLab comment with gid %s, epic ID %d", gid, gitlabEpic.ID))
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

	return gitlabEpic, nil
}
