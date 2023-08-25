package j2g

import (
	"context"
	"fmt"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/config"
)

func ConvertByProject(gl *gitlab.Client, jr *jira.Client) {
	cfg := config.GetConfig()

	for jiraProjectID, gitlabProjectPath := range cfg.Projects {
		jiraProject, _, err := jr.Project.Get(context.Background(), jiraProjectID)
		if err != nil {
			log.Fatalf("Error getting Jira project: %s", err)
		}

		gitlabProject, _, err := gl.Projects.GetProject(gitlabProjectPath, nil)
		if err != nil {
			log.Fatalf("Error getting GitLab project: %s", err)
		}

		log.Infof("Jira project: %s", jiraProject.Name)
		log.Infof("GitLab project: %s", gitlabProject.Name)

		// Get all issues from Jira project

		// 프로젝트 뽑을 때 같이 뽑히는 것들 (Jira)
		// Version
		// Component
		// IssueType
		// Description

		//* Create Project Milestones
		for _, version := range jiraProject.Versions {
			createMilestoneFromJiraVersion(jr, gl, gitlabProject.ID, &version)
		}

		_, _, err = gl.Projects.EditProject(gitlabProjectPath, &gitlab.EditProjectOptions{
			Description: gitlab.String(jiraProject.Description),
		})
		if err != nil {
			log.Fatalf("Error editing GitLab project: %s", err)
		}

		jql := fmt.Sprintf("project=%s Order by key ASC", jiraProjectID)
		startIndex := 0
		for {
			issues, _, err := jr.Issue.Search(context.Background(), jql, &jira.SearchOptions{
				StartAt: startIndex,
				Fields:  []string{"*all"},
			})
			if err != nil {
				log.Fatalf("Error getting Jira issues: %s", err)
			}

			if len(issues) == 0 {
				break
			}

			for _, issue := range issues {
				log.Infof("Converting issue: %s", issue.Key)
				ConvertJiraIssueToGitLabIssue(gl, jr, gitlabProjectPath, &issue)

				break //!
			}

			break //!

			startIndex += len(issues)
		}
	}
}
