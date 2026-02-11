package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type S3Config struct {
	Region     string
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
}

func NewS3Client(ctx context.Context, cfg S3Config) (*s3.Client, error) {
	awsCfg := aws.Config{
		Region:       cfg.Region,
		BaseEndpoint: aws.String(cfg.Endpoint),
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey,
				""),
		),
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(cfg.BucketName),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) && apiError.ErrorCode() == "BucketAlreadyOwnedByYou" {
			return client, nil
		}
		return nil, fmt.Errorf("failed to create s3 bucket: %w", err)
	}

	return client, nil
}
