package utility

import (
	"fmt"
	"mime/multipart"
	"os"
	s3 "testify/aws"

	"github.com/aws/aws-sdk-go/aws"
	aws_s3 "github.com/aws/aws-sdk-go/service/s3"
)

var awsConfig = s3.AWSConfig{
	AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
	Region:          os.Getenv("AWS_REGION"),
	AccessKeySecret: os.Getenv("AWS_SECRET_ACCESS_KEY"),
}

func SaveImageToFile(fileHeader *multipart.FileHeader, filename string, id string, directory string) error {
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()
	filename = fmt.Sprintf("%s/%s/%s/%s", "assets", directory, id, filename)

	err = s3.UploadObject("testify-jee", filename, s3.CreateSession(awsConfig), awsConfig, file)

	if err != nil {
		return err
	}
	return nil
}

// Function to delete all question images for a specific qid from the "assets" directory.
func DeleteQuestionImagesByQID(qid string) error {

	svc := s3.CreateS3Session(s3.CreateSession(awsConfig))

	path := fmt.Sprintf("assets/questions/%s", qid)

	// Prepare the delete object input
	input := &aws_s3.DeleteObjectInput{
		Bucket: aws.String("testify-jee"),
		Key:    aws.String(path),
	}

	// Delete the object
	_, err := svc.DeleteObject(input)
	if err != nil {
		return err
	}
	return nil
}
