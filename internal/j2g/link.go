package j2g

import (
	"fmt"
	"log"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	gitlab "github.com/xanzy/go-gitlab"
)

type JiraIssueLink struct {
	jiraIssue   *jira.Issue
	gitlabIssue *gitlab.Issue
}

type JiraEpicLink struct {
	jiraIssue  *jira.Issue
	gitlabEpic *gitlab.Epic
}

//* 능동태만 수행한다.

//* Jira
// is blocked by
// blocks //* -> relates to
// is cloned by
// clones //* -> relates to
// is duplicated by //* -> relates to
// duplicates //* -> relates to
// relates to

//* GitLab
// relates to
// blocks
// is blocked by

func convertLinkType(linkType string) *string {
	linkTypeMap := map[string]string{
		"is blocked by":    "is blocked by",
		"blocks":           "blocks",
		"is cloned by":     "related to",
		"clones":           "relates to",
		"is duplicated by": "related to",
		"duplicates":       "relates to",
		"relates to":       "relates to",
	}

	if convertedLinkType, ok := linkTypeMap[linkType]; ok {
		return &convertedLinkType
	} else {
		log.Fatalf("Unknown link type: %s", linkType)
		return nil
	}
}

func Link(gl *gitlab.Client, jr *jira.Client, epicLinks map[string]*JiraEpicLink, issueLinks map[string]*JiraIssueLink) {
	for _, issueLink := range issueLinks {
		pid := fmt.Sprintf("%d", issueLink.gitlabIssue.ProjectID)
		gitlabIssue, jiraIssue := issueLink.gitlabIssue, issueLink.jiraIssue

		//* Jira Issue Parent -> GitLab Epic
		// Jira는 Epic의 부모 Epic이 없고, GitLab은 Epic이 다른 Epic의 부모가 될 수 있다.
		if jiraIssue.Fields.Parent != nil {

			if _, ok := epicLinks[jiraIssue.Fields.Parent.Key]; ok {
				gl.Issues.UpdateIssue(pid, gitlabIssue.IID, &gitlab.UpdateIssueOptions{
					EpicID: &epicLinks[jiraIssue.Fields.Parent.Key].gitlabEpic.ID,
				})
			}
		}

		//* Issue
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
						gl.IssueLinks.CreateIssueLink(pid, gitlabIssue.IID, &gitlab.CreateIssueLinkOptions{
							// IID: &issueLinks[innerIssueLink.OutwardIssue.Key].gitlabIssue.IID,
							TargetProjectID: &pid,
							TargetIssueIID:  &targetIssueIID,
							LinkType:        convertLinkType(outwardType),
						})
					}
				}
			}
		}
	}

	for _, epicLink := range epicLinks {
		pid := fmt.Sprintf("%d", epicLink.gitlabEpic.GroupID)
		gitlabEpic, jiraIssue := epicLink.gitlabEpic, epicLink.jiraIssue

		//* Epic
		if epicLink.jiraIssue.Fields.IssueLinks != nil {
			for _, innerIssueLink := range jiraIssue.Fields.IssueLinks {
				outwardIssue := innerIssueLink.OutwardIssue
				outwardType := innerIssueLink.Type.Name

				// GitLab Epic은 GitLab Epic 끼리만 연결할 수 있다.
				if outwardIssue == nil || outwardIssue.Fields.Type.Name != "Issue" {
					continue
				}

				if outwardIssue != nil {
					if _, ok := epicLinks[outwardIssue.Key]; ok {
						targetEpicIID := fmt.Sprintf("%d", epicLinks[outwardIssue.Key].gitlabEpic.IID)
						// gl.Epics.CreateIssueLink(pid, gitlabEpic.IID, &gitlab.CreateIssueLinkOptions{
						// 	// IID: &epicLinks[innerIssueLink.OutwardIssue.Key].gitlabEpic.IID,
						// 	TargetProjectID: &pid,
						// 	TargetIssueIID:  &targetEpicIID,
						// 	LinkType:        convertLinkType(outwardType),
						// })
						// TODO : Epic Link는 아직 지원하지 않는다. https://docs.gitlab.com/ee/api/linked_epics.html
					}
				}
			}
		}
	}
}
