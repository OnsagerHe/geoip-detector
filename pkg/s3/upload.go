package s3

import (
	"bytes"
	"context"
	"os"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (s *Bucket) s3Path(p string) string {
	return path.Join(s.RootDirectory, p)
}

func (s *Bucket) UploadFile(ctx context.Context, filePath string, key string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = s.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.Config.Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String("application/octet-stream"),
		ACL:         types.ObjectCannedACLPublicRead,
	})

	if err != nil {
		return err
	}
	return nil
}

func (s *Bucket) UploadBytes(ctx context.Context, key string, value []byte) error {
	_, err := s.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.Config.Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(value),
		ACL:    types.ObjectCannedACLPublicRead,
	})

	if err != nil {
		return err
	}

	return nil
}
