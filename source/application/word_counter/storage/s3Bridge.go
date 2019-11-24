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
	client     *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

func New(bucketName string) *S3Bridge {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion), }))
	client := s3.New(sess)
 	uploader := s3manager.NewUploader(sess)
 	downloader := s3manager.NewDownloader(sess)
 	out := S3Bridge{bucketName, client, uploader, downloader}
 	return &out
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

func (bridge *S3Bridge) Delete(keys []string) error {
	objs := make([]*s3.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objs[i] = &(s3.ObjectIdentifier{Key: aws.String(key)})
	} 
	input := &s3.DeleteObjectsInput{
    	Bucket: aws.String(bridge.bucketName),
    	Delete: &s3.Delete{
        	Objects: objs,
        	Quiet: aws.Bool(false),
    	},
	}
	_, err := bridge.client.DeleteObjects(input)
	if err != nil {
        	return fmt.Errorf("failed to delete files, %v", err)
	}
	return nil
}

func (bridge *S3Bridge) List() ([]string, error) {
	input := &s3.ListObjectsInput{
    	Bucket:  aws.String(bridge.bucketName),
	}

	result, err := bridge.client.ListObjects(input)
	if err != nil {
    	return nil, fmt.Errorf("failed to list files, %v", err)
	}
	res := make([]string, len(result.Contents))
	for i, content := range result.Contents {
		res[i] = *(content.Key)
	}
	return res, nil
}