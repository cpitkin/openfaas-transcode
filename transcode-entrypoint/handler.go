package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type minioObject struct {
	Key string `json:"Key"`
}

type fileInfo struct {
	Bucket   string `json:"bucket"`
	File     string `json:"file"`
	FileType string `json:"FileType"`
}

// Handle a serverless request
func Handle(req []byte) string {

	// Get values about the file from the req
	mObj := minioObject{}
	json.Unmarshal(req, &mObj)

	split := strings.Split(mObj.Key, "/")
	file := fileInfo{split[0], split[2], split[1]}

	fileBuf := new(bytes.Buffer)
	json.NewEncoder(fileBuf).Encode(file)

	reqFunc, err := http.NewRequest("POST", "http://192.168.1.102:8081/async-function/transcode-worker", fileBuf)

	reqFunc.Header.Set("Content-Type", "application/json")
	reqFunc.Header.Set("X-Callback-URL", "http://192.168.1.102:8081/async-function/transcode-move")

	client := &http.Client{}
	resp, err := client.Do(reqFunc)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	defer resp.Body.Close()

	return fmt.Sprintf("response headers: %v", resp.Header)
}
