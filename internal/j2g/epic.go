/*
 * This file is part of the InfoGrab project.
 *
 * Copyright (C) 2023 InfoGrab
 *
 * This program is free software: you can redistribute it and/or modify it
 * it is available under the terms of the GNU Lesser General Public License
 * by the Free Software Foundation, either version 3 of the License or by the Free Software Foundation
 * (at your option) any later version.
 */

package j2g

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb-public/j2lab/internal/config"
	"gitlab.com/infograb-public/j2lab/internal/gitlabx"
	"gitlab.com/infograb-public/j2lab/internal/utils"
	"golang.org/x/sync/errgroup"
)

func ConvertJiraIssueToGitLabEpic(gl *gitlab.Client, jr *jira.Client, jiraIssue *jira.Issue, userMap UserMap, existingLabels map[string]string) (*gitlab.Epic, error) {
	log := logrus.WithField("jiraEpic", jiraIssue.Key)
	var g errgroup.Group
	g.SetLimit(5)
	mutex := sync.RWMutex{}

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting config")
	}

	gid := cfg.GitLab.Epic

	labels, err := convertJiraToGitLabLabels(gl, gid, jiraIssue, existingLabels, true)
	if err != nil {
		return nil, errors.Wrap(err, "Error converting Jira labels to GitLab labels")
	}

	gitlabCreateEpicOptions := gitlabx.CreateEpicOptions{
		Title:        gitlab.String(jiraIssue.Fields.Summary),
		Color:        utils.RandomColor(),
		CreatedAt:    (*time.Time)(&jiraIssue.Fields.Created),
		Labels:       labels,
		DueDateFixed: (*gitlab.ISOTime)(&jiraIssue.Fields.Duedate),
	}

	//* Attachment for Description and Comments
	//! Epic Attachment는 API가 없는 관계로 우회한다.
	// 1. cfg.Project.GitLab.Issue 프로젝트에 attachement를 붙인다.
	// 2. 결과 markdown을 절대 경로로 바꾼 후 epic description에 붙인다
	pid := cfg.GitLab.Issue
	usedAttachment := make(map[string]bool)

	attachments := make(map[string]*Attachment) // Filename -> Markdown
	for _, jiraAttachment := range jiraIssue.Fields.Attachments {
		g.Go(func(jiraAttachment *jira.Attachment) func() error {
			return func() error {
				attachment, err := convertJiraAttachmentToMarkdown(gl, jr, pid, jiraAttachment)
				if err != nil {
					return errors.Wrap(err, "Error converting Jira attachment to GitLab attachment")
				}

				regexp := regexp.MustCompile(`!\[(.+)\]\((.+)\)`)
				matches := regexp.FindStringSubmatch(attachment.Markdown)

				if len(matches) != 3 {
					return errors.Wrap(err, "Error parsing markdown")
				}

				alt := matches[1]
				url := matches[2]

				absUrl := fmt.Sprintf("%s/%s/%s", cfg.GitLab.Host, cfg.GitLab.Issue, url)

				mutex.Lock()
				attachments[jiraAttachment.Filename] = &Attachment{
					Markdown:  fmt.Sprintf("![%s](%s)", alt, absUrl),
					Filename:  attachment.Filename,
					CreatedAt: attachment.CreatedAt,
					Alt:       alt,
					URL:       absUrl,
				}
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
	gitlabCreateEpicOptions.Description = description

	for _, attachment := range usedImages {
		usedAttachment[attachment] = true
	}

	//* StartDate
	if cfg.Jira.CustomField.EpicStartDate != "" {
		startDateStr, ok := jiraIssue.Fields.Unknowns[cfg.Jira.CustomField.EpicStartDate].(string)
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
	if jiraIssue.Fields.Duedate != (jira.Date{}) {
		gitlabCreateEpicOptions.DueDateIsFixed = gitlab.Bool(true)
		gitlabCreateEpicOptions.DueDateFixed = (*gitlab.ISOTime)(&jiraIssue.Fields.Duedate)
	}

	//* 에픽을 생성합니다.
	gitlabEpic, _, err := gitlabx.CreateEpic(gl, cfg.GitLab.Epic, &gitlabCreateEpicOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating GitLab epic")
	}
	log.Debugf("Created GitLab epic: %d from Jira issue: %s", gitlabEpic.IID, jiraIssue.Key)

	//* Comment -> Comment
	for _, jiraComment := range jiraIssue.Fields.Comments.Comments {
		g.Go(func(jiraComment *jira.Comment) func() error {
			return func() error {
				body, _, usedImages, err := formatNote(jiraIssue.Key, jiraComment, userMap, attachments, true)
				if err != nil {
					return errors.Wrap(err, "Error formatting comment")
				}

				for _, attachment := range usedImages {
					mutex.Lock()
					usedAttachment[attachment] = true
					mutex.Unlock()
				}

				createEpicNoteOptions := gitlab.CreateEpicNoteOptions{
					Body: body,
				}

				_, _, err = gl.Notes.CreateEpicNote(gid, gitlabEpic.ID, &createEpicNoteOptions)
				if err != nil {
					return errors.Wrap(err, "Error creating note")
				}
				return nil
			}
		}(jiraComment))
	}

	if err := g.Wait(); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error creating GitLab comment with gid %s, epic ID %d", gid, gitlabEpic.ID))
	}

	//* Reamin Attachment -> Comment
	for id, markdown := range attachments {
		if used, ok := usedAttachment[id]; ok || used {
			continue
		}

		g.Go(func(markdown *Attachment) func() error {
			return func() error {
				_, _, err = gl.Notes.CreateEpicNote(gid, gitlabEpic.ID, &gitlab.CreateEpicNoteOptions{
					Body: &markdown.Markdown,
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
		gl.Epics.UpdateEpic(gid, gitlabEpic.IID, &gitlab.UpdateEpicOptions{
			StateEvent: gitlab.String("close"),
		})
		log.Debugf("Closed GitLab epic: %d", gitlabEpic.IID)
	}

	return gitlabEpic, nil
}
