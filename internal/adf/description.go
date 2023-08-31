package adf

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/config"
)

func GetIssueDescriptionADF(issueKey string) (*ADFBlock, error) {
	cfg := config.GetConfig()
	client := &http.Client{}

	url := fmt.Sprintf("%s/rest/api/3/issue/%s?fields=description", cfg.Jira.Host, issueKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	auth := cfg.Jira.Email + ":" + cfg.Jira.Token
	basicToken := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+basicToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	fields, ok := result["fields"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Missing 'fields' in the response")
	}

	descriptionData, ok := fields["description"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Missing 'description' in the response")
	}
	descriptionJSON, err := json.Marshal(descriptionData)
	if err != nil {
		return nil, err
	}
	var adfBlock ADFBlock
	err = json.Unmarshal(descriptionJSON, &adfBlock)
	if err != nil {
		return nil, err
	}
	return &adfBlock, nil
}
