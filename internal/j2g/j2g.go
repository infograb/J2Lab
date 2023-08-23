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
	return &gitlab.CreateIssueNoteOptions{
		Body: &body,
	}
}

func ConvertJiraIssueToGitLabIssue(gl *gitlab.Client, jr *jira.Client, pid interface{}, jiraIssue *jira.Issue) *gitlab.Issue {

	// 선처리 작업
	// TODO: version -> milestone : 마일스톤을 만든 후 해당 ID를 매핑한다. //! 뭐지?
	// fixVersion -> milestone
	fixMilestone := createMilestoneFromJiraVersion(jr, gl, pid, jiraIssue.Fields.FixVersions[0].ID)

	// TODO: epic -> epic : GitLab 프로젝트는 반드시 상위 그룹이 있어야 한다.
	//? 어느 부모에 에픽을 넣어야 하지?
	// gitlabCreateIssueOptions.EpicID

	// Assignee
	assignee, err := convertJiraUserToGitLabUser(gl, jiraIssue.Fields.Assignee)
	if err != nil {
		log.Fatalf("Error converting jira user to gitlab user: %s", err)
	}

	gitlabCreateIssueOptions := &gitlab.CreateIssueOptions{
		AssigneeIDs: &[]int{assignee.ID},
		Title:       &jiraIssue.Fields.Summary,
		Description: &jiraIssue.Fields.Description,
		CreatedAt:   (*time.Time)(&jiraIssue.Fields.Created),
		DueDate:     (*gitlab.ISOTime)(&jiraIssue.Fields.Duedate),
		Labels:      (*gitlab.Labels)(&jiraIssue.Fields.Labels),
		// Weight: StoryPoint,
		MilestoneID: &fixMilestone.ID,
		// EpicID: ,
	}

	// 이슈를 생성합니다.
	gitlabIssue, gitlabResponse, err := gl.Issues.CreateIssue(pid, gitlabCreateIssueOptions)
	if err != nil {
		fmt.Println(gitlabResponse)
	}

	// Create Comment on Issue
	for _, jiraComment := range jiraIssue.Fields.Comments.Comments {
		_, _, err := gl.Notes.CreateIssueNote(pid, gitlabIssue.IID, convertToGitLabComment(jiraComment))
		if err != nil {
			fmt.Println(err)
		}
	}

	// attachment -> comments의 attachment
	for _, jiraAttachment := range jiraIssue.Fields.Attachments {
		fileUrl := jiraAttachment.Content
		fileName := jiraAttachment.Filename

		createIssueNoteFromFile(gl, pid, gitlabIssue, fileUrl, fileName)
	}

	// 필요하다면 이슈 닫기
	if jiraIssue.Fields.Resolution != nil {
		gitlabIssue.ClosedAt = (*time.Time)(&jiraIssue.Fields.Resolutiondate)
	}

	return nil
}

// 이슈를 모두 만든 후 차례로 연결
// TODO: subtasks -> tasks
// TODO: issuelinks ( jira issue title - related - jira issue title ) - relate issue
