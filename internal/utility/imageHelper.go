package utility

import (
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	s3 "testify/aws"
)

func SaveImageToFile(fileHeader *multipart.FileHeader, filename string) error {
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	err = s3.UploadObject("testify-jee", filename, s3.CreateSession(s3.AWSConfig{
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		Region:          os.Getenv("AWS_REGION"),
		AccessKeySecret: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}), s3.AWSConfig{
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		Region:          os.Getenv("AWS_REGION"),
		AccessKeySecret: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}, file)

	if err != nil {
		return err
	}
	return nil
}

// Function to delete all question images for a specific qid from the "assets" directory.
func DeleteQuestionImagesByQID(qid string) {
	// Get a list of all files in the "assets" directory
	files, err := ioutil.ReadDir("assets")
	if err != nil {
		fmt.Println("Failed to read the assets directory:", err)
		return
	}

	// Iterate through the files and delete question images with the given qid prefix
	for _, file := range files {
		filename := file.Name()
		if strings.HasPrefix(filename, qid) {
			imagePath := filepath.Join("assets", filename)
			err := os.Remove(imagePath)
			if err != nil {
				fmt.Println("Failed to delete question image:", err)
			}
		}
	}
}
