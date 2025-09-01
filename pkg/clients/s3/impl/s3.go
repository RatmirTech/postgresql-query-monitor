package impl

import (
	"bytes"

	s3_internal "github.com/dreadew/go-common/pkg/clients/s3"
	s3_config "github.com/dreadew/go-common/pkg/config/s3"
	"github.com/dreadew/go-common/pkg/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.uber.org/zap"
)

type s3Client struct {
	client *s3.S3
}

func New(config *s3_config.S3Config) (s3_internal.S3Client, error) {
	logger := logger.GetLogger()

	session, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(
			config.AccessKey,
			config.SecretKey,
			"",
		),
	})
	if err != nil {
		logger.Error("error while creating s3 client", zap.String("error", err.Error()))
		return nil, err
	}

	client := s3.New(session)

	return &s3Client{
		client: client,
	}, nil
}

func (s *s3Client) CreateBucketIfNotExists(bucket string) error {
	_, err := s.client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchBucket {
			_, err = s.client.CreateBucket(&s3.CreateBucketInput{
				Bucket: aws.String(bucket),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *s3Client) PutObject(bucket, key string, stream []byte) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(stream),
	}

	_, err := s.client.PutObject(input)
	if err != nil {
		return err
	}

	return nil
}

func (s *s3Client) GetObject(bucket, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	res, err := s.client.GetObject(input)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *s3Client) DeleteObject(bucket, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(input)
	if err != nil {
		return err
	}

	return nil
}
