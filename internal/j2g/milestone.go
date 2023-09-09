package j2g

import (
	"time"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

type Milestone struct {
	*gitlab.Milestone
	JiraVersion *jira.Version
}

func createMilestoneFromJiraVersion(jr *jira.Client, gl *gitlab.Client, pid interface{}, jiraVersion *jira.Version) (*Milestone, error) {
	log.Infof("Creating milestone: %s", jiraVersion.Name)

	var startDate time.Time
	if jiraVersion.StartDate != "" {
		parsedDate, err := time.Parse("2006-01-02", jiraVersion.StartDate)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing Start Date")
		}
		startDate = parsedDate
	}

	var releaseDate time.Time
	if jiraVersion.ReleaseDate != "" {
		parsedDate, err := time.Parse("2006-01-02", jiraVersion.ReleaseDate)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing release Date")
		}
		releaseDate = parsedDate
	}

	option := gitlab.CreateMilestoneOptions{
		Title:       &jiraVersion.Name,
		Description: &jiraVersion.Description,
		StartDate:   (*gitlab.ISOTime)(&startDate),
		DueDate:     (*gitlab.ISOTime)(&releaseDate),
	}

	milestone, _, err := gl.Milestones.CreateMilestone(pid, &option)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating milestone")
	}

	return &Milestone{
		Milestone:   milestone,
		JiraVersion: jiraVersion,
	}, nil
}
