package s3

import (
	"bytes"
	"fmt"
	"mime/multipart"

	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)
func UploadTshirtPictures(){

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

    bucketName := os.Getenv("bucketName")
	region:=os.Getenv("region")
    r2Endpoint:=os.Getenv("r2Endpoint")
	accessKey:=os.Getenv("accessKey")
    secretKey:=os.Getenv("secretKey")
	imagePath:=os.Getenv("imagePath")
	imageKey:=os.Getenv("imageKey")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{URL: r2Endpoint, SigningRegion: region}, nil
		})),
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	// Upload the file
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(imageKey),
		Body:   file,
	})
	if err != nil {
		log.Fatalf("Failed to upload image: %v", err)
	}

	fmt.Println("Image uploaded successfully to Cloudflare R2!")

}
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

func URLImageUpooad(bucket string, fileName string, sess *session.Session, awsConfig AWSConfig, file bytes.Buffer) error {

	uploader := s3manager.NewUploader(sess)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(file.Bytes()),
		ContentType: aws.String("image/png"),
	})

	if err != nil {
		fmt.Printf("failed to upload object, %v\n", err)
		return err
	}

	fmt.Printf("Successfully uploaded %q to %q\n", fileName, bucket)
	return nil
}
