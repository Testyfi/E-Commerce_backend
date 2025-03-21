package utility

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)
func SaveImageToCloudFlare(fileHeader *multipart.FileHeader, filename string, id string, directory string) error {
	
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()
    bucketName    := os.Getenv("BUCKET_NAME")  // Your Cloudflare R2 bucket name
	region        := os.Getenv("REGION")       // Cloudflare uses 'auto' as the region
	r2Endpoint    := os.Getenv("R2_ENDPOINT")
	accessKey     := os.Getenv("ACCESS_KEY")
	secretKey     := os.Getenv("SECRET_KEY")
	
	imageKey      := filename// Name of the file in the bucket
	
	
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
	if err != nil {
		return err
	}
	return nil
}
