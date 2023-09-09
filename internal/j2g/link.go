package j2g

import (
	"fmt"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/config"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/gitlabx"
	"golang.org/x/sync/errgroup"
)

type JiraIssueLink struct {
	*jira.Issue
	gitlabIssue *gitlab.Issue
}

type JiraEpicLink struct {
	*jira.Issue
	gitlabEpic *gitlab.Epic
}

func convertLinkType(linkType string) (*string, error) {
	linkTypeMap := map[string]string{
		// Jira issue type -> GitLab issue/epic type
		"Blocks":    "blocks",
		"Cloners":   "relates_to",
		"Duplicate": "relates_to",
		"Relates":   "relates_to",
	}

	if convertedLinkType, ok := linkTypeMap[linkType]; ok {
		return &convertedLinkType, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Unknown link type: %s", linkType))
	}
}

func Link(gl *gitlab.Client, jr *jira.Client, epicLinks map[string]*JiraEpicLink, issueLinks map[string]*JiraIssueLink) error {
	var g errgroup.Group
	g.SetLimit(5)

	cfg, err := config.GetConfig()
	if err != nil {
		return errors.Wrap(err, "Error getting config")
	}

	//* Find the parent Issues or Epics
	for _, jiraIssue := range issueLinks {
		pid := fmt.Sprintf("%d", jiraIssue.gitlabIssue.ProjectID)

		// Jira는 Epic의 부모 Epic이 없고, GitLab은 Epic이 다른 Epic의 부모가 될 수 있다.
		parentKey := ""

		if cfg.Project.Jira.CustomField.ParentEpic != "" {
			if parentEpic, ok := jiraIssue.Fields.Unknowns[cfg.Project.Jira.CustomField.ParentEpic]; ok {
				if parentEpic != nil {
					parentKey = parentEpic.(string)
				}
			}
		}

		if jiraIssue.Fields.Parent != nil {
			parentKey = jiraIssue.Fields.Parent.Key
		}

		if parentKey != "" {
			g.Go(func(jiraIssue *JiraIssueLink, parentKey string) func() error {
				return func() error {
					//* If this Issue has a parent Epic
					if parentEpicLink, ok := epicLinks[parentKey]; ok {
						_, _, err := gl.Issues.UpdateIssue(pid, jiraIssue.gitlabIssue.IID, &gitlab.UpdateIssueOptions{
							EpicID: &parentEpicLink.gitlabEpic.ID,
						})
						if err != nil {
							return errors.Wrap(err, fmt.Sprintf("Error linking GitLab issue %s with its parent epic %s", jiraIssue.Key, parentKey))
						}
						log.Infof("Linked issue %s(%d) to parent epic %s(%d)", jiraIssue.Key, jiraIssue.gitlabIssue.IID, parentKey, parentEpicLink.gitlabEpic.IID)
					}

					//* If this Issue has a parent Issue (Subtask)
					if parentIssueLink, ok := issueLinks[parentKey]; ok {
						parentIssueIID := fmt.Sprintf("%d", parentIssueLink.gitlabIssue.IID)
						_, _, err := gl.IssueLinks.CreateIssueLink(pid, jiraIssue.gitlabIssue.IID, &gitlab.CreateIssueLinkOptions{
							// IID: &issueLinks[innerIssueLink.OutwardIssue.Key].gitlabIssue.IID,
							TargetProjectID: gitlab.String(pid),
							TargetIssueIID:  gitlab.String(parentIssueIID),
							LinkType:        gitlab.String("blocks"),
						})
						if err != nil {
							return errors.Wrap(err, fmt.Sprintf("Error linking GitLab issue %s with its parent issue %s", jiraIssue.Key, parentKey))
						}
						log.Infof("Linked issue %s(%d) to parent issue %s(%d)", jiraIssue.Key, jiraIssue.gitlabIssue.IID, parentKey, parentIssueLink.gitlabIssue.IID)
					}
					return nil
				}
			}(jiraIssue, parentKey))
		}
	}

	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "Error Link issue with its parent")
	}

	//* Link Issue with other issues
	for _, jiraIssue := range issueLinks {
		pid := fmt.Sprintf("%d", jiraIssue.gitlabIssue.ProjectID)

		if jiraIssue.Fields.IssueLinks != nil {
			for _, innerIssueLink := range jiraIssue.Fields.IssueLinks {
				outwardIssue := innerIssueLink.OutwardIssue
				outwardType := innerIssueLink.Type.Name
				if outwardIssue == nil || outwardIssue.Fields.Type.Name == "Epic" {
					continue
				}

				if outwardIssue != nil {
					if _, ok := issueLinks[outwardIssue.Key]; ok {
						g.Go(func(jiraIssue *JiraIssueLink) func() error {
							return func() error {
								targetIssueIID := fmt.Sprintf("%d", issueLinks[outwardIssue.Key].gitlabIssue.IID)
								linkType, err := convertLinkType(outwardType)
								if err != nil {
									return errors.Wrap(err, fmt.Sprintf("Error Converting link type: %s", outwardType))
								}

								_, _, err = gl.IssueLinks.CreateIssueLink(pid, jiraIssue.gitlabIssue.IID, &gitlab.CreateIssueLinkOptions{
									TargetProjectID: &pid,
									TargetIssueIID:  &targetIssueIID,
									LinkType:        linkType,
								})
								if err != nil {
									return errors.Wrap(err, fmt.Sprintf("Error Creating Issue link from %s to %s", jiraIssue.Key, outwardIssue.Key))
								}

								log.Infof("Linked issue %s(%d) to %s(%d) with link type %s", jiraIssue.Key, jiraIssue.gitlabIssue.IID, outwardIssue.Key, issueLinks[outwardIssue.Key].gitlabIssue.IID, outwardType)
								return nil
							}
						}(jiraIssue))
					}
				}
			}
		}
	}

	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "Error Link issue with other issue")
	}

	//* Link Epic with other epics
	for _, jiraIssue := range epicLinks {
		gid := fmt.Sprintf("%d", jiraIssue.gitlabEpic.GroupID)

		if jiraIssue.Fields.IssueLinks != nil {
			for _, innerIssueLink := range jiraIssue.Fields.IssueLinks {
				outwardIssue := innerIssueLink.OutwardIssue
				outwardType := innerIssueLink.Type.Name

				// GitLab Epic은 GitLab Epic 끼리만 연결할 수 있다.
				if outwardIssue == nil || outwardIssue.Fields.Type.Name == "Issue" {
					continue
				}

				if outwardIssue != nil {
					if _, ok := issueLinks[outwardIssue.Key]; ok {
						if _, ok := epicLinks[outwardIssue.Key]; ok {
							g.Go(func(jiraIssue *JiraEpicLink) func() error {
								return func() error {
									targetEpicIID := fmt.Sprintf("%d", epicLinks[outwardIssue.Key].gitlabEpic.IID)
									linkType, err := convertLinkType(outwardType)
									if err != nil {
										return errors.Wrap(err, "Error creating GitLab epic link")
									}
									_, _, err = gitlabx.CreateEpicLink(gl, gid, jiraIssue.gitlabEpic.IID, &gitlabx.CreateEpicLinkOptions{
										TargetGroupID: &gid,
										TargetEpicIID: &targetEpicIID,
										LinkType:      linkType,
									})
									if err != nil {
										return errors.Wrap(err, "Error creating GitLab epic link")
									}

									log.Infof("Linked epic %s(%d) to %s(%d) with link type %s", jiraIssue.Key, jiraIssue.gitlabEpic.IID, outwardIssue.Key, epicLinks[outwardIssue.Key].gitlabEpic.IID, outwardType)
									return nil
								}
							}(jiraIssue))
						}
					}
				}
			}
		}
	}

	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "Error Link epic with other epic")
	}

	return nil
}
