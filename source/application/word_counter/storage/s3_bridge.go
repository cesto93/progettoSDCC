package storage

import (
	"fmt"
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
 	"github.com/aws/aws-sdk-go/aws/session"
 	"github.com/aws/aws-sdk-go/service/s3"
 	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)
const (
	awsRegion = "us-east-1"
)

type S3Bridge struct {
	bucketName string
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

type BridgeStorage interface {
	Read(key string) ([]byte, error)
	Write(key string, data []byte) error
}

func New(bucketName string) S3Bridge {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion), }))
 	uploader := s3manager.NewUploader(sess)
 	downloader := s3manager.NewDownloader(sess)
 	out := S3Bridge{bucketName, uploader, downloader}
 	return out
}

func (bridge *S3Bridge) Read(key string) ([]byte, error) {
	buffer := aws.NewWriteAtBuffer([]byte{})
	_, err := bridge.downloader.Download(buffer, &s3.GetObjectInput{
    	Bucket: aws.String(bridge.bucketName),
   		Key:    aws.String(key),
	})
	if err != nil {
    	return nil, fmt.Errorf("failed to download file, %v", err)
	}
	return buffer.Bytes(), nil;
}

func (bridge *S3Bridge) Write(key string, data []byte) error {
	_, err := bridge.uploader.Upload(&s3manager.UploadInput{
    	Bucket: aws.String(bridge.bucketName),
    	Key:    aws.String(key),
    	Body:   bytes.NewReader(data),
	})
	if err != nil {
    	return fmt.Errorf("failed to upload file, %v", err)
	}
	return nil
}