package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
		Region: cfg.Region,
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey,
				""),
		),
	}

	if cfg.Endpoint != "" {
		awsCfg.BaseEndpoint = aws.String(cfg.Endpoint)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	input := &s3.CreateBucketInput{
		Bucket: aws.String(cfg.BucketName),
	}
	if cfg.Region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(cfg.Region),
		}
	}

	_, err := client.CreateBucket(ctx, input)
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) && apiError.ErrorCode() == "BucketAlreadyOwnedByYou" {
			return client, nil
		}
		return nil, fmt.Errorf("failed to create s3 bucket: %w", err)
	}

	return client, nil
}
