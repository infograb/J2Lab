package j2g

import (
	"fmt"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/utils"
)

func convertJiraToGitLabLabels(gl *gitlab.Client, jr *jira.Client, id interface{}, jiraIssue *jira.Issue, isGroup bool) (*gitlab.Labels, error) {
	labels := jiraIssue.Fields.Labels

	//* Issue Type
	issueType := fmt.Sprintf("type::%s", jiraIssue.Fields.Type.Name)
	label, err := createOrRetrieveLabel(gl, jr, id, issueType, jiraIssue.Fields.Type.Description, isGroup)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating Issue Type label")
	}
	labels = append(labels, label.Name)

	//* Component
	for _, jiraComponent := range jiraIssue.Fields.Components {
		name := fmt.Sprintf("component::%s", jiraComponent.Name)
		label, err := createOrRetrieveLabel(gl, jr, id, name, jiraComponent.Description, isGroup)
		if err != nil {
			return nil, errors.Wrap(err, "Error creating Component label")
		}
		labels = append(labels, label.Name)
	}

	//* Status
	status := fmt.Sprintf("status::%s", jiraIssue.Fields.Status.Name)
	label, err = createOrRetrieveLabel(gl, jr, id, status, jiraIssue.Fields.Status.Description, isGroup)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating Status label")
	}
	labels = append(labels, label.Name)

	//* Priority
	priority := fmt.Sprintf("priority::%s", jiraIssue.Fields.Priority.Name)
	label, err = createOrRetrieveLabel(gl, jr, id, priority, jiraIssue.Fields.Priority.Description, isGroup)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating Priority label")
	}
	labels = append(labels, label.Name)

	return (*gitlab.Labels)(&labels), nil
}

func createOrRetrieveLabel(gl *gitlab.Client, jr *jira.Client, id interface{}, name string, description string, isGroup bool) (*gitlab.Label, error) {
	var label *gitlab.Label
	var groupLabel *gitlab.GroupLabel
	var err error

	if isGroup {
		groupLabel, _, err = gl.GroupLabels.GetGroupLabel(id, name)
		label = (*gitlab.Label)(groupLabel)
	} else {
		label, _, err = gl.Labels.GetLabel(id, name)
	}

	if err != nil {
		gitlabCreateLabelOptions := &gitlab.CreateLabelOptions{
			Name:        &name,
			Description: &description,
			Color:       utils.RandomColor(),
		}

		var label *gitlab.Label
		var groupLabel *gitlab.GroupLabel
		var err error

		if isGroup {
			log.Debugf("Creating group label %s to %s", name, id)
			groupLabel, _, err = gl.GroupLabels.CreateGroupLabel(id, (*gitlab.CreateGroupLabelOptions)(gitlabCreateLabelOptions))
			label = (*gitlab.Label)(groupLabel)
		} else {
			label, _, err = gl.Labels.CreateLabel(id, gitlabCreateLabelOptions)
		}
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error creating label with %s", name))
		}

		log.Infof("Created label: %s", label.Name)
		return label, nil
	}

	return label, nil
}
