package gitlabx

import (
	"fmt"
	"net/http"
	"time"

	gitlab "github.com/xanzy/go-gitlab"
)

type CreateIssueOptions struct {
	IID                                *int            `url:"iid,omitempty" json:"iid,omitempty"`
	Title                              *string         `url:"title,omitempty" json:"title,omitempty"`
	Description                        *string         `url:"description,omitempty" json:"description,omitempty"`
	Confidential                       *bool           `url:"confidential,omitempty" json:"confidential,omitempty"`
	AssigneeIDs                        *[]int          `url:"assignee_ids,omitempty" json:"assignee_ids,omitempty"`
	MilestoneID                        *int            `url:"milestone_id,omitempty" json:"milestone_id,omitempty"`
	Labels                             *gitlab.Labels  `url:"labels,comma,omitempty" json:"labels,omitempty"`
	CreatedAt                          *time.Time      `url:"created_at,omitempty" json:"created_at,omitempty"`
	DueDate                            *gitlab.ISOTime `url:"due_date,omitempty" json:"due_date,omitempty"`
	EpicID                             *int            `url:"epic_id,omitempty" json:"epic_id,omitempty"`
	MergeRequestToResolveDiscussionsOf *int            `url:"merge_request_to_resolve_discussions_of,omitempty" json:"merge_request_to_resolve_discussions_of,omitempty"`
	DiscussionToResolve                *string         `url:"discussion_to_resolve,omitempty" json:"discussion_to_resolve,omitempty"`
	Weight                             *int            `url:"weight,omitempty" json:"weight,omitempty"`
	IssueType                          *string         `url:"issue_type,omitempty" json:"issue_type,omitempty"`

	//* 추가
	AssigneeID *int `url:"assignee_id,omitempty" json:"assignee_id,omitempty"`
}

func CreateIssue(gl *gitlab.Client, pid interface{}, opt *CreateIssueOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Issue, *gitlab.Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/issues", gitlab.PathEscape(project))

	req, err := gl.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	i := new(gitlab.Issue)
	resp, err := gl.Do(req, i)
	if err != nil {
		return nil, resp, err
	}

	return i, resp, nil
}
