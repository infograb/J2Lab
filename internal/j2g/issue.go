package j2g

import (
	"fmt"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

// comment -> comments : GitLab 작성자는 API owner이지만, 텍스트로 Jira 작성자를 표현
func convertToGitLabComment(jiraComment *jira.Comment) *gitlab.CreateIssueNoteOptions {
	body := fmt.Sprintf("%s\n\nauthored by %s at %s", jiraComment.Body, jiraComment.Author.DisplayName, jiraComment.Created)
	created, err := time.Parse("2006-01-02T15:04:05.000-0700", jiraComment.Created)
	if err != nil {
		log.Fatalf("Error parsing time: %s", err)
	}

	return &gitlab.CreateIssueNoteOptions{
		Body:      &body,
		CreatedAt: &created,
	}
}

func ConvertJiraIssueToGitLabIssue(gl *gitlab.Client, jr *jira.Client, pid interface{}, jiraIssue *jira.Issue) *gitlab.Issue {
	// TODO: epic -> epic : GitLab 프로젝트는 반드시 상위 그룹이 있어야 한다.
	//? 어느 부모에 에픽을 넣어야 하지?
	// gitlabCreateIssueOptions.EpicID

	gitlabCreateIssueOptions := &gitlab.CreateIssueOptions{
		Title:       &jiraIssue.Fields.Summary,
		Description: &jiraIssue.Fields.Description, // TODO: Jira ADF -> GitLab Markdown
		CreatedAt:   (*time.Time)(&jiraIssue.Fields.Created),
		DueDate:     (*gitlab.ISOTime)(&jiraIssue.Fields.Duedate),
		// Weight: StoryPoint,
		// MilestoneID: &fixMilestone.ID,
		// EpicID: ,
		Labels: convertJiraToGitLabLabels(gl, jr, pid, jiraIssue),
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

	//* 이슈를 생성합니다.
	gitlabIssue, _, err := gl.Issues.CreateIssue(pid, gitlabCreateIssueOptions)
	if err != nil {
		log.Fatalf("Error creating GitLab issue: %s", err)
	}

	//* Comment -> Comment
	// TODO : Jira ADF -> GitLab Markdown
	for _, jiraComment := range jiraIssue.Fields.Comments.Comments {
		_, _, err := gl.Notes.CreateIssueNote(pid, gitlabIssue.IID, convertToGitLabComment(jiraComment))
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
		log.Infof("Closing issue: %d", gitlabIssue.IID)
		gl.Issues.UpdateIssue(pid, gitlabIssue.IID, &gitlab.UpdateIssueOptions{
			StateEvent: gitlab.String("close"),
			UpdatedAt:  (*time.Time)(&jiraIssue.Fields.Resolutiondate), // 적용안됨
		})
	}

	return nil
}

// 이슈를 모두 만든 후 차례로 연결
// TODO: subtasks -> tasks
// TODO: issuelinks ( jira issue title - related - jira issue title ) - relate issue
