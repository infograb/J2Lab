package j2g

import (
	"fmt"
	"log"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/gitlabx"
)

type IssueLink struct {
	jiraIssue   *jira.Issue
	gitlabIssue *gitlab.Issue
}

type EpicLink struct {
	jiraIssue  *jira.Issue
	gitlabEpic *gitlab.Epic
}

func convertLinkType(linkType string) *string {
	linkTypeMap := map[string]string{
		// Jira issue type -> GitLab issue/epic type
		"Blocks":    "blocks",
		"Cloners":   "relates_to",
		"Duplicate": "relates_to",
		"Relates":   "relates_to",
	}

	if convertedLinkType, ok := linkTypeMap[linkType]; ok {
		return &convertedLinkType
	} else {
		log.Fatalf("Unknown link type: %s", linkType)
		return nil
	}
}

func Link(gl *gitlab.Client, jr *jira.Client, epicLinks map[string]*EpicLink, issueLinks map[string]*IssueLink) {
	//* Jira Issue Parent -> GitLab Epic
	//* Jira Subtask Parent -> GitLab Issue // 단 이 경우, block link를 건다.
	for _, issueLink := range issueLinks {
		pid := fmt.Sprintf("%d", issueLink.gitlabIssue.ProjectID)
		gitlabIssue, jiraIssue := issueLink.gitlabIssue, issueLink.jiraIssue

		// Jira는 Epic의 부모 Epic이 없고, GitLab은 Epic이 다른 Epic의 부모가 될 수 있다.
		if jiraIssue.Fields.Parent != nil {
			parentKey := jiraIssue.Fields.Parent.Key
			if parentEpicLink, ok := epicLinks[parentKey]; ok {
				_, _, err := gl.Issues.UpdateIssue(pid, gitlabIssue.IID, &gitlab.UpdateIssueOptions{
					EpicID: &parentEpicLink.gitlabEpic.ID,
				})
				if err != nil {
					log.Fatalf("Error adding GitLab epic parent: %s", err)
				}
			} else if parentIssueLink, ok := issueLinks[parentKey]; ok {
				// TODO!!!
				parentIssueIID := fmt.Sprintf("%d", parentIssueLink.gitlabIssue.IID)
				_, _, err := gl.IssueLinks.CreateIssueLink(pid, gitlabIssue.IID, &gitlab.CreateIssueLinkOptions{
					// IID: &issueLinks[innerIssueLink.OutwardIssue.Key].gitlabIssue.IID,
					TargetProjectID: gitlab.String(pid),
					TargetIssueIID:  gitlab.String(parentIssueIID),
					LinkType:        gitlab.String("blocks"),
				})
				if err != nil {
					log.Fatalf("Error creating GitLab issue link: %s", err)
				}
			}
		}
	}

	//* Link Issue with other issues
	for _, issueLink := range issueLinks {
		pid := fmt.Sprintf("%d", issueLink.gitlabIssue.ProjectID)
		gitlabIssue, jiraIssue := issueLink.gitlabIssue, issueLink.jiraIssue

		if issueLink.jiraIssue.Fields.IssueLinks != nil {
			for _, innerIssueLink := range jiraIssue.Fields.IssueLinks {
				outwardIssue := innerIssueLink.OutwardIssue
				outwardType := innerIssueLink.Type.Name

				// GitLab Issue는 GitLab Issue 끼리만 연결할 수 있다.
				// TODO 아마 subtasks도 제외해야 할 수도.
				if outwardIssue == nil || outwardIssue.Fields.Type.Name == "Epic" {
					continue
				}

				if outwardIssue != nil {
					if _, ok := issueLinks[outwardIssue.Key]; ok {
						targetIssueIID := fmt.Sprintf("%d", issueLinks[outwardIssue.Key].gitlabIssue.IID)
						_, _, err := gl.IssueLinks.CreateIssueLink(pid, gitlabIssue.IID, &gitlab.CreateIssueLinkOptions{
							// IID: &issueLinks[innerIssueLink.OutwardIssue.Key].gitlabIssue.IID,
							TargetProjectID: &pid,
							TargetIssueIID:  &targetIssueIID,
							LinkType:        convertLinkType(outwardType),
						})
						if err != nil {
							log.Fatalf("Error creating GitLab issue link: %s", err)
						}
					}
				}
			}
		}
	}

	//* Link Epic with other epics
	for _, epicLink := range epicLinks {
		gid := fmt.Sprintf("%d", epicLink.gitlabEpic.GroupID)
		gitlabEpic, jiraIssue := epicLink.gitlabEpic, epicLink.jiraIssue

		if epicLink.jiraIssue.Fields.IssueLinks != nil {
			for _, innerIssueLink := range jiraIssue.Fields.IssueLinks {
				outwardIssue := innerIssueLink.OutwardIssue
				outwardType := innerIssueLink.Type.Name

				// GitLab Epic은 GitLab Epic 끼리만 연결할 수 있다.
				if outwardIssue == nil || outwardIssue.Fields.Type.Name == "Issue" {
					continue
				}

				if outwardIssue != nil {
					if _, ok := epicLinks[outwardIssue.Key]; ok {
						targetEpicIID := fmt.Sprintf("%d", epicLinks[outwardIssue.Key].gitlabEpic.IID)
						_, _, err := gitlabx.CreateEpicLink(gl, gid, gitlabEpic.IID, &gitlabx.CreateEpicLinkOptions{
							TargetGroupID: &gid,
							TargetEpicIID: &targetEpicIID,
							LinkType:      convertLinkType(outwardType),
						})
						if err != nil {
							log.Fatalf("Error creating GitLab epic link: %s", err)
						}
					}
				}
			}
		}
	}
}
