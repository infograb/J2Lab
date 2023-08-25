package j2g

import (
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func createOrRetrieveMiletone(gl *gitlab.Client, pid interface{}, option gitlab.CreateMilestoneOptions, closed bool) *gitlab.Milestone {
	milestones, _, err := gl.Milestones.ListMilestones(pid, &gitlab.ListMilestonesOptions{
		Title: option.Title,
	})
	if err != nil {
		log.Fatalf("Error getting milestone: %s", err)
	}

	if len(milestones) > 0 {
		return milestones[0]
	}

	milestone, _, err := gl.Milestones.CreateMilestone(pid, &option)
	if err != nil {
		log.Fatalf("Error creating milestone: %s", err)
	}

	if closed {
		_, _, err := gl.Milestones.UpdateMilestone(pid, milestone.ID, &gitlab.UpdateMilestoneOptions{
			StateEvent: gitlab.String("close"),
		})
		if err != nil {
			log.Fatalf("Error closing milestone: %s", err)
		}
	}

	return milestone
}

func createMilestoneFromJiraVersion(jr *jira.Client, gl *gitlab.Client, pid interface{}, jiraVersion *jira.Version) *gitlab.Milestone {
	log.Infof("Creating milestone: %s", jiraVersion.Name)

	var startDate time.Time
	if jiraVersion.StartDate != "" {
		parsedDate, err := time.Parse("2006-01-02", jiraVersion.StartDate)
		if err != nil {
			log.Fatalf("Error parsing time: %s", err)
		}
		startDate = parsedDate
	}

	var releaseDate time.Time
	if jiraVersion.ReleaseDate != "" {
		parsedDate, err := time.Parse("2006-01-02", jiraVersion.ReleaseDate)
		if err != nil {
			log.Fatalf("Error parsing time: %s", err)
		}
		releaseDate = parsedDate
	}

	option := gitlab.CreateMilestoneOptions{
		Title:       &jiraVersion.Name,
		Description: &jiraVersion.Description,
		StartDate:   (*gitlab.ISOTime)(&startDate),
		DueDate:     (*gitlab.ISOTime)(&releaseDate),
	}
	milestone := createOrRetrieveMiletone(gl, pid, option, *jiraVersion.Archived || *jiraVersion.Released)

	return milestone
}
