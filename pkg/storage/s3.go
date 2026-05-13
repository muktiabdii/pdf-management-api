package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client is the global AWS S3 client instance.
var S3Client *s3.Client

// BucketName is the name of the S3 bucket used for file storage.
var BucketName string

// Connect initializes the AWS S3 client with credentials from environment variables.
// Environment variables required:
//   - AWS_REGION: AWS region (e.g., ap-southeast-1)
//   - AWS_ACCESS_KEY: AWS access key ID
//   - AWS_SECRET_KEY: AWS secret access key
//   - AWS_BUCKET: S3 bucket name
func Connect() {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY"),
			os.Getenv("AWS_SECRET_KEY"),
			"",
		)),
	)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	S3Client = s3.NewFromConfig(cfg)
	BucketName = os.Getenv("AWS_BUCKET")
	log.Println("S3 connected successfully")
}

// UploadFile uploads a file from a multipart.File to the S3 bucket.
// It returns the S3 file path or an error if the upload fails.
func UploadFile(ctx context.Context, key string, file multipart.File, contentType string) (string, error) {
	_, err := S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(BucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Return S3 file path in standard format
	url := fmt.Sprintf("/uploads/pdf/%s", key)
	return url, nil
}

// DeleteFile removes a file from the S3 bucket by its key.
// Returns an error if the deletion fails.
func DeleteFile(ctx context.Context, key string) error {
	_, err := S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}
	return nil
}

// UploadFileFromReader uploads a file from an io.Reader to the S3 bucket.
// It allows for uploading files generated in-memory (e.g., PDF streams).
// The size parameter is the total file size in bytes.
func UploadFileFromReader(ctx context.Context, key string, reader io.Reader, contentType string, size int64) (string, error) {
	_, err := S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(BucketName),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	url := fmt.Sprintf("/uploads/pdf/%s", key)
	return url, nil
}
