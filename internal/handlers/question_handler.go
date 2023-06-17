package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	// Serialize questions to JSON
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
		context.TODO(),
		bson.D{{"_id", questionID}},
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

// func EditQuestion(w http.ResponseWriter, r *http.Request) {
// 	// Get the question ID from the request parameters
// 	id := chi.URLParam(r, "id")

// 	// Get the new question details from the request body
// 	var updatedQuestion models.Question
// 	if err := json.NewDecoder(r.Body).Decode(&updatedQuestion); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	// Retrieve the question from the database using the ID or any other unique identifier
// 	question, err := database.GetQuestionByID(id)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Update the question details based on the new values
// 	question.Question = updatedQuestion.Question
// 	question.Type = updatedQuestion.Type
// 	question.Options = updatedQuestion.Options
// 	question.CorrectAnswer = updatedQuestion.CorrectAnswer

// 	// Save the updated question back to the database
// 	err = db.UpdateQuestion(question)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Return the updated question as a JSON response
// 	json.NewEncoder(w).Encode(question)
// }
