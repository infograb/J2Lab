package j2g

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// TODO: 업로드한 이미지의 ID를 통해 마크다운 문법 작성
func handleMediaGroup(block ADFBlock) {
	imageData := downloadImage(block)
	// TODO: PROJECT ID 얻는 함수
	gitLabProjectID := "" //func
	uploadImageToGitLab(imageData, gitLabProjectID)
}
func downloadImage(block ADFBlock) []byte {
	var imageData []byte
	return imageData
}
func uploadImageToGitLab(imageData []byte, gitLabProjectID string) (string, error) {

	//cfg token을 얻을 수 있음, Project ID는 API를 통해
	client := &http.Client{}
	url := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/uploads", gitLabProjectID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(imageData))
	if err != nil {
		return "", err
	}
	//req.Header.Add("Authorization", "Bearer "+gitLabToken)
	req.Header.Add("Content-Type", "image/png") // or whatever the image type is
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if url, ok := result["url"].(string); ok {
		return url, nil
	}
	return "", fmt.Errorf("Failed to get the URL")
}
