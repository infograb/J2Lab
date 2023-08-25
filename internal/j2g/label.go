package j2g

import (
	"fmt"
	"math/rand"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func convertJiraToGitLabLabels(gl *gitlab.Client, jr *jira.Client, pid interface{}, jiraIssue *jira.Issue) *gitlab.Labels {
	labels := jiraIssue.Fields.Labels

	//* Issue Type
	issueType := fmt.Sprintf("type::%s", jiraIssue.Fields.Type.Name)
	label := createOrRetrieveLabel(gl, jr, pid, issueType, jiraIssue.Fields.Type.Description)
	labels = append(labels, label.Name)

	//* Component
	for _, jiraComponent := range jiraIssue.Fields.Components {
		name := fmt.Sprintf("component::%s", jiraComponent.Name)
		label := createOrRetrieveLabel(gl, jr, pid, name, jiraComponent.Description)
		labels = append(labels, label.Name)
	}

	//* Status
	status := fmt.Sprintf("status::%s", jiraIssue.Fields.Status.Name)
	label = createOrRetrieveLabel(gl, jr, pid, status, jiraIssue.Fields.Status.Description)
	labels = append(labels, label.Name)

	//* Priority
	priority := fmt.Sprintf("priority::%s", jiraIssue.Fields.Priority.Name)
	label = createOrRetrieveLabel(gl, jr, pid, priority, jiraIssue.Fields.Priority.Description)
	labels = append(labels, label.Name)

	return (*gitlab.Labels)(&labels)
}

func createOrRetrieveLabel(gl *gitlab.Client, jr *jira.Client, pid interface{}, name string, description string) *gitlab.Label {
	label, _, err := gl.Labels.GetLabel(pid, name)

	if err != nil {
		rand.Seed(time.Now().UnixNano())
		red, green, blue := rand.Intn(256), rand.Intn(256), rand.Intn(256)
		colorHex := fmt.Sprintf("#%02X%02X%02X", red, green, blue)

		// Label이 없으면 생성
		label, _, err := gl.Labels.CreateLabel(pid, &gitlab.CreateLabelOptions{
			Name:        &name,
			Description: &description,
			Color:       &colorHex,
		})
		if err != nil {
			log.Fatalf("Error creating label: %s", err)
		}

		log.Infof("Created label: %s", label.Name)
		return label
	}

	return label
}
