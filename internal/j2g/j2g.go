package j2g

import (
	"context"
	"fmt"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/config"
)

// Debug 모드에서는 오로지 한 번만 호출한다.
func paginateJiraIssues(jr *jira.Client, jql string, convertFunc func(*jira.Issue), debug bool) { // TODO Debug 모드 제거
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
			convertFunc(&issue)
		}
		startIndex += len(issues)
	}
}

func ConvertByProject(gl *gitlab.Client, jr *jira.Client) {
	cfg := config.GetConfig()

	//* Get Project Information
	jiraProjectID := cfg.Project.Jira.Name
	gitlabProjectPath := cfg.Project.GitLab.Issue

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

	//* Project Milestones
	for _, version := range jiraProject.Versions {
		createMilestoneFromJiraVersion(jr, gl, gitlabProject.ID, &version)
	}

	//* Project Description
	_, _, err = gl.Projects.EditProject(gitlabProjectPath, &gitlab.EditProjectOptions{
		Description: gitlab.String(jiraProject.Description),
	})
	if err != nil {
		log.Fatalf("Error editing GitLab project: %s", err)
	}

	var prefixJql string
	if cfg.Project.Jira.Jql != "" {
		prefixJql = fmt.Sprintf("%s AND", cfg.Project.Jira.Jql)
	} else {
		prefixJql = ""
	}

	var epicLinks map[string]*JiraEpicLink
	var issueLinks map[string]*JiraIssueLink

	//* Epic
	epicJql := fmt.Sprintf("%s project=%s AND type = Epic Order by key ASC", prefixJql, jiraProjectID)
	paginateJiraIssues(jr, epicJql, func(jiraIssue *jira.Issue) {
		log.Infof("Converting epic: %s", jiraIssue.Key)
		gitlabEpic := ConvertJiraIssueToGitLabEpic(gl, jr, jiraIssue)
		epicLinks[jiraIssue.Key] = &JiraEpicLink{jiraIssue, gitlabEpic}
	}, true)

	//* Issue
	issueJql := fmt.Sprintf("%s project=%s AND type != Epic Order by key ASC", prefixJql, jiraProjectID)
	paginateJiraIssues(jr, issueJql, func(jiraIssue *jira.Issue) {
		log.Infof("Converting issue: %s", jiraIssue.Key)
		gitlabIssue := ConvertJiraIssueToGitLabIssue(gl, jr, jiraIssue)
		issueLinks[jiraIssue.Key] = &JiraIssueLink{jiraIssue, gitlabIssue}
	}, true)

	//* Link
	Link(gl, jr, epicLinks, issueLinks)
}
