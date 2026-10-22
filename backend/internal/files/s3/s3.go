package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"

	"github.com/taheralfayad/bean_game/backend/internal/config"
	"github.com/taheralfayad/bean_game/backend/internal/files"
)

var _ files.Storage = (*Storage)(nil)

type Storage struct {
	client    *awss3.Client
	presigner *awss3.PresignClient
	bucket    string
}

func New(cfg config.FilesConfig) *Storage {
	awsCfg := aws.Config{Region: cfg.Region}
	if cfg.AccessKey != "" {
		awsCfg.Credentials = credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")
	}

	client := awss3.NewFromConfig(awsCfg, func(o *awss3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.PathStyle
	})

	return &Storage{
		client:    client,
		presigner: awss3.NewPresignClient(client),
		bucket:    cfg.Bucket,
	}
}

func (s *Storage) EnsureBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &awss3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err == nil {
		return nil
	}
	if !isNotFound(err) {
		return fmt.Errorf("head bucket %q: %w", s.bucket, err)
	}

	_, err = s.client.CreateBucket(ctx, &awss3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil && !isAlreadyOwnedByYou(err) {
		return fmt.Errorf("create bucket %q: %w", s.bucket, err)
	}

	return nil
}

func (s *Storage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          r,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put object %q: %w", key, err)
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object %q: %w", key, err)
	}
	return out.Body, nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}
	return nil
}

func (s *Storage) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, err := s.presigner.PresignGetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, awss3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("presign object %q: %w", key, err)
	}
	return req.URL, nil
}

func isNotFound(err error) bool {
	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return true
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NotFound" {
		return true
	}
	var respErr *smithyhttp.ResponseError
	return errors.As(err, &respErr) && respErr.HTTPStatusCode() == http.StatusNotFound
}

func isAlreadyOwnedByYou(err error) bool {
	var owned *types.BucketAlreadyOwnedByYou
	if errors.As(err, &owned) {
		return true
	}
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && apiErr.ErrorCode() == "BucketAlreadyOwnedByYou"
}
