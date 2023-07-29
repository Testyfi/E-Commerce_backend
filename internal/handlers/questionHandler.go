package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	database "testify/database"
	models "testify/internal/models"
	utility "testify/internal/utility"
	"time"

	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var questionCollection *mongo.Collection = database.OpenCollection(database.Client, "question")

func GetQuestions(w http.ResponseWriter, r *http.Request) {
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil || pageSize <= 0 {
		pageSize = 10 // default page size
	}

	pageIndex, err := strconv.Atoi(r.URL.Query().Get("pageIndex"))
	if err != nil || pageIndex < 0 {
		pageIndex = 0 // default page index
	}

	// Calculate the number of documents to skip
	skip := pageIndex * pageSize

	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(pageSize))

	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	cur, err := questionCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	questions := []models.Question{}
	for cur.Next(ctx) {
		var question models.Question
		err := cur.Decode(&question)
		if err != nil {
			http.Error(w, "Interal Server Error", http.StatusInternalServerError)
			return
		}
		questions = append(questions, question)
	}
	if err := cur.Err(); err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}

	totalQuestions, err := questionCollection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}

	// Respond with the questions and total number of questions
	response := map[string]interface{}{
		"questions":      questions,
		"totalQuestions": totalQuestions,
	}
	// Set response headers and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreateQuestion(w http.ResponseWriter, r *http.Request) {
	var question models.Question
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	data := r.FormValue("data")
	err = json.Unmarshal([]byte(data), &question)
	if err != nil {
		http.Error(w, "Failed to parse JSON data", http.StatusBadRequest)
		return
	}

	// Checking if question already exists
	alreadyExists, err := questionCollection.CountDocuments(ctx, bson.M{"question": question.Question})
	defer cancel()
	if err != nil {
		http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
		return
	}
	if alreadyExists > 0 {
		http.Error(w, "Question already exists!", http.StatusConflict)
		return
	}

	question.ID = primitive.NewObjectID()
	question.Q_id = question.ID.Hex()
	question.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	questionImages := r.MultipartForm.File["questionImages"]
	question.Images = []string{}
	for i, fileHeader := range questionImages {
		// Save the uploaded file to the "assets" directory
		imageName := fmt.Sprintf("%s%d", question.Q_id, i)
		err := utility.SaveImageToFile(fileHeader, imageName)
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}

		// Update the Images array with the path to the saved image
		question.Images = append(question.Images, imageName)
	}

	if question.Type == "Multiple Choice" {
		optionAImage := r.MultipartForm.File["optionAImage"]
		optionBImage := r.MultipartForm.File["optionBImage"]
		optionCImage := r.MultipartForm.File["optionCImage"]
		optionDImage := r.MultipartForm.File["optionDImage"]

		// Process option A image
		if len(optionAImage) > 0 {
			fileHeader := optionAImage[0]
			imagePath := fmt.Sprintf("%sA", question.ID.Hex())
			err := utility.SaveImageToFile(fileHeader, imagePath)
			if err != nil {
				fmt.Println(err)
				http.Error(w, "Failed to save option A image", http.StatusInternalServerError)
				return
			}
			// Update the Option A Image field with the path to the saved image
			question.Options[0].Image = imagePath
		}

		// Process option B image
		if len(optionBImage) > 0 {
			fileHeader := optionBImage[0]
			imagePath := fmt.Sprintf("%sB", question.ID.Hex())
			err := utility.SaveImageToFile(fileHeader, imagePath)
			if err != nil {
				http.Error(w, "Failed to save option B image", http.StatusInternalServerError)
				return
			}
			// Update the Option B Image field with the path to the saved image
			question.Options[1].Image = imagePath
		}

		// Process option C image
		if len(optionCImage) > 0 {
			fileHeader := optionCImage[0]
			imagePath := fmt.Sprintf("%sC", question.ID.Hex())
			err := utility.SaveImageToFile(fileHeader, imagePath)
			if err != nil {
				http.Error(w, "Failed to save option C image", http.StatusInternalServerError)
				return
			}
			// Update the Option C Image field with the path to the saved image
			question.Options[2].Image = imagePath
		}

		// Process option D image
		if len(optionDImage) > 0 {
			fileHeader := optionDImage[0]
			imagePath := fmt.Sprintf("%sD", question.ID.Hex())
			err := utility.SaveImageToFile(fileHeader, imagePath)
			if err != nil {
				http.Error(w, "Failed to save option D image", http.StatusInternalServerError)
				return
			}
			// Update the Option D Image field with the path to the saved image
			question.Options[3].Image = imagePath
		}
	}

	// Create the question in the database
	insertResult, err := questionCollection.InsertOne(ctx, question)
	if err != nil {
		http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer cancel()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insertResult)
	w.WriteHeader(http.StatusOK)
	return
}

func GetQuestionByID(w http.ResponseWriter, r *http.Request) {
	questionID := chi.URLParam(r, "id")
	question := models.Question{}
	err := questionCollection.FindOne(
		ctx,
		bson.M{"q_id": questionID},
	).Decode(&question)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Question not found"+err.Error())
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error retrieving question")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(question)

}

func DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	questionID := chi.URLParam(r, "id")

	result, err := questionCollection.DeleteOne(ctx, bson.M{"q_id": questionID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error deleting question")
		return
	}

	if result.DeletedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Question not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

}

func EditQuestion(w http.ResponseWriter, r *http.Request) {

	questionID := chi.URLParam(r, "id")

	var updatedQuestion models.Question
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	data := r.FormValue("data")

	err = json.Unmarshal([]byte(data), &updatedQuestion)
	if err != nil {
		fmt.Println("Found error")
		http.Error(w, "Failed to parse JSON data", http.StatusBadRequest)
		return
	}
	utility.DeleteQuestionImagesByQID(questionID)
	questionImages := r.MultipartForm.File["questionImages"]
	updatedQuestion.Images = []string{}
	for i, fileHeader := range questionImages {
		// Save the uploaded file to the "assets" directory
		imageName := fmt.Sprintf("%s%d", questionID, i)
		err := utility.SaveImageToFile(fileHeader, imageName)
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}

		// Update the Images array with the path to the saved image
		updatedQuestion.Images = append(updatedQuestion.Images, imageName)
	}

	if updatedQuestion.Type == "Multiple Choice" {
		optionAImage := r.MultipartForm.File["optionAImage"]
		optionBImage := r.MultipartForm.File["optionBImage"]
		optionCImage := r.MultipartForm.File["optionCImage"]
		optionDImage := r.MultipartForm.File["optionDImage"]

		// Process option A image
		if len(optionAImage) > 0 {
			fileHeader := optionAImage[0]
			imagePath := fmt.Sprintf("%sA", questionID)
			err := utility.SaveImageToFile(fileHeader, imagePath)
			if err != nil {
				http.Error(w, "Failed to save option A image", http.StatusInternalServerError)
				return
			}
			// Update the Option A Image field with the path to the saved image
			updatedQuestion.Options[0].Image = imagePath
		}

		// Process option B image
		if len(optionBImage) > 0 {
			fileHeader := optionBImage[0]
			imagePath := fmt.Sprintf("%sB", questionID)
			err := utility.SaveImageToFile(fileHeader, imagePath)
			if err != nil {
				http.Error(w, "Failed to save option B image", http.StatusInternalServerError)
				return
			}
			// Update the Option B Image field with the path to the saved image
			updatedQuestion.Options[1].Image = imagePath
		}

		// Process option C image
		if len(optionCImage) > 0 {
			fileHeader := optionCImage[0]
			imagePath := fmt.Sprintf("%sC", questionID)
			err := utility.SaveImageToFile(fileHeader, imagePath)
			if err != nil {
				http.Error(w, "Failed to save option C image", http.StatusInternalServerError)
				return
			}
			// Update the Option C Image field with the path to the saved image
			updatedQuestion.Options[2].Image = imagePath
		}

		// Process option D image
		if len(optionDImage) > 0 {
			fileHeader := optionDImage[0]
			imagePath := fmt.Sprintf("%sD", questionID)
			err := utility.SaveImageToFile(fileHeader, imagePath)
			if err != nil {
				http.Error(w, "Failed to save option D image", http.StatusInternalServerError)
				return
			}
			// Update the Option D Image field with the path to the saved image
			updatedQuestion.Options[3].Image = imagePath
		}
	}

	filter := bson.M{"q_id": questionID}
	update := bson.M{"$set": bson.M{
		"question":      updatedQuestion.Question,
		"images":        updatedQuestion.Images,
		"type":          updatedQuestion.Type,
		"options":       updatedQuestion.Options,
		"correctanswer": updatedQuestion.CorrectAnswer,
		"subject_tags":  updatedQuestion.Subject_Tags,
	}}

	result, err := questionCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error updating question")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func DeleteMany(w http.ResponseWriter, r *http.Request) {
	var ids []string
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		fmt.Fprintf(w, "Invalid request body")
		return
	}

	// Iterate through the ids and delete the corresponding question images
	for _, qid := range ids {
		utility.DeleteQuestionImagesByQID(qid)
	}

	result, err := questionCollection.DeleteMany(ctx, bson.M{"q_id": bson.M{"$in": ids}})
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error deleting questions")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func UploadCSV(w http.ResponseWriter, r *http.Request) {
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	qFile, _, err := r.FormFile("questionCsvFile")
	if err != nil {
		http.Error(w, "Failed to retrieve the file", http.StatusBadRequest)
		return
	}
	defer qFile.Close()

	// Parse the CSV file
	reader := csv.NewReader(qFile)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "Failed to parse the CSV file", http.StatusBadRequest)
		return
	}

	mid := make(map[string]string)

	// Insert CSV records in the database
	for _, record := range records {

		question := models.Question{
			Question:      record[0],
			Images:        strings.Split(record[1], ", "),
			Type:          record[2],
			Subject_Tags:  strings.Split(record[3], ", "),
			Q_id:          record[4],
			ID:            primitive.NewObjectID(),
			CorrectAnswer: record[5],
			Created_at:    time.Now(),
			Options:       make([]models.Option, 0),
		}
		qid := question.ID.Hex()
		mid[record[4]] = qid
		question.Q_id = qid
		// Checking if question already exists
		alreadyExists, err := questionCollection.CountDocuments(ctx, bson.M{"question": question.Question})
		defer cancel()
		if err != nil {
			http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
			return
		}
		if alreadyExists > 0 {
			http.Error(w, "Question already exists!", http.StatusConflict)
			return
		}
		// Inserting Question
		_, err = questionCollection.InsertOne(ctx, question)
		if err != nil {
			http.Error(w, "Failed to insert record into the database", http.StatusInternalServerError)
			return
		}
	}
	optionFile, _, err := r.FormFile("optionCsvFile")
	if err != nil {
		http.Error(w, "Failed to retrieve the file", http.StatusBadRequest)
		return
	}
	defer optionFile.Close()
	reader = csv.NewReader(optionFile)
	optionRecords, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "Failed to parse the CSV file", http.StatusBadRequest)
		return
	}
	for _, optionRecord := range optionRecords {

		qid := optionRecord[2]
		optionText := optionRecord[0]
		optionImage := optionRecord[1]

		// 	// Find the question in MongoDB by qid
		filter := bson.M{"q_id": mid[qid]}
		update := bson.M{
			"$push": bson.M{"options": models.Option{
				Text:  optionText,
				Image: optionImage,
			}},
		}

		_, err = questionCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("Failed to update question with qid '%s': %v\n", qid, err)
		} else {
			log.Printf("Updated question with qid '%s'\n", qid)
		}
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
}
