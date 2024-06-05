package services

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"mime/multipart"
)

// S3Client is a wrapper for the AWS S3 client
type S3Client struct {
	Client *s3.Client
	Config aws.Config
}

// NewS3Client initializes a new S3 client
func NewS3Client() *S3Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	client := s3.NewFromConfig(cfg)
	log.Println("S3 client connected", client)
	return &S3Client{Client: client, Config: cfg}
}

// ListBuckets lists all buckets in the S3 account
func (s *S3Client) ListBuckets() {
	result, err := s.Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatalf("unable to list buckets, %v", err)
	}

	fmt.Println("Buckets:")
	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n", aws.ToString(b.Name), aws.ToTime(b.CreationDate))
	}
}

// UploadFile uploads a file to a specified S3 bucket and returns the URL
func (s *S3Client) UploadFile(bucket, fileType, key string, file multipart.File) (string, error) {
	// Define the directory based on file type
	var directory string
	switch fileType {
	case "session":
		directory = "audioSessions"
	case "background":
		directory = "audioBackground"
	default:
		return "", fmt.Errorf("invalid file type %q", fileType)
	}

	// Construct the full key with directory
	fullKey := fmt.Sprintf("%s/%s", directory, key)

	uploader := manager.NewUploader(s.Client)
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fullKey),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("unable to upload %q to %q, %v", key, bucket, err)
	}

	region := s.Config.Region
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, fullKey)
	log.Printf("Successfully uploaded %q to %q\n", key, url)
	return url, nil
}
