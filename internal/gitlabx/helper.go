package gitlabx

import (
	"fmt"
	"strconv"

	"github.com/xanzy/go-gitlab"
)

func parseID(id interface{}) (string, error) {
	switch v := id.(type) {
	case int:
		return strconv.Itoa(v), nil
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("invalid ID type %#v, the ID must be an int or a string", id)
	}
}

func Unpaginate[T any](
	gl *gitlab.Client,
	function func(opt *gitlab.ListOptions) ([]*T, *gitlab.Response, error),
) ([]*T, error) {
	var items []*T
	page := 1
	perPage := 100

	for {
		opt := &gitlab.ListOptions{
			Page:    page,
			PerPage: perPage,
		}

		ts, resp, err := function(opt)
		if err != nil {
			return nil, err
		}

		items = append(items, ts...)

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		page = resp.NextPage
	}

	return items, nil
}
