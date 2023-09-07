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
	jiraProjectID := cfg.Project.Jira.Name
	gitlabProjectPath := cfg.Project.GitLab.Issue

	jiraProject, _, err := jr.Project.Get(context.Background(), jiraProjectID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error getting Jira project: %s", jiraProjectID))
	}

	gitlabProject, _, err := gl.Projects.GetProject(gitlabProjectPath, nil)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error getting GitLab project: %s", gitlabProjectPath))
	}

	//* Get Jira Issues
	jiraEpics, jiraIssues, err := GetJiraIssues(jr, jiraProjectID, cfg.Project.Jira.Jql)
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
			return errors.Errorf("User %s with id %s is not a member of GitLab project %s", user.Username, user.ID, gitlabProjectPath)
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
			g.Go(func(version jira.Version) func() error {
				return func() error {
					_, err := createMilestoneFromJiraVersion(jr, gl, gitlabProject.ID, &version)
					if err != nil {
						return errors.Wrap(err, "Error creating GitLab milestone")
					}
					return nil
				}
			}(version))
		}
	}

	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "Error creating GitLab milestones")
	}

	epicLinks := make(map[string]*JiraEpicLink)
	issueLinks := make(map[string]*JiraIssueLink)

	//* Epic
	log.Infof("Converting %d epics", len(jiraEpics))
	for _, jiraEpic := range jiraEpics {
		g.Go(func(epic *jira.Issue) func() error {
			return func() error {
				log.Infof("Converting epic: %s", epic.Key)
				gitlabEpic, err := ConvertJiraIssueToGitLabEpic(gl, jr, epic, userMap)
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
				gitlabIssue, err := ConvertJiraIssueToGitLabIssue(gl, jr, jiraIssue, userMap)
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

	return nil
}
