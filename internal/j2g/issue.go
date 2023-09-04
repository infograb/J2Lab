package j2g

import (
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/adf"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/config"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/jirax"
)

func ConvertJiraIssueToGitLabIssue(gl *gitlab.Client, jr *jira.Client, jiraIssue *jirax.Issue, userMap UserMap) (*gitlab.Issue, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting config")
	}

	pid := cfg.Project.GitLab.Issue

	labels, err := convertJiraToGitLabLabels(gl, jr, pid, jiraIssue, false)
	if err != nil {
		return nil, errors.Wrap(err, "Error converting Jira labels to GitLab labels")
	}

	gitlabCreateIssueOptions := &gitlab.CreateIssueOptions{
		Title:     &jiraIssue.Fields.Summary,
		CreatedAt: (*time.Time)(&jiraIssue.Fields.Created),
		DueDate:   (*gitlab.ISOTime)(&jiraIssue.Fields.Duedate),
		Labels:    labels,
	}

	//* Attachment for Description and Comments
	usedAttachment := make(map[string]bool)
	ch := make(chan Attachment, 5)

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
		log.Debugf("Converted attachment: %s to %s", result.ID, result.Markdown)
	}

	//* Description -> Description
	var descriptionMediaMarkdown []*adf.Media
	for _, id := range jiraIssue.Fields.DescriptionMedia {
		if markdown, ok := markdownList[id]; ok {
			descriptionMediaMarkdown = append(descriptionMediaMarkdown, markdown)
			usedAttachment[id] = true
		} else {
			log.Warnf("Unable to find media in Description with ID %s", id)
		}
	}
	description, err := formatDescription(jiraIssue.Key, jiraIssue.Fields.Description, descriptionMediaMarkdown, userMap, true)
	if err != nil {
		return nil, errors.Wrap(err, "Error formatting description")
	}
	gitlabCreateIssueOptions.Description = description

	//* Assignee
	if jiraIssue.Fields.Assignee != nil {
		if assignee, ok := userMap[jiraIssue.Fields.Assignee.AccountID]; ok {
			gitlabCreateIssueOptions.AssigneeIDs = &[]int{assignee.ID}
		}
	}

	//* Version -> Milestone
	if len(jiraIssue.Fields.FixVersions) > 0 {
		milestone, err := createOrRetrieveMiletone(gl, pid, gitlab.CreateMilestoneOptions{
			Title: &jiraIssue.Fields.FixVersions[0].Name,
		}, false)

		if err != nil {
			return nil, errors.Wrap(err, "Error creating GitLab milestone")
		}

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
		return nil, errors.Wrap(err, "Error creating GitLab issue")
	}
	log.Debugf("Created GitLab issue: %d from Jira issue: %s", gitlabIssue.IID, jiraIssue.Key)

	//* Comment -> Comment
	for _, jiraComment := range jiraIssue.Fields.Comments.Comments {
		var commentMediaMarkdown []*adf.Media
		for _, id := range jiraIssue.Fields.DescriptionMedia {
			if markdown, ok := markdownList[id]; ok {
				commentMediaMarkdown = append(commentMediaMarkdown, markdown)
				usedAttachment[id] = true
			} else {
				log.Warnf("Unable to find media in Comment with ID %s", id)
			}
		}
		options, err := convertToIssueNoteOptions(jiraIssue.Key, jiraComment, commentMediaMarkdown, userMap, true)
		if err != nil {
			return nil, errors.Wrap(err, "Error formatting comment")
		}

		// TODO 왜 병렬이 안돼지...
		gl.Notes.CreateIssueNote(pid, gitlabIssue.IID, options)
	}

	//* Reamin Attachment -> Comment
	for id, markdown := range markdownList {
		if used, ok := usedAttachment[id]; ok || used {
			continue
		}

		createdAt, err := time.Parse("2006-01-02T15:04:05.000-0700", markdown.CreatedAt)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing time")
		}

		// TODO 왜 병렬이 안돼지...
		_, _, err = gl.Notes.CreateIssueNote(pid, gitlabIssue.IID, &gitlab.CreateIssueNoteOptions{
			Body:      &markdown.Markdown,
			CreatedAt: &createdAt,
		})
		if err != nil {
			return nil, errors.Wrap(err, "Error creating note")
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

	return gitlabIssue, nil
}
