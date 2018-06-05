package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
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
		log.Fatalf("%v", err)
	}
	port, err := ioutil.ReadFile("/run/secrets/minio_port")
	if err != nil {
		log.Fatalf("%v", err)
	}
	aKID, err := ioutil.ReadFile("/run/secrets/access_key")
	if err != nil {
		log.Fatalf("%v", err)
	}
	sAKey, err := ioutil.ReadFile("/run/secrets/secret_key")
	if err != nil {
		log.Fatalf("%v", err)
	}

	ffURL, err := ioutil.ReadFile("/run/secrets/ff_minio_url")
	if err != nil {
		log.Fatalf("%v", err)
	}
	ffAKID, err := ioutil.ReadFile("/run/secrets/ff_access_key")
	if err != nil {
		log.Fatalf("%v", err)
	}
	ffSKey, err := ioutil.ReadFile("/run/secrets/ff_secret_key")
	if err != nil {
		log.Fatalf("%v", err)
	}
	useSSL := false

	trim := "\n\r"

	ohEndpoint := strings.Trim(string(url), trim) + ":" + strings.Trim(string(port), trim)
	ohAaccessKeyID := strings.Trim(string(aKID), trim)
	ohSecretAccessKey := strings.Trim(string(sAKey), trim)

	ffEndpoint := strings.Trim(string(ffURL), trim) + ":" + strings.Trim(string(port), trim)
	ffAaccessKeyID := strings.Trim(string(ffAKID), trim)
	ffSecretAccessKey := strings.Trim(string(ffSKey), trim)

	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Unable to read standard input: %v", err)
	}
	if len(input) == 0 {
		log.Fatalf("A JSON object is required")
	}

	// Get values about the file from the req
	file := fileInfo{}
	json.Unmarshal([]byte(input), &file)

	fullPath = file.FileType + "/" + file.File

	// Minio client
	ohMC, err := initializeMinio(ohEndpoint, ohAaccessKeyID, ohSecretAccessKey, useSSL)
	if err != nil {
		log.Fatalf("OH: %v", err)
	}

	ffMC, err := initializeMinio(ffEndpoint, ffAaccessKeyID, ffSecretAccessKey, useSSL)
	if err != nil {
		log.Fatalf("RI: %v", err)
	}

	objInfo, err := ffMC.StatObject("movies", file.File, minio.StatObjectOptions{})
	if err != nil {
		log.Fatalf("RI stat: %v", err)
	}
	if objInfo.Size == 0 {
		log.Fatalf("RI stat no object: %v", err)
	}

	err = ohMC.RemoveObject(file.Bucket, fullPath)
	if err != nil {
		log.Fatalf("OH remove transcode: %v", err)
	}

	err = ohMC.RemoveObject("complete", fullPath)
	if err != nil {
		log.Fatalf("OH remove complete: %v", err)
	}
}
