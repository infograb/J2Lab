package j2g

import (
	"context"
	"strconv"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func createMilestoneFromJiraVersion(jr *jira.Client, gl *gitlab.Client, pid interface{}, jiraVersionIDstr string) *gitlab.Milestone {

	jiraFixVersionID, err := strconv.Atoi(jiraVersionIDstr)
	if err != nil {
		log.Fatalf("Error converting jira fix version id: %s", err)
	}

	jiraFixVersion, _, err := jr.Version.Get(context.Background(), jiraFixVersionID)
	if err != nil {
		log.Fatalf("Error getting jira fix version: %s", err)
	}

	jiraFixVersionStartDate, err := time.Parse("2006-01-02", jiraFixVersion.StartDate)
	if err != nil {
		log.Fatal(err)
	}

	jiraFixVersionReleaseDate, err := time.Parse("2006-01-02", jiraFixVersion.StartDate)
	if err != nil {
		log.Fatal(err)
	}

	milestone, _, err := gl.Milestones.CreateMilestone(pid, &gitlab.CreateMilestoneOptions{
		Title:       &jiraFixVersion.Name,
		Description: &jiraFixVersion.Description,
		StartDate:   (*gitlab.ISOTime)(&jiraFixVersionStartDate),
		DueDate:     (*gitlab.ISOTime)(&jiraFixVersionReleaseDate),
	})
	if err != nil {
		log.Fatalf("Error creating milestone: %s", err)
	}

	return milestone
}
