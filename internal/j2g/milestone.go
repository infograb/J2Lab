package j2g

import (
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func createMilestoneFromJiraVersion(jr *jira.Client, gl *gitlab.Client, pid interface{}, jiraVersion *jira.Version) *gitlab.Milestone {
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

	milestone, res, err := gl.Milestones.CreateMilestone(pid, &gitlab.CreateMilestoneOptions{
		Title:       &jiraVersion.Name,
		Description: &jiraVersion.Description,
		StartDate:   (*gitlab.ISOTime)(&startDate),
		DueDate:     (*gitlab.ISOTime)(&releaseDate),
	})
	if res.StatusCode == 400 {
		log.Debugf("Milestone already exists: %s", jiraVersion.Name)
		return nil
	} else if err != nil {
		log.Fatalf("Error creating milestone: %s", err)
	}

	// Closed Milestone if it is released or archived
	if *jiraVersion.Archived || *jiraVersion.Released {
		_, _, err := gl.Milestones.UpdateMilestone(pid, milestone.ID, &gitlab.UpdateMilestoneOptions{
			StateEvent: gitlab.String("close"),
		})
		if err != nil {
			log.Fatalf("Error closing milestone: %s", err)
		}
	}

	return milestone
}
