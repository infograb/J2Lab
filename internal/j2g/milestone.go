package j2g

import (
	"time"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func createOrRetrieveMiletone(gl *gitlab.Client, pid interface{}, option gitlab.CreateMilestoneOptions, closed bool) (*gitlab.Milestone, error) {
	milestones, _, err := gl.Milestones.ListMilestones(pid, &gitlab.ListMilestonesOptions{
		Title: option.Title,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error getting milestone")
	}

	if len(milestones) > 0 {
		return milestones[0], nil
	}

	milestone, _, err := gl.Milestones.CreateMilestone(pid, &option)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating milestone")
	}

	if closed {
		_, _, err := gl.Milestones.UpdateMilestone(pid, milestone.ID, &gitlab.UpdateMilestoneOptions{
			StateEvent: gitlab.String("close"),
		})
		if err != nil {
			return nil, errors.Wrap(err, "Error closing milestone")
		}
	}

	return milestone, nil
}

func createMilestoneFromJiraVersion(jr *jira.Client, gl *gitlab.Client, pid interface{}, jiraVersion *jira.Version) (*gitlab.Milestone, error) {
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
	milstone, err := createOrRetrieveMiletone(gl, pid, option, *jiraVersion.Archived || *jiraVersion.Released)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating milestone")
	}

	return milstone, nil
}
