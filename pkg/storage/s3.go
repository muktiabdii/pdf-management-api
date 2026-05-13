package storage

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"io"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var S3Client *s3.Client
var BucketName string

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

	// return path format: /uploads/pdf/<filename>
	url := fmt.Sprintf("/uploads/pdf/%s", key)
	return url, nil
}

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