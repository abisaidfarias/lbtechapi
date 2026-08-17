package services

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gabriel-vasile/mimetype"
)

// IStorageService is the storage service
type IStorageService interface {
	UploadFile([]byte) (string, error)
	UploadUserFile([]byte, string) (responses.UploadFileResponse, error)
	// UploadFileWithKey uploads bytes using the given S3 object key (path inside the bucket).
	UploadFileWithKey([]byte, string) (string, error)
	// UploadFileWithKeyAndName uploads bytes using the given S3 object key, but sets
	// Content-Disposition so the file downloads under downloadFileName instead of the
	// (URL-safe) key. The key itself is untouched, so it stays safe to embed in a URL.
	UploadFileWithKeyAndName(filesf []byte, key, downloadFileName string) (string, error)
}

type storageService struct {
}

func NewStorageService() IStorageService {
	return &storageService{}
}

func (s *storageService) UploadFile(filesf []byte) (string, error) {
	mtype := mimetype.Detect(filesf)
	ext := ""
	if mtype != nil {
		ext = mtype.Extension()
	}
	fileName := functions.RandomFileName(ext)
	url, err := s.uploadToS3(filesf, fileName, "", "")
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *storageService) UploadUserFile(filesf []byte, originalFileName string) (responses.UploadFileResponse, error) {
	var empty responses.UploadFileResponse
	if err := utils.ValidateUploadContentType(filesf); err != nil {
		return empty, err
	}

	mtype := mimetype.Detect(filesf)
	ext := ""
	if mtype != nil {
		ext = mtype.Extension()
	}

	key, storedFileName, displayOriginal, err := utils.BuildUniqueUploadKey(originalFileName, ext)
	if err != nil {
		return empty, err
	}

	url, err := s.uploadToS3(filesf, key, displayOriginal, mtypeString(mtype))
	if err != nil {
		return empty, err
	}

	return responses.UploadFileResponse{
		URL:              url,
		OriginalFileName: displayOriginal,
		StoredFileName:   storedFileName,
	}, nil
}

func (s *storageService) UploadFileWithKey(filesf []byte, key string) (string, error) {
	return s.UploadFileWithKeyAndName(filesf, key, "")
}

func (s *storageService) UploadFileWithKeyAndName(filesf []byte, key, downloadFileName string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return s.UploadFile(filesf)
	}
	return s.uploadToS3(filesf, key, strings.TrimSpace(downloadFileName), "")
}

func (s *storageService) uploadToS3(filesf []byte, key, downloadFileName, contentType string) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		return "", fmt.Errorf("failed to connect to aws s3, %v", err)
	}

	uploader := s3manager.NewUploader(sess)
	reader := bytes.NewReader(filesf)

	input := &s3manager.UploadInput{
		Bucket: aws.String("lbtechimages"),
		Key:    aws.String(key),
		Body:   reader,
		ACL:    aws.String(s3.BucketCannedACLPublicRead),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	if downloadFileName != "" {
		input.ContentDisposition = aws.String(fmt.Sprintf(`attachment; filename="%s"`, escapeContentDispositionFilename(downloadFileName)))
	}

	result, err := uploader.Upload(input)
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}
	return result.Location, nil
}

func mtypeString(mtype *mimetype.MIME) string {
	if mtype == nil {
		return ""
	}
	return mtype.String()
}

func escapeContentDispositionFilename(name string) string {
	return strings.NewReplacer(`\`, `_`, `"`, `_`).Replace(name)
}
