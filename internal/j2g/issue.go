package j2g

import (
	"sync"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/config"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/gitlabx"
	"golang.org/x/sync/errgroup"
)

func ConvertJiraIssueToGitLabIssue(gl *gitlab.Client, jr *jira.Client, jiraIssue *jira.Issue, userMap UserMap, existingLabels map[string]string) (*gitlab.Issue, error) {
	log := logrus.WithField("jiraIssue", jiraIssue.Key)
	var g errgroup.Group
	g.SetLimit(5)
	mutex := sync.RWMutex{}

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting config")
	}

	pid := cfg.Project.GitLab.Issue

	labels, err := convertJiraToGitLabLabels(gl, pid, jiraIssue, existingLabels, false)
	if err != nil {
		return nil, errors.Wrap(err, "Error converting Jira labels to GitLab labels")
	}

	gitlabCreateIssueOptions := &gitlabx.CreateIssueOptions{
		Title:     &jiraIssue.Fields.Summary,
		CreatedAt: (*time.Time)(&jiraIssue.Fields.Created),
		DueDate:   (*gitlab.ISOTime)(&jiraIssue.Fields.Duedate),
		Labels:    labels,
	}

	//* Attachment for Description and Comments
	usedAttachment := make(map[string]bool)

	attachments := make(AttachmentMap)
	for _, jiraAttachment := range jiraIssue.Fields.Attachments {
		g.Go(func(jiraAttachment *jira.Attachment) func() error {
			return func() error {
				attachment, err := convertJiraAttachmentToMarkdown(gl, jr, pid, jiraAttachment)
				if err != nil {
					return errors.Wrap(err, "Error converting Jira attachment to GitLab attachment")
				}

				mutex.Lock()
				attachments[jiraAttachment.Filename] = attachment
				mutex.Unlock()
				log.Debugf("Converted attachment: %s to %s", jiraAttachment.Filename, attachment.Markdown)
				return nil
			}
		}(jiraAttachment))
	}

	if err := g.Wait(); err != nil {
		return nil, errors.Wrap(err, "Error converting Jira attachment to GitLab attachment")
	}

	//* Description -> Description
	description, usedImages, err := formatDescription(jiraIssue, userMap, attachments, true)
	if err != nil {
		return nil, errors.Wrap(err, "Error formatting description")
	}
	gitlabCreateIssueOptions.Description = description

	for _, attachment := range usedImages {
		usedAttachment[attachment] = true
	}

	//* Assignee
	if jiraIssue.Fields.Assignee != nil {
		if assignee, ok := userMap[jiraIssue.Fields.Assignee.Name]; ok {
			gitlabCreateIssueOptions.AssigneeIDs = &[]int{assignee.ID}
			gitlabCreateIssueOptions.AssigneeID = &assignee.ID
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
			log.Debugf("Unable to convert story point from Jira issue %s to GitLab weight", jiraIssue.Key)
		}
	}

	//* 이슈를 생성합니다.
	gitlabIssue, _, err := gitlabx.CreateIssue(gl, pid, gitlabCreateIssueOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating GitLab issue")
	}
	log.Debugf("Created GitLab issue: %d from Jira issue: %s", gitlabIssue.IID, jiraIssue.Key)

	//* Comment -> Comment
	for _, jiraComment := range jiraIssue.Fields.Comments.Comments {
		g.Go(func(jiraComment *jira.Comment) func() error {
			return func() error {
				note, created, usedImages, err := formatNote(jiraIssue.Key, jiraComment, userMap, attachments, true)
				if err != nil {
					return errors.Wrap(err, "Error formatting comment")
				}

				for _, attachment := range usedImages {
					mutex.Lock()
					usedAttachment[attachment] = true
					mutex.Unlock()
				}

				options := gitlab.CreateIssueNoteOptions{
					Body:      note,
					CreatedAt: created,
				}

				_, _, err = gl.Notes.CreateIssueNote(pid, gitlabIssue.IID, &options)
				if err != nil {
					return errors.Wrap(err, "Error creating note")
				}
				return nil
			}
		}(jiraComment))
	}

	if err := g.Wait(); err != nil {
		return nil, errors.Wrap(err, "Error creating GitLab issue")
	}

	//* Reamin Attachment -> Comment
	for id, markdown := range attachments {
		if used, ok := usedAttachment[id]; ok || used {
			continue
		}

		createdAt, err := time.Parse("2006-01-02T15:04:05.000-0700", markdown.CreatedAt)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing time")
		}

		g.Go(func(attachment *Attachment) func() error {
			return func() error {
				_, _, err = gl.Notes.CreateIssueNote(pid, gitlabIssue.IID, &gitlab.CreateIssueNoteOptions{
					Body:      &attachment.Markdown,
					CreatedAt: &createdAt,
				})
				if err != nil {
					return errors.Wrap(err, "Error creating note")
				}
				return nil
			}
		}(markdown))
	}

	if err := g.Wait(); err != nil {
		return nil, errors.Wrap(err, "Error creating GitLab issue")
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
