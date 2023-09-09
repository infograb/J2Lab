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

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb-public/j2lab/internal/utils"
)

func convertJiraToGitLabLabels(gl *gitlab.Client, id interface{}, jiraIssue *jira.Issue, existingLabels map[string]string, isGroup bool) (*gitlab.Labels, error) {
	labels := jiraIssue.Fields.Labels

	//* Issue Type
	issueType := fmt.Sprintf("type::%s", jiraIssue.Fields.Type.Name)
	if _, ok := existingLabels[issueType]; !ok {
		_, err := createLabel(gl, id, issueType, jiraIssue.Fields.Type.Description, isGroup)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error creating Issue Type label with %s", issueType))
		}
	}
	labels = append(labels, issueType)

	//* Component
	for _, jiraComponent := range jiraIssue.Fields.Components {
		name := fmt.Sprintf("component:%s", jiraComponent.Name)
		if _, ok := existingLabels[name]; !ok {
			_, err := createLabel(gl, id, name, jiraComponent.Description, isGroup)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("Error creating Component label with %s", name))
			}
		}
		labels = append(labels, name)
	}

	//* Status
	status := fmt.Sprintf("status::%s", jiraIssue.Fields.Status.Name)
	if _, ok := existingLabels[status]; !ok {
		_, err := createLabel(gl, id, status, jiraIssue.Fields.Status.Description, isGroup)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error creating Status label with %s", status))
		}
	}
	labels = append(labels, status)

	//* Priority
	priority := fmt.Sprintf("priority::%s", jiraIssue.Fields.Priority.Name)
	if _, ok := existingLabels[priority]; !ok {
		_, err := createLabel(gl, id, priority, jiraIssue.Fields.Priority.Description, isGroup)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error creating Priority label with %s", priority))
		}
	}
	labels = append(labels, priority)

	return (*gitlab.Labels)(&labels), nil
}

func createLabel(gl *gitlab.Client, id interface{}, name string, description string, isGroup bool) (*gitlab.Label, error) {
	var label *gitlab.Label
	var groupLabel *gitlab.GroupLabel
	var r *gitlab.Response
	var err error

	gitlabCreateLabelOptions := &gitlab.CreateLabelOptions{
		Name:        &name,
		Description: &description,
		Color:       utils.RandomColor(),
	}

	if isGroup {
		log.Debugf("Creating group label %s to %s", name, id)
		groupLabel, r, err = gl.GroupLabels.CreateGroupLabel(id, (*gitlab.CreateGroupLabelOptions)(gitlabCreateLabelOptions))
		label = (*gitlab.Label)(groupLabel)
	} else {
		label, r, err = gl.Labels.CreateLabel(id, gitlabCreateLabelOptions)
	}
	if r.StatusCode == 409 || r.StatusCode == 400 {
		log.Debugf("Label %s already exists", name)
	} else if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error creating label with %s", name))
	} else {
		log.Infof("Created label: %s", label.Name)
	}

	return label, nil
}
