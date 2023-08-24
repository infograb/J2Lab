package utils

// import (
// 	"bytes"
// 	"io"
// 	"net/http"

// 	log "github.com/sirupsen/logrus"
// )

// func DownloadFile(url string) (*bytes.Reader, error) {
// 	response, err := http.Get(url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer response.Body.Close()

// 	if response.StatusCode != http.StatusOK {
// 		log.Fatalf("Error downloading file: %s", response.Status)
// 	}

// 	imageData, err := io.ReadAll(response.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return bytes.NewReader(imageData), nil
// }
