package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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

	// Minio client
	ohMC, err := initializeMinio(ohEndpoint, ohAaccessKeyID, ohSecretAccessKey, useSSL)
	if err != nil {
		log.Fatalf("OH: %v", err)
	}

	ffMC, err := initializeMinio(ffEndpoint, ffAaccessKeyID, ffSecretAccessKey, useSSL)
	if err != nil {
		log.Fatalf("RI: %v", err)
	}

	fullPath = file.FileType + "/" + file.File

	// Get the object from Minio
	mFile, err := ohMC.GetObject("complete", fullPath, minio.GetObjectOptions{})
	if err != nil {
		log.Fatalf("Minio get object: %v", err)
	}
	defer mFile.Close()

	// Create the folder strucutre to hold the files
	dr := os.MkdirAll("/data/", 0755)
	if dr != nil {
		log.Fatalf("Directory create: %v", err)
	}

	// Create a file in which to store the object
	tmpFile := "/data/" + file.File
	localFile, err := os.Create(tmpFile)
	if err != nil {
		log.Fatalf("File create: %v", err)
	}
	defer localFile.Close()

	// Store the object in a file
	if _, err = io.Copy(localFile, mFile); err != nil {
		log.Fatalf("Copy file: %v", err)
	}

	ffMC.FPutObject("movies", file.File, tmpFile, minio.PutObjectOptions{})
	if err != nil {
		log.Fatalf("PUT: %v", err)
	}

	// Remove the file from the local filesystem
	os.Remove("/data/" + file.File)

	// Encode the file struct
	fileBuf := new(bytes.Buffer)
	json.NewEncoder(fileBuf).Encode(file)

	// Setup request to the move function
	reqFunc, err := http.NewRequest("POST", "http://192.168.1.102:8081/function/transcode-delete", fileBuf)
	reqFunc.Header.Set("Content-Type", "application/json")

	// Make a call to the next function
	client := &http.Client{}
	resp, err := client.Do(reqFunc)
	if err != nil {
		log.Fatalf("Post function hook: %v", err)
	}
	defer resp.Body.Close()
}
