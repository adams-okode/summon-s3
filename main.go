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
	helpFlag := flag.Bool("h", false, "Show help")
	helpLongFlag := flag.Bool("help", false, "Show help")
	versionFlag := flag.Bool("V", false, "Show version")
	flag.Parse()

	// Handle help and version flags
	if *helpFlag || *helpLongFlag {
		fmt.Println("Usage: summon <space-name>/<key-path>")
		fmt.Println("Fetches a secret from an S3-compatible storage service (e.g., DigitalOcean Spaces).")
		fmt.Println("Flags:")
		fmt.Println("  -h, --help  Show this help message")
		fmt.Println("  -V          Show version")
		os.Exit(0)
	}
	if *versionFlag {
		fmt.Println("summon-s3 version 0.1.0") // Replace with actual version
		os.Exit(0)
	}

	// Check for correct number of arguments
	args := flag.Args()
	if len(args) != 1 {
		printAndExit("You must pass in one argument in the format <space-name>/<key-path>")
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