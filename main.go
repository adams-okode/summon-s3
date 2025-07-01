package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/aws"
)

func main() {
	// Define flags
	if len(os.Args) != 2 {Add commentMore actions
		printAndExit("You must pass in one argument")
	}
	variableName := args[0]
	bucketName := strings.Split(variableName, "/")[0]
	keyName := strings.Join(strings.Split(variableName, "/")[1:], "/")

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		printAndExit(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize S3 client with explicit endpoint, allowing override from environment
	endpoint := os.Getenv("AWS_ENDPOINT_URL")
	svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	// Make sure bucket exists
	_, err = svc.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: &bucketName,
	})
	if err != nil {
		printAndExit(fmt.Sprintf("HeadBucket failed: %v %v", err, variableName))
	}

	// Fetch object
	getParams := &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &keyName,
	}
	resp, err := svc.GetObject(context.Background(), getParams)
	if err != nil {
		printAndExit(fmt.Sprintf("GetObject failed: %v %v", err, variableName))
	}

	defer resp.Body.Close()
	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		printAndExit(fmt.Sprintf("Failed to read object: %v", err))
	}

	fmt.Print(string(contents))
}

func printAndExit(err string) {
	os.Stderr.Write([]byte(err + "\n"))
	os.Exit(1)
}