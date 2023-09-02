package j2g

import (
	"context"
	"fmt"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/config"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/gitlabx"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/jirax"
)

type UserMap map[string]*gitlab.User // Jria Account ID to GitLab ID

// ! Entry
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

	//* JQL
	var prefixJql string
	if cfg.Project.Jira.Jql != "" {
		prefixJql = fmt.Sprintf("(%s) AND", cfg.Project.Jira.Jql)
	} else {
		prefixJql = ""
	}

	//* Get Jira Issues for Epic
	epicLinks := make(map[string]*EpicLink)
	epicJql := fmt.Sprintf("%s project=%s AND type = Epic Order by key ASC", prefixJql, jiraProjectID)
	jiraEpics, err := jirax.Unpaginate[jira.Issue](jr, func(searchOptions *jira.SearchOptions) ([]jira.Issue, *jira.Response, error) {
		return jr.Issue.Search(context.Background(), epicJql, searchOptions)
	})
	if err != nil {
		log.Fatalf("Error getting Jira issues for GitLab Epics: %s", err)
	}

	//* Get Jira Issues for Issue
	issueLinks := make(map[string]*IssueLink)
	issueJql := fmt.Sprintf("%s project=%s AND type != Epic Order by key ASC", prefixJql, jiraProjectID)
	jiraIssues, err := jirax.Unpaginate[jira.Issue](jr, func(searchOptions *jira.SearchOptions) ([]jira.Issue, *jira.Response, error) {
		return jr.Issue.Search(context.Background(), issueJql, searchOptions)
	})
	if err != nil {
		log.Fatalf("Error getting Jira issues for GitLab Issues: %s", err)
	}

	//* User Map
	userMap := newUserMap(gl, append(jiraEpics, jiraIssues...))

	//* Project Description
	_, _, err = gl.Projects.EditProject(gitlabProjectPath, &gitlab.EditProjectOptions{
		Description: gitlab.String(jiraProject.Description),
	})
	if err != nil {
		log.Fatalf("Error editing GitLab project: %s", err)
	}

	//* Project Milestones
	//* Sensitive to the title
	existingMilestones, err := gitlabx.Unpaginate[gitlab.Milestone](gl, func(opt *gitlab.ListOptions) ([]*gitlab.Milestone, *gitlab.Response, error) {
		return gl.Milestones.ListMilestones(gitlabProject.ID, &gitlab.ListMilestonesOptions{ListOptions: *opt})
	})
	if err != nil {
		log.Fatalf("Error getting GitLab milestones from GitLab: %s", err)
	}

	for _, version := range jiraProject.Versions {
		exist := false
		for _, milestone := range existingMilestones {
			if milestone.Title == version.Name {
				log.Infof("Milestone already exists: %s", version.Name)
				exist = true
				break
			}
		}

		if !exist {
			createMilestoneFromJiraVersion(jr, gl, gitlabProject.ID, &version)
		}
	}

	//* Epic
	for _, jiraEpic := range jiraEpics {
		log.Infof("Converting epic: %s", jiraEpic.Key)
		gitlabEpic := ConvertJiraIssueToGitLabEpic(gl, jr, &jiraEpic)
		epicLinks[jiraEpic.Key] = &EpicLink{&jiraEpic, gitlabEpic}
	}

	//* Issue
	for _, jiraIssue := range jiraIssues {
		log.Infof("Converting issue: %s", jiraIssue.Key)
		gitlabIssue := ConvertJiraIssueToGitLabIssue(gl, jr, &jiraIssue, userMap)
		issueLinks[jiraIssue.Key] = &IssueLink{&jiraIssue, gitlabIssue}
	}

	//* Link
	Link(gl, jr, epicLinks, issueLinks)
}
