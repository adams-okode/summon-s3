package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	// Check for exactly one argument
	if len(os.Args) != 2 {
		printAndExit("You must pass in one argument")
	}
	variableName := os.Args[1]
	bucketName := strings.Split(variableName, "/")[0]
	keyName := strings.Join(strings.Split(variableName, "/")[1:], "/")

	// Load AWS configuration with custom endpoint if provided
	endpoint := os.Getenv("AWS_ENDPOINT_URL")
	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = "us-east-1"
	}
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithEndpointResolver(
			aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
				if endpoint != "" && service == s3.ServiceID {
					return aws.Endpoint{
						URL:           endpoint,
						SigningRegion: region,
					}, nil
				}
				// Fallback to default
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			}),
		),
	)
	if err != nil {
		printAndExit(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize S3 client
	svc := s3.NewFromConfig(cfg)

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