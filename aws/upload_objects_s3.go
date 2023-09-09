package s3

import (
	"fmt"
	"mime/multipart"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func UploadObject(bucket string, fileName string, sess *session.Session, awsConfig AWSConfig, file multipart.File) error {

	// Open file to upload

	// Upload to s3
	uploader := s3manager.NewUploader(sess)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileName),
		Body:        file,
		ContentType: aws.String("image/png"),
	})

	if err != nil {
		fmt.Printf("failed to upload object, %v\n", err)
		return err
	}

	fmt.Printf("Successfully uploaded %q to %q\n", fileName, bucket)
	return nil
}
