package j2g

import (
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/config"
)

func ConvertJiraIssueToGitLabIssue(gl *gitlab.Client, jr *jira.Client, jiraIssue *jira.Issue) *gitlab.Issue {
	cfg := config.GetConfig()
	pid := cfg.Project.GitLab.Issue

	gitlabCreateIssueOptions := &gitlab.CreateIssueOptions{
		Title:       &jiraIssue.Fields.Summary,
		Description: formatDescription(jiraIssue.Key, jiraIssue.Fields.Description),
		CreatedAt:   (*time.Time)(&jiraIssue.Fields.Created),
		DueDate:     (*gitlab.ISOTime)(&jiraIssue.Fields.Duedate),
		Labels:      convertJiraToGitLabLabels(gl, jr, pid, jiraIssue, false),
	}

	//* Assignee
	assignee, err := convertJiraUserToGitLabUser(gl, jiraIssue.Fields.Assignee)
	if err != nil {
		log.Fatalf("Error converting jira user to gitlab user: %s", err)
	} else if assignee != nil {
		gitlabCreateIssueOptions.AssigneeIDs = &[]int{assignee.ID}
	}

	//* Version -> Milestone
	if len(jiraIssue.Fields.FixVersions) > 0 {
		milestone := createOrRetrieveMiletone(gl, pid, gitlab.CreateMilestoneOptions{
			Title: &jiraIssue.Fields.FixVersions[0].Name,
		}, false)

		gitlabCreateIssueOptions.MilestoneID = &milestone.ID
	}

	//* Storypoint -> Weight (if custom field is provided)
	if cfg.Project.Jira.CustomField.StoryPoint != "" {
		storyPoint, ok := jiraIssue.Fields.Unknowns[cfg.Project.Jira.CustomField.StoryPoint].(float64)
		if ok {
			storyPointInt := int(storyPoint)
			gitlabCreateIssueOptions.Weight = &storyPointInt
		} else {
			log.Warnf("Unable to convert story point from Jira issue %s to GitLab weight", jiraIssue.Key)
		}
	}

	//* 이슈를 생성합니다.
	gitlabIssue, _, err := gl.Issues.CreateIssue(pid, gitlabCreateIssueOptions)
	if err != nil {
		log.Fatalf("Error creating GitLab issue: %s", err)
	}
	log.Debugf("Created GitLab issue: %d from Jira issue: %s", gitlabIssue.IID, jiraIssue.Key)

	//* Comment -> Comment
	for _, jiraComment := range jiraIssue.Fields.Comments.Comments {
		_, _, err := gl.Notes.CreateIssueNote(pid, gitlabIssue.IID, convertToGitLabComment(jiraIssue.Key, jiraComment))
		if err != nil {
			log.Fatalf("Error creating GitLab comment: %s", err)
		}
	}

	//* attachment -> comments의 attachment
	for _, jiraAttachment := range jiraIssue.Fields.Attachments {
		markdown := convertJiraAttachementToMarkdown(gl, jr, pid, jiraAttachment)
		createdAt, err := time.Parse("2006-01-02T15:04:05.000-0700", jiraAttachment.Created)
		if err != nil {
			log.Fatalf("Error parsing time: %s", err)
		}

		_, _, err = gl.Notes.CreateIssueNote(pid, gitlabIssue.IID, &gitlab.CreateIssueNoteOptions{
			Body:      &markdown,
			CreatedAt: &createdAt,
		})
		if err != nil {
			log.Fatalf("Error creating GitLab comment: %s", err)
		}
	}

	//* Resolution -> Close issue (CloseAt)
	if jiraIssue.Fields.Resolution != nil {
		gl.Issues.UpdateIssue(pid, gitlabIssue.IID, &gitlab.UpdateIssueOptions{
			StateEvent: gitlab.String("close"),
			UpdatedAt:  (*time.Time)(&jiraIssue.Fields.Resolutiondate), // 적용안됨
		})
		log.Debugf("Closed GitLab issue: %d", gitlabIssue.IID)
	}

	return gitlabIssue
}
