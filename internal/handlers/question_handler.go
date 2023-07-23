package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	database "testify/database"
	models "testify/internal/models"
	"time"

	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var questionCollection *mongo.Collection = database.OpenCollection(database.Client, "question")

func GetQuestions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	cur, err := questionCollection.Find(ctx, bson.M{})
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

	data, err := json.Marshal(questions)
	if err != nil {
		http.Error(w, "Failed to serialize questions", http.StatusInternalServerError)
		return
	}
	// Set response headers and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	return
}

func CreateQuestion(w http.ResponseWriter, r *http.Request) {
	var question models.Question
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	if err := json.NewDecoder(r.Body).Decode(&question); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		print(err.Error())
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
	err := json.NewDecoder(r.Body).Decode(&updatedQuestion)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		fmt.Fprintf(w, "Invalid request body")
		return
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
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error updating question")
		return
	}

	if result.ModifiedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Question not found")
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
