/*
 * This file is part of the InfoGrab project.
 *
 * Copyright (C) 2023 InfoGrab
 *
 * This program is free software: you can redistribute it and/or modify it
 * it is available under the terms of the GNU Lesser General Public License
 * by the Free Software Foundation, either version 3 of the License or by the Free Software Foundation
 * (at your option) any later version.
 */

package j2g

import (
	"context"
	"fmt"
	"sync"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/config"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/gitlabx"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/jirax"
	"golang.org/x/sync/errgroup"
)

func GetJiraIssues(jr *jira.Client, jiraProjectID string, jql string) ([]*jira.Issue, []*jira.Issue, error) {
	//* JQL
	var prefixJql string
	if jql != "" {
		prefixJql = fmt.Sprintf("(%s) AND", jql)
	} else {
		prefixJql = ""
	}

	//* Get Jira Issues for Epic
	epicJql := fmt.Sprintf("%s project = %s AND type = Epic Order by key ASC", prefixJql, jiraProjectID)
	jiraEpics, err := jirax.UnpaginateIssue(jr, epicJql)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error getting Jira issues for GitLab Epics")
	}

	//* Get Jira Issues for Issue
	issueJql := fmt.Sprintf("%s project = %s AND type != Epic Order by key ASC", prefixJql, jiraProjectID)
	jiraIssues, err := jirax.UnpaginateIssue(jr, issueJql)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error getting Jira issues for GitLab Issues")
	}

	return jiraEpics, jiraIssues, nil
}

// ! Entry
func ConvertByProject(gl *gitlab.Client, jr *jira.Client) error {
	var g errgroup.Group
	g.SetLimit(5)
	mutex := sync.RWMutex{}

	cfg, err := config.GetConfig()
	if err != nil {
		return errors.Wrap(err, "Error getting config")
	}

	//* Get Project Information
	jiraProjectID := cfg.Jira.Name
	gitlabProjectPath := cfg.GitLab.Issue

	jiraProject, _, err := jr.Project.Get(context.Background(), jiraProjectID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error getting Jira project: %s", jiraProjectID))
	}

	gitlabProject, _, err := gl.Projects.GetProject(gitlabProjectPath, nil)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error getting GitLab project: %s", gitlabProjectPath))
	}

	//* Get Jira Issues
	jiraEpics, jiraIssues, err := GetJiraIssues(jr, jiraProjectID, cfg.Jira.Jql)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error getting Jira issues: %s", jiraProjectID))
	}

	//* User Map
	userMap, err := newUserMap(gl, append(jiraEpics, jiraIssues...), cfg.Users)
	if err != nil {
		return errors.Wrap(err, "Error creating user map")
	}

	//* Check if Users are members of GitLab project
	// TODO : Unpaginate

	gitlabProjectMembers, err := gitlabx.Unpaginate[gitlab.ProjectMember](gl, func(opt *gitlab.ListOptions) ([]*gitlab.ProjectMember, *gitlab.Response, error) {
		return gl.ProjectMembers.ListAllProjectMembers(gitlabProjectPath, &gitlab.ListProjectMembersOptions{ListOptions: *opt})
	})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error getting GitLab project members: %s", gitlabProjectPath))
	}

	for _, user := range userMap {
		exist := false
		for _, member := range gitlabProjectMembers {
			if member.Username == user.Username {
				exist = true
				break
			}
		}

		if !exist {
			return errors.Errorf("User %s with id %d is not a member of GitLab project %s", user.Username, user.ID, gitlabProjectPath)
		}
	}

	//* Project Description
	_, _, err = gl.Projects.EditProject(gitlabProjectPath, &gitlab.EditProjectOptions{
		Description: gitlab.String(jiraProject.Description),
	})
	if err != nil {
		return errors.Wrap(err, "Error editing GitLab project: %s")
	}

	//* Project Milestones
	//* Sensitive to the title
	existingMilestones, err := gitlabx.Unpaginate[gitlab.Milestone](gl, func(opt *gitlab.ListOptions) ([]*gitlab.Milestone, *gitlab.Response, error) {
		return gl.Milestones.ListMilestones(gitlabProject.ID, &gitlab.ListMilestonesOptions{ListOptions: *opt})
	})
	if err != nil {
		return errors.Wrap(err, "Error getting GitLab milestones from GitLab: %s")
	}

	milestones := make(map[string]*Milestone)
	for _, version := range jiraProject.Versions {
		jiraVersion := version
		exist := false
		for _, gitlabMilesone := range existingMilestones {
			if gitlabMilesone.Title == version.Name {
				log.Infof("Milestone already exists: %s", version.Name)
				milestones[version.Name] = &Milestone{gitlabMilesone, &jiraVersion}
				exist = true
				break
			}
		}

		if !exist {
			g.Go(func(version jira.Version) func() error {
				return func() error {
					milestone, err := createMilestoneFromJiraVersion(jr, gl, gitlabProject.ID, &version)
					if err != nil {
						return errors.Wrap(err, "Error creating GitLab milestone")
					}

					mutex.Lock()
					milestones[version.Name] = milestone
					mutex.Unlock()
					return nil
				}
			}(version))
		}
	}

	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "Error creating GitLab milestones")
	}

	//* Project and Group Labels
	existingGroupLabels := make(map[string]string)
	existingProjectLabels := make(map[string]string)

	gruopLabels, err := gitlabx.Unpaginate[gitlab.GroupLabel](gl, func(opt *gitlab.ListOptions) ([]*gitlab.GroupLabel, *gitlab.Response, error) {
		return gl.GroupLabels.ListGroupLabels(cfg.GitLab.Epic, &gitlab.ListGroupLabelsOptions{
			ListOptions:              *opt,
			IncludeAncestorGroups:    gitlab.Bool(true),
			IncludeDescendantGrouops: gitlab.Bool(true),
			OnlyGroupLabels:          gitlab.Bool(true),
		})
	})
	if err != nil {
		return errors.Wrap(err, "Error getting GitLab group labels from GitLab")
	}

	for _, label := range gruopLabels {
		existingGroupLabels[label.Name] = label.Name
	}

	projectLabels, err := gitlabx.Unpaginate[gitlab.Label](gl, func(opt *gitlab.ListOptions) ([]*gitlab.Label, *gitlab.Response, error) {
		return gl.Labels.ListLabels(gitlabProject.ID, &gitlab.ListLabelsOptions{ListOptions: *opt,
			IncludeAncestorGroups: gitlab.Bool(true),
		})
	})
	if err != nil {
		return errors.Wrap(err, "Error getting GitLab project labels from GitLab")
	}

	for _, label := range projectLabels {
		existingProjectLabels[label.Name] = label.Name
	}

	//* Main Game
	epicLinks := make(map[string]*JiraEpicLink)
	issueLinks := make(map[string]*JiraIssueLink)

	//* Epic
	log.Infof("Converting %d epics", len(jiraEpics))
	for _, jiraEpic := range jiraEpics {
		g.Go(func(epic *jira.Issue) func() error {
			return func() error {
				log.Infof("Converting epic: %s", epic.Key)
				gitlabEpic, err := ConvertJiraIssueToGitLabEpic(gl, jr, epic, userMap, existingGroupLabels)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("Error converting epic: %s", epic.Key))
				}

				mutex.Lock()
				epicLinks[epic.Key] = &JiraEpicLink{epic, gitlabEpic}
				mutex.Unlock()

				return nil
			}
		}(jiraEpic))
	}

	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "Error converting epic")
	}

	//* Issue
	log.Infof("Converting %d issues", len(jiraIssues))
	for _, jiraIssue := range jiraIssues {
		g.Go(func(jiraIssue *jira.Issue) func() error {
			return func() error {
				log.Infof("Converting issue: %s", jiraIssue.Key)
				gitlabIssue, err := ConvertJiraIssueToGitLabIssue(gl, jr, jiraIssue, userMap, existingProjectLabels, milestones)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("Error converting issue: %s", jiraIssue.Key))
				}

				mutex.Lock()
				issueLinks[jiraIssue.Key] = &JiraIssueLink{jiraIssue, gitlabIssue}
				mutex.Unlock()

				return nil
			}
		}(jiraIssue))
	}

	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "Error converting issue")
	}

	//* Link
	err = Link(gl, jr, epicLinks, issueLinks)
	if err != nil {
		return errors.Wrap(err, "Error linking")
	}

	//* Close Milestone
	for _, milestone := range milestones {
		if *milestone.JiraVersion.Archived || *milestone.JiraVersion.Released {
			_, _, err := gl.Milestones.UpdateMilestone(gitlabProject.ID, milestone.ID, &gitlab.UpdateMilestoneOptions{
				StateEvent: gitlab.String("close"),
			})
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Error closing milestone: %s", milestone.JiraVersion.Name))
			}
		}
	}

	log.Infof("You are successfully migrated %s to %s", jiraProjectID, gitlabProjectPath)

	return nil
}
