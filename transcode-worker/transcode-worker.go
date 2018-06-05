package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	minio "github.com/minio/minio-go"
)

type fileInfo struct {
	Bucket   string `json:"bucket"`
	File     string `json:"file"`
	FileType string `json:"FileType"`
}

var fullPath string

func initializeMinio(endpoint string, accessKeyID string, secretAccessKey string, useSSL bool) (*minio.Client, error) {

	// Initialize Minio client object.
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return nil, err
	}
	return minioClient, nil
}

func main() {

	url, err := ioutil.ReadFile("/run/secrets/minio_url")
	if err != nil {
		log.Fatal("%v", err)
	}
	port, err := ioutil.ReadFile("/run/secrets/minio_port")
	if err != nil {
		log.Fatal("%v", err)
	}
	aKID, err := ioutil.ReadFile("/run/secrets/access_key")
	if err != nil {
		log.Fatal("%v", err)
	}
	sAKey, err := ioutil.ReadFile("/run/secrets/secret_key")
	if err != nil {
		log.Fatal("%v", err)
	}
	useSSL := false

	trim := "\n\r"

	endpoint := strings.Trim(string(url), trim) + ":" + strings.Trim(string(port), trim)
	accessKeyID := strings.Trim(string(aKID), trim)
	secretAccessKey := strings.Trim(string(sAKey), trim)

	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("Unable to read standard input:", err)
	}
	if len(input) == 0 {
		log.Fatalf("A JSON object is required")
	}

	// Get values about the file from the req
	file := fileInfo{}
	json.Unmarshal([]byte(input), &file)

	// Minio client
	minioClient, err := initializeMinio(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalf("%v", err)
	}

	fullPath = file.FileType + "/" + file.File

	// Get the object from Minio
	mFile, err := minioClient.GetObject(file.Bucket, fullPath, minio.GetObjectOptions{})
	if err != nil {
		log.Fatalf("Minio get object: %v", err)
	}
	defer mFile.Close()

	// Create the folder strucutre to hold the files
	dr := os.MkdirAll("/data/raw", 0755)
	dt := os.MkdirAll("/data/transcoded", 0755)
	if dr != nil || dt != nil {
		log.Fatalf("Directory create: %v", err)
	}

	// Create a file in which to store the object
	tmpFile := "/data/raw/" + file.File
	localFile, err := os.Create(tmpFile)
	if err != nil {
		log.Fatalf("File create: %v", err)
	}
	defer localFile.Close()

	// Store the object in a file
	if _, err = io.Copy(localFile, mFile); err != nil {
		log.Fatalf("Copy file: %v", err)
	}

	// Run the transcoder over the file
	cmd := exec.Command("transcode-video", tmpFile, "--output", "/data/transcoded/"+file.File)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Printf("transcode: %v", stdoutStderr)

	// Upload completed file to complete bucket
	minioClient.FPutObject("complete", file.FileType+"/"+file.File, "/data/transcoded/"+file.File, minio.PutObjectOptions{})
	if err != nil {
		log.Fatalf("%v", err)
	}

	// Remove the file from the local filesystem
	os.Remove("/data/transcoded/" + file.File)
	os.Remove("/data/raw/" + file.File)

	// Encode the file struct
	fileBuf := new(bytes.Buffer)
	json.NewEncoder(fileBuf).Encode(file)

	// Setup request to the move function
	reqFunc, err := http.NewRequest("POST", "http://192.168.1.102:8081/function/transcode-move", fileBuf)
	reqFunc.Header.Set("Content-Type", "application/json")

	// Make a call to the next function
	client := &http.Client{}
	resp, err := client.Do(reqFunc)
	if err != nil {
		log.Fatalf("Post function hook: %v", err)
	}
	defer resp.Body.Close()

}
