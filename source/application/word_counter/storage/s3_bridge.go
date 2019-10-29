package storage

import (
	"time"
	"os"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
 	"github.com/aws/aws-sdk-go/aws/awserr"
 	"github.com/aws/aws-sdk-go/aws/request"
 	"github.com/aws/aws-sdk-go/aws/session"
 	"github.com/aws/aws-sdk-go/service/s3"
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
	read(key string) []byte
	write(key string, data []bytes)
}

func New(string bucketName) S3Bridge {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion), }))
 	uploader := s3manager.NewUploader(sess)
 	downloader := s3manager.NewDownloader(sess)
 	out := S3Bridge{ bucketName, uploader, downloader}
 	return out
}

func (bridge *S3Bridge) read(key string) []byte {
	buffer := aws.NewWriteAtBuffer([]byte{})
	_, err := bridge.downloader.Download(buffer, &s3.GetObjectInput{
    	Bucket: aws.String(bridge.bucketName),
   		Key:    aws.String(key),
	})
	if err != nil {
    	return fmt.Errorf("failed to download file, %v", err)
	}
	return buffer.Bytes();
}

func (bridge *S3Bridge) write(key string, data []bytes) {
	result, err := uploader.Upload(&s3manager.UploadInput{
    	Bucket: aws.String(bridge.bucketName),
    	Key:    aws.String(string),
    	Body:   data,
	})
	if err != nil {
    	return fmt.Errorf("failed to upload file, %v", err)
	}
}