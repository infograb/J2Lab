package j2g

import (
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/adf"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/config"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/jirax"
)

func ConvertJiraIssueToGitLabIssue(gl *gitlab.Client, jr *jira.Client, jiraIssue *jirax.Issue, userMap UserMap) *gitlab.Issue {
	cfg := config.GetConfig()
	pid := cfg.Project.GitLab.Issue

	gitlabCreateIssueOptions := &gitlab.CreateIssueOptions{
		Title:     &jiraIssue.Fields.Summary,
		CreatedAt: (*time.Time)(&jiraIssue.Fields.Created),
		DueDate:   (*gitlab.ISOTime)(&jiraIssue.Fields.Duedate),
		Labels:    convertJiraToGitLabLabels(gl, jr, pid, jiraIssue, false),
	}

	//* Attachment for Description and Comments
	ch := make(chan ConvertJiraAttachmentToMarkdownResult, len(jiraIssue.Fields.Attachments))

	markdownList := make(map[string]*adf.Media) // ID -> Markdown
	for _, jiraAttachment := range jiraIssue.Fields.Attachments {
		go convertJiraAttachmentToMarkdown(gl, jr, pid, jiraAttachment, ch)
	}

	for range jiraIssue.Fields.Attachments {
		result := <-ch
		markdownList[result.ID] = &adf.Media{
			Markdown:  result.Markdown,
			CreatedAt: result.CreatedAt,
		}
	}

	//* Description -> Description
	var descriptionMediaMarkdown []*adf.Media
	for _, id := range jiraIssue.Fields.DescriptionMedia {
		if markdown, ok := markdownList[id]; ok {
			descriptionMediaMarkdown = append(descriptionMediaMarkdown, markdown)
			delete(markdownList, id)
		} else {
			log.Warnf("Unable to find media with ID %s", id)
		}
	}
	description := formatDescription(jiraIssue.Key, jiraIssue.Fields.Description, descriptionMediaMarkdown, userMap, true)
	gitlabCreateIssueOptions.Description = description

	//* Assignee
	if jiraIssue.Fields.Assignee != nil {
		if assignee, ok := userMap[jiraIssue.Fields.Assignee.AccountID]; ok {
			gitlabCreateIssueOptions.AssigneeIDs = &[]int{assignee.ID}
		}
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

	//* Reamin Attachment -> Comment
	for _, markdown := range markdownList {
		createdAt, err := time.Parse("2006-01-02T15:04:05.000-0700", markdown.CreatedAt)
		if err != nil {
			log.Fatalf("Error parsing time: %s", err)
		}

		_, _, err = gl.Notes.CreateIssueNote(pid, gitlabIssue.IID, &gitlab.CreateIssueNoteOptions{
			Body:      &markdown.Markdown,
			CreatedAt: &createdAt,
		})
		if err != nil {
			log.Fatalf("Error creating GitLab comment: %s", err)
		}
	}

	//* Comment -> Comment
	for _, jiraComment := range jiraIssue.Fields.Comments.Comments {
		var commentMediaMarkdown []*adf.Media
		for _, id := range jiraIssue.Fields.DescriptionMedia {
			if markdown, ok := markdownList[id]; ok {
				commentMediaMarkdown = append(commentMediaMarkdown, markdown)
			} else {
				log.Warnf("Unable to find media with ID %s", id)
			}
		}
		options := convertToIssueNoteOptions(jiraIssue.Key, jiraComment, commentMediaMarkdown, userMap, true)
		_, _, err := gl.Notes.CreateIssueNote(pid, gitlabIssue.IID, options)
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
