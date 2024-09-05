package ceph

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

var s3Client *s3.Client

// init function
func init() {
	accessKey := "L4ZATJRF4G1YX559NBDE"
	secretKey := "OFSX0li3BmMXAP4ZyBURF5Dga58iLPzBnZCsUOXC"
	endpoint := "http://127.0.0.1:7480"
	_, err := NewS3Client(accessKey, secretKey, endpoint)

	if err != nil {
		fmt.Printf("Failed to initialize s3 client for ceph: %v", err)
	}
	fmt.Println("Ceph Connection OK")
}

func NewS3Client(accessKey, secretKey, endpoint string) (*s3.Client, error) {
	if s3Client != nil {
		return s3Client, nil
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("westcoast"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		fmt.Println("Couldn't load default config. Please check credentials")
		fmt.Println(err)
		return nil, err
	}
	s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
	_, err = s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		fmt.Println("Couldn't list buckets. Please check connection.")
		fmt.Println(err)
		return nil, err
	}
	return s3Client, nil
}

// BucketExists: check bucket
func BucketExists(bucketName string) (bool, error) {
	_, err := s3Client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	exists := true
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				exists = false
				err = nil
			default:
				fmt.Println("An error occur: %v", err)
			}
		}
	} else {
		fmt.Printf("Bucket %v exists and you own it!", bucketName)
	}
	return exists, err
}

// CreateBucket
func CreateBucket(name string) error {
	_, err := s3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		fmt.Printf("Couldn't create a bucket, name %v, error: %v", name, err)
	}
	return err
}

// UploadFileSync: upload file and puts the data into an object in a bucket
// objectKey := path
func UploadFileSync(bucketName string, objectKey string, fileData *os.File) error {
	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   fileData,
	})
	if err != nil {
		fmt.Printf("Error uploading file: %v", err)
		return err
	}
	return nil
}

// DownloadFileSync: Download file and return a *os.File, err
func DownloadFileSync(bucketName string, objectKey string, fileName string) (*os.File, error) {
	result, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		fmt.Printf("Error downloading file from ceph: %v", err)
		return nil, err
	}
	defer result.Body.Close()
	file, err := os.Create("/tmp/" + fileName)
	if err != nil {
		fmt.Printf("Error Creating file from ceph: %v", err)
		return nil, err
	}
	defer file.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		fmt.Printf("Error Reading from object body file from ceph: %v", err)
		return nil, err
	}
	_, err = file.Write(body)
	if err != nil {
		fmt.Printf("Error Writing content to local filefrom ceph: %v", err)
		return nil, err
	}
	return file, err
}
