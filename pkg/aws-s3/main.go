package awss3

import (
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type AwsS3 interface {
	DownloadFile(filePath, fileKey string) error
	UploadFile(filePath, fileKey string) error
	IsFileExists(fileKey string) error
}

func (a awsS3) DeleteFile(fileKey string) error {
	fmt.Printf("\n[#] deleting file %s\n", fileKey)
	defer fmt.Printf("\n[#] deleted file %s\n", fileKey)

	// Delete the file from the bucket
	_, err := a.svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(a.bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return err
	}

	return nil
}

func (a awsS3) createBucket() error {
	fmt.Printf("\n[#] creating bucket %s\n", a.bucketName)
	defer fmt.Printf("\n[#] created bucket %s\n", a.bucketName)

	_, err := a.svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(a.bucketName),
	})
	if err != nil {
		return err
	}

	return nil
}

func (a awsS3) IsFileExists(fileKey string) error {
	_, err := a.svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(a.bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		// If the file does not exist or there is an error, handle the error
		// fmt.Println(err)
		return err
	}
	return nil
}

func (a awsS3) checkS3Created() error {
	_, err := a.svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(a.bucketName),
	})
	if err != nil {
		return err
	}
	return nil
}

func (a awsS3) createBucketIfNotCreated() error {
	if err := a.checkS3Created(); err != nil {
		return a.createBucket()
	}
	return nil
}

func getS3Client(accessKey, accessSecret string) (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-west-1"), // Replace with your desired region
		Credentials: credentials.NewStaticCredentials(accessKey, accessSecret, ""),
	})
	if err != nil {
		return nil, err
	}

	return s3.New(sess), nil
}

type awsS3 struct {
	svc        *s3.S3
	bucketName string
}

// downloadFile implements AwsS3.
func (a awsS3) DownloadFile(filePath, fileKey string) error {
	fmt.Printf("\n[#] downloading file %s\n", fileKey)
	defer fmt.Printf("\n[#] downloaded file %s\n", fileKey)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Download the file from S3
	resp, err := a.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(a.bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return err
	}

	// Write the downloaded file data to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// uploadFile implements AwsS3.
func (a awsS3) UploadFile(filePath, fileKey string) error {
	fmt.Printf("\n[#] uploading file %s\n", fileKey)
	defer fmt.Printf("\n[#] uploaded file %s\n", fileKey)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = a.svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(a.bucketName),
		Key:    aws.String(fileKey),
		Body:   file,
	})
	if err != nil {
		return err
	}

	return nil
}

func NewAwsS3Client(accessKey, accessSec, bucketName string) (AwsS3, error) {
	svc, err := getS3Client(accessKey, accessSec)
	if err != nil {
		return nil, err
	}

	a := awsS3{
		svc:        svc,
		bucketName: bucketName,
	}
	if err := a.createBucketIfNotCreated(); err != nil {
		return nil, err
	}

	return a, nil
}
