package jirax

import (
	"context"
	"log"
	"regexp"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
)

func parsePlainToMediaName(plain string) []string {
	re, err := regexp.Compile(`!([^|]+)\|width=(\d+),height=(\d+)!`)
	if err != nil {
		log.Fatalf("Error compiling regexp: %s", err)
	}

	// Make a list of all matches
	matches := re.FindAllStringSubmatch(plain, -1)

	mediaNames := make([]string, len(matches))
	for i, match := range matches {
		mediaNames[i] = match[1]
	}

	return mediaNames
}

func UnpaginateIssue(
	jr *jira.Client,
	jql string,
) ([]*Issue, error) {
	issueService := IssueService{client: jr}

	var result []*Issue

	searchOptions := &jira.SearchOptions{
		StartAt:    0,
		MaxResults: 100,
		Fields:     []string{"*all"},
	}

	for {
		itemsV2, _, err := jr.Issue.Search(context.Background(), jql, searchOptions)
		if err != nil {
			return nil, err
		}

		itemsV3, r, err := issueService.Search(context.Background(), jql, searchOptions)
		if err != nil {
			return nil, err
		}

		//* Mapping Media
		for i, itemV3 := range itemsV3 {
			itemV2 := itemsV2[i]

			attachments := itemV3.Fields.Attachments

			// Mapping Description with Attachment
			descriptionMediaNames := parsePlainToMediaName(itemV2.Fields.Description)

			descriptionMedia := make([]string, len(descriptionMediaNames))
			descriptionMediaCount := 0
			for _, mediaName := range descriptionMediaNames {
				for _, attachment := range attachments {
					if attachment.Filename == mediaName {
						descriptionMedia[descriptionMediaCount] = attachment.ID
						descriptionMediaCount++
						break
					}
				}
			}

			itemV3.Fields.DescriptionMedia = descriptionMedia

			// Mapping Comment with Attachment
			for idx, comment := range itemV2.Fields.Comments.Comments {
				commentMediaNames := parsePlainToMediaName(comment.Body)

				commentMedia := make([]string, len(commentMediaNames))
				commentMediaCount := 0
				for _, mediaName := range commentMediaNames {
					for _, attachment := range attachments {
						if attachment.Filename == mediaName {
							commentMedia[commentMediaCount] = attachment.ID
							commentMediaCount++
							break
						}
					}
				}

				itemV3.Fields.Comments.Comments[idx].BodyMedia = commentMedia
			}

			result = append(result, &itemV3)
		}

		if err != nil {
			return nil, err
		}

		if r.StartAt+r.MaxResults >= r.Total {
			break
		}

		searchOptions.StartAt += len(itemsV3)
	}

	return result, nil
}
