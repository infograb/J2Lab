package jirax

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/trivago/tgo/tcontainer"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/adf"
)

type SearchResult struct {
	Expand     string  `json:"expand" structs:"expand"`
	StartAt    int     `json:"startAt" structs:"startAt"`
	MaxResults int     `json:"maxResults" structs:"maxResults"`
	Total      int     `json:"total" structs:"total"`
	Issues     []Issue `json:"issues" structs:"issues"`
}

// Issue represents a Jira issue.
type Issue struct {
	Expand         string                    `json:"expand,omitempty" structs:"expand,omitempty"`
	ID             string                    `json:"id,omitempty" structs:"id,omitempty"`
	Self           string                    `json:"self,omitempty" structs:"self,omitempty"`
	Key            string                    `json:"key,omitempty" structs:"key,omitempty"`
	Fields         *IssueFields              `json:"fields,omitempty" structs:"fields,omitempty"`
	RenderedFields *jira.IssueRenderedFields `json:"renderedFields,omitempty" structs:"renderedFields,omitempty"`
	Changelog      *jira.Changelog           `json:"changelog,omitempty" structs:"changelog,omitempty"`
	Transitions    []jira.Transition         `json:"transitions,omitempty" structs:"transitions,omitempty"`
	Names          map[string]string         `json:"names,omitempty" structs:"names,omitempty"`
}

type IssueFields struct {
	// TODO Missing fields
	//      * "workratio": -1,
	//      * "lastViewed": null,
	//      * "environment": null,
	Expand                        string                 `json:"expand,omitempty" structs:"expand,omitempty"`
	Type                          jira.IssueType         `json:"issuetype,omitempty" structs:"issuetype,omitempty"`
	Project                       jira.Project           `json:"project,omitempty" structs:"project,omitempty"`
	Environment                   string                 `json:"environment,omitempty" structs:"environment,omitempty"`
	Resolution                    *jira.Resolution       `json:"resolution,omitempty" structs:"resolution,omitempty"`
	Priority                      *jira.Priority         `json:"priority,omitempty" structs:"priority,omitempty"`
	Resolutiondate                jira.Time              `json:"resolutiondate,omitempty" structs:"resolutiondate,omitempty"`
	Created                       jira.Time              `json:"created,omitempty" structs:"created,omitempty"`
	Duedate                       jira.Date              `json:"duedate,omitempty" structs:"duedate,omitempty"`
	Watches                       *jira.Watches          `json:"watches,omitempty" structs:"watches,omitempty"`
	Assignee                      *jira.User             `json:"assignee,omitempty" structs:"assignee,omitempty"`
	Updated                       jira.Time              `json:"updated,omitempty" structs:"updated,omitempty"`
	Description                   *adf.ADF               `json:"description,omitempty" structs:"description,omitempty"`
	Summary                       string                 `json:"summary,omitempty" structs:"summary,omitempty"`
	Creator                       *jira.User             `json:"Creator,omitempty" structs:"Creator,omitempty"`
	Reporter                      *jira.User             `json:"reporter,omitempty" structs:"reporter,omitempty"`
	Components                    []*jira.Component      `json:"components,omitempty" structs:"components,omitempty"`
	Status                        *jira.Status           `json:"status,omitempty" structs:"status,omitempty"`
	Progress                      *jira.Progress         `json:"progress,omitempty" structs:"progress,omitempty"`
	AggregateProgress             *jira.Progress         `json:"aggregateprogress,omitempty" structs:"aggregateprogress,omitempty"`
	TimeTracking                  *jira.TimeTracking     `json:"timetracking,omitempty" structs:"timetracking,omitempty"`
	TimeSpent                     int                    `json:"timespent,omitempty" structs:"timespent,omitempty"`
	TimeEstimate                  int                    `json:"timeestimate,omitempty" structs:"timeestimate,omitempty"`
	TimeOriginalEstimate          int                    `json:"timeoriginalestimate,omitempty" structs:"timeoriginalestimate,omitempty"`
	Worklog                       *jira.Worklog          `json:"worklog,omitempty" structs:"worklog,omitempty"`
	IssueLinks                    []*jira.IssueLink      `json:"issuelinks,omitempty" structs:"issuelinks,omitempty"`
	Comments                      *Comments              `json:"comment,omitempty" structs:"comment,omitempty"`
	FixVersions                   []*jira.FixVersion     `json:"fixVersions,omitempty" structs:"fixVersions,omitempty"`
	AffectsVersions               []*jira.AffectsVersion `json:"versions,omitempty" structs:"versions,omitempty"`
	Labels                        []string               `json:"labels,omitempty" structs:"labels,omitempty"`
	Subtasks                      []*jira.Subtasks       `json:"subtasks,omitempty" structs:"subtasks,omitempty"`
	Attachments                   []*jira.Attachment     `json:"attachment,omitempty" structs:"attachment,omitempty"`
	Epic                          *jira.Epic             `json:"epic,omitempty" structs:"epic,omitempty"`
	Sprint                        *jira.Sprint           `json:"sprint,omitempty" structs:"sprint,omitempty"`
	Parent                        *jira.Parent           `json:"parent,omitempty" structs:"parent,omitempty"`
	AggregateTimeOriginalEstimate int                    `json:"aggregatetimeoriginalestimate,omitempty" structs:"aggregatetimeoriginalestimate,omitempty"`
	AggregateTimeSpent            int                    `json:"aggregatetimespent,omitempty" structs:"aggregatetimespent,omitempty"`
	AggregateTimeEstimate         int                    `json:"aggregatetimeestimate,omitempty" structs:"aggregatetimeestimate,omitempty"`
	Unknowns                      tcontainer.MarshalMap
}

type Comments struct {
	Comments   []*Comment `json:"comments,omitempty" structs:"comments,omitempty"`
	Self       string     `json:"self,omitempty" structs:"self,omitempty"`
	MaxResults int        `json:"maxResults,omitempty" structs:"maxResults,omitempty"`
	Total      int        `json:"total,omitempty" structs:"total,omitempty"`
	StartAt    int        `json:"startAt,omitempty" structs:"startAt,omitempty"`
}

// Comment represents a comment by a person to an issue in Jira.
type Comment struct {
	ID           string                 `json:"id,omitempty" structs:"id,omitempty"`
	Self         string                 `json:"self,omitempty" structs:"self,omitempty"`
	Name         string                 `json:"name,omitempty" structs:"name,omitempty"`
	Author       jira.User              `json:"author,omitempty" structs:"author,omitempty"`
	Body         *adf.ADF               `json:"body,omitempty" structs:"body,omitempty"`
	UpdateAuthor jira.User              `json:"updateAuthor,omitempty" structs:"updateAuthor,omitempty"`
	Updated      string                 `json:"updated,omitempty" structs:"updated,omitempty"`
	Created      string                 `json:"created,omitempty" structs:"created,omitempty"`
	Visibility   jira.CommentVisibility `json:"visibility,omitempty" structs:"visibility,omitempty"`

	// A list of comment properties. Optional on create and update.
	Properties []jira.EntityProperty `json:"properties,omitempty" structs:"properties,omitempty"`
}

type service struct {
	client *jira.Client
}

type IssueService service

type searchResult struct {
	Issues     []Issue `json:"issues" structs:"issues"`
	StartAt    int     `json:"startAt" structs:"startAt"`
	MaxResults int     `json:"maxResults" structs:"maxResults"`
	Total      int     `json:"total" structs:"total"`
}

func (s *IssueService) Search(ctx context.Context, jql string, options *jira.SearchOptions) ([]Issue, *jira.Response, error) {
	u := url.URL{
		Path: "rest/api/3/search",
	}
	uv := url.Values{}
	if jql != "" {
		uv.Add("jql", jql)
	}

	if options != nil {
		if options.StartAt != 0 {
			uv.Add("startAt", strconv.Itoa(options.StartAt))
		}
		if options.MaxResults != 0 {
			uv.Add("maxResults", strconv.Itoa(options.MaxResults))
		}
		if options.Expand != "" {
			uv.Add("expand", options.Expand)
		}
		if strings.Join(options.Fields, ",") != "" {
			uv.Add("fields", strings.Join(options.Fields, ","))
		}
		if options.ValidateQuery != "" {
			uv.Add("validateQuery", options.ValidateQuery)
		}
	}

	u.RawQuery = uv.Encode()

	req, err := s.client.NewRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return []Issue{}, nil, err
	}

	v := new(searchResult)
	resp, err := s.client.Do(req, v)
	if err != nil {
		err = jira.NewJiraError((*jira.Response)(resp), err)
	}
	return v.Issues, resp, err
}
