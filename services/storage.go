package services

import (
	"fmt"
	"strings"

	"bytes"

	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gabriel-vasile/mimetype"
)

// IStorageService is the storage service
type IStorageService interface {
	UploadFile([]byte) (string, error)
	// UploadFileWithKey uploads bytes using the given S3 object key (path inside the bucket).
	UploadFileWithKey([]byte, string) (string, error)
}

type storageService struct {
}

func NewStorageService() IStorageService {
	return &storageService{}
}

func (s *storageService) UploadFile(filesf []byte) (string, error) {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		return "", fmt.Errorf("failed to connect to aws s3, %v", err)
	}
	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)
	mtype := mimetype.Detect(filesf)
	reader := bytes.NewReader(filesf)
	fileName := functions.RandomFileName(mtype.Extension())
	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("lbtechimages"),
		Key:    aws.String(fileName),
		Body:   reader,
		ACL:    aws.String(s3.BucketCannedACLPublicRead),
		
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}
	return result.Location, nil

}

func (s *storageService) UploadFileWithKey(filesf []byte, key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return s.UploadFile(filesf)
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		return "", fmt.Errorf("failed to connect to aws s3, %v", err)
	}
	uploader := s3manager.NewUploader(sess)
	reader := bytes.NewReader(filesf)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("lbtechimages"),
		Key:    aws.String(key),
		Body:   reader,
		ACL:    aws.String(s3.BucketCannedACLPublicRead),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}
	return result.Location, nil
}
