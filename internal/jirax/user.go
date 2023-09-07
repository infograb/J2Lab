package jirax

import (
	"context"
	"fmt"
	"net/url"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/pkg/errors"
)

type UserQueryOptions struct {
	AccountId string
	Username  string
	Key       string

	//* Use [expand](#expansion) to include additional information about users in the response. This parameter accepts a comma-separated list.
	// Expand options include:
	// - `groups` includes all groups and nested groups to which the user belongs.
	// - `applicationRoles` includes details of all the applications to which the user has access.
	Expand string
}

func GetUser(jr *jira.Client, options *UserQueryOptions) (*jira.User, *jira.Response, error) {
	u, err := url.Parse(jr.BaseURL.Host)
	if err != nil {
		return nil, nil, errors.Wrap(err, fmt.Sprintf("Error parsing Jira URL: %s", jr.BaseURL))
	}

	u.Path = "/rest/api/2/user"

	q := u.Query()
	if options.AccountId != "" {
		q.Set("accountId", options.AccountId)
	}
	if options.Username != "" {
		q.Set("username", options.Username)
	}
	if options.Key != "" {
		q.Set("key", options.Key)
	}
	if options.Expand != "" {
		q.Set("expand", options.Expand)
	}
	u.RawQuery = q.Encode()

	user := new(jira.User)
	req, err := jr.NewRequest(context.Background(), "GET", u.String(), nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error creating request")
	}

	resp, err := jr.Do(req, user)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error getting user")
	}

	return user, resp, nil
}
