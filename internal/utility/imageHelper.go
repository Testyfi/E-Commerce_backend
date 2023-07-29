package utility

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func SaveImageToFile(fileHeader *multipart.FileHeader, filename string) error {
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	// Create the "assets" directory if it doesn't exist
	err = os.MkdirAll("./assets", os.ModePerm)
	if err != nil {
		return err
	}

	imagePath := path.Join("./assets", filename)
	out, err := os.Create(imagePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	return err
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
