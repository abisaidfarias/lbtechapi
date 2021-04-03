package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
)

// IStorageService is the storage service
type IStorageService interface {
	UploadImage([]byte) (string, error)
}

type storageService struct {
}

func NewStorageService() IStorageService {
	return &storageService{}
}

func (s *storageService) UploadImage(filesf []byte) (string, error) {

	accountKey, accountName, _, _ := functions.GetAccountInfo()
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return "", err
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	containerName := utils.ImageContainer
	pathUrl, _ := url.Parse(
		fmt.Sprintf(utils.BaseUrlAzureBlob, accountName, containerName))

	containerURL := azblob.NewContainerURL(*pathUrl, p)

	ctx := context.Background()

	fileName := functions.RandomImageString()
	err = ioutil.WriteFile(fileName, filesf, 0700)
	if err != nil {
		return "", err
	}
	blobURL := containerURL.NewBlockBlobURL(fileName)
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	_, err = azblob.UploadFileToBlockBlob(ctx, file, blobURL, azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16})
	if err != nil {
		return "", err
	}
	file.Close()
	os.Remove(fileName)
	return fmt.Sprintf("%s/%s", pathUrl, fileName), nil
}
