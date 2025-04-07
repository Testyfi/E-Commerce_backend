package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

var productsCollection *mongo.Collection = database.OpenCollection(database.Client, "products")


func GetProducts(w http.ResponseWriter, r *http.Request) {
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
	findOptions.SetSort(bson.M{"created_at": -1})

	cur, err := productsCollection.Find(context.Background(), bson.M{}, findOptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())

	questions := []models.Question{}
	for cur.Next(context.Background()) {
		var question models.Question
		err := cur.Decode(&question)
		if err != nil {
			http.Error(w, "Interal Server Error", http.StatusInternalServerError)
			return
		}

		for i := 0; i < len(question.Images); i++ {
			question.Images[i] = fmt.Sprintf("%s/%s/%s", questionsAWS_S3_API, question.Q_id, question.Images[i])
		}
		if question.Type == "Single Correct" || question.Type == "Multiple Choice" {
			if len(question.Options[0].Image) > 0 {
				question.Options[0].Image = fmt.Sprintf("%s/%s/%s", questionsAWS_S3_API, question.Q_id, question.Options[0].Image)
			}
			if len(question.Options[1].Image) > 0 {
				question.Options[1].Image = fmt.Sprintf("%s/%s/%s", questionsAWS_S3_API, question.Q_id, question.Options[1].Image)
			}
			if len(question.Options[2].Image) > 0 {
				question.Options[2].Image = fmt.Sprintf("%s/%s/%s", questionsAWS_S3_API, question.Q_id, question.Options[2].Image)
			}
			if len(question.Options[3].Image) > 0 {
				question.Options[3].Image = fmt.Sprintf("%s/%s/%s", questionsAWS_S3_API, question.Q_id, question.Options[3].Image)
			}
		}
		questions = append(questions, question)
	}
	if err := cur.Err(); err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}

	totalQuestions, err :=productsCollection.CountDocuments(context.Background(), bson.M{})
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

func AddProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	data := r.FormValue("data")
	//fmt.Println(data)
	err = json.Unmarshal([]byte(data), &product)
	if err != nil {
		http.Error(w, "Failed to parse JSON data", http.StatusBadRequest)
		return

	}
	
	//fmt.Println("hellow")
	//fmt.Println(product)

	// Checking if question already exists
	alreadyExists, err := productsCollection.CountDocuments(context.Background(), bson.M{"productName": product.Product})

	if err != nil {
		http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
		return
	}
	if alreadyExists > 0 {
		http.Error(w, "Product already exists!", http.StatusConflict)
		return
	}
	
	product.ID = primitive.NewObjectID()
	product.P_id = product.ID.Hex()
	product.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	
	productImages := r.MultipartForm.File["productImages"]
	product.Images = []string{}
	
	
	
	for i, fileHeader := range productImages {
		// Save the uploaded file to the "assets" directory
		imageName := fmt.Sprintf("%s%d.jpg", product.P_id, i)
		err := utility.SaveImageToCloudFlare(fileHeader, imageName, product.P_id, "products")
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}

		// Update the Images array with the path to the saved image
		product.Images = append(product.Images, imageName)
	}

	

	// Create the question in the database
	insertResult, err := productsCollection.InsertOne(context.Background(), product)
	if err != nil {
		http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insertResult)
	w.WriteHeader(http.StatusOK)
	return
}

func GetProductByID(w http.ResponseWriter, r *http.Request) {
	questionID := chi.URLParam(r, "id")
	question := models.Question{}
	err := productsCollection.FindOne(
		context.Background(),
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
	for i := 0; i < len(question.Images); i++ {
		question.Images[i] = fmt.Sprintf("%s/%s/%s", questionsAWS_S3_API, questionID, question.Images[i])
	}

	if len(question.Options[0].Image) > 0 {
		question.Options[0].Image = fmt.Sprintf("%s/%s/%s", questionsAWS_S3_API, questionID, question.Options[0].Image)
	}
	if len(question.Options[1].Image) > 0 {
		question.Options[1].Image = fmt.Sprintf("%s/%s/%s", questionsAWS_S3_API, questionID, question.Options[1].Image)
	}
	if len(question.Options[2].Image) > 0 {
		question.Options[2].Image = fmt.Sprintf("%s/%s/%s", questionsAWS_S3_API, questionID, question.Options[2].Image)
	}
	if len(question.Options[3].Image) > 0 {
		question.Options[3].Image = fmt.Sprintf("%s/%s/%s", questionsAWS_S3_API, questionID, question.Options[3].Image)
	}

	if len(question.Solution) > 0 {
		question.Solution = fmt.Sprintf("%s/%s/%s", questionsAWS_S3_API, questionID, question.Solution)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(question)

}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	questionID := chi.URLParam(r, "id")

	result, err := productsCollection.DeleteOne(context.Background(), bson.M{"q_id": questionID})
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

func EditProduct(w http.ResponseWriter, r *http.Request) {

	questionID := chi.URLParam(r, "id")

	var updatedQuestion models.Question

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
	err = utility.DeleteQuestionImagesByQID(questionID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error deleting questions")
		return
	}
	if updatedQuestion.Type != "List-Type" {
		updatedQuestion.List1 = []string{}
		updatedQuestion.List2 = []string{}
	}
	questionImages := r.MultipartForm.File["questionImages"]
	updatedQuestion.Images = []string{}
	updatedQuestion.Solution = ""
	solutionImage := r.MultipartForm.File["solutionImage"]
	if len(solutionImage) > 0 {
		fileHeader := solutionImage[0]
		imagePath := fmt.Sprintf("%ssol.png", questionID)
		err := utility.SaveImageToFile(fileHeader, imagePath, questionID, "questions")
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to save Solution Image File", http.StatusInternalServerError)
			return
		}
		updatedQuestion.Solution = imagePath
	}
	for i, fileHeader := range questionImages {
		// Save the uploaded file to the "assets" directory
		imageName := fmt.Sprintf("%s%d.png", questionID, i)
		err := utility.SaveImageToFile(fileHeader, imageName, questionID, "questions")
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}
		// Update the Images array with the path to the saved image
		updatedQuestion.Images = append(updatedQuestion.Images, imageName)
	}

	if updatedQuestion.Type == "Multiple Choice" || updatedQuestion.Type == "Single Correct" {
		optionAImage := r.MultipartForm.File["optionAImage"]
		optionBImage := r.MultipartForm.File["optionBImage"]
		optionCImage := r.MultipartForm.File["optionCImage"]
		optionDImage := r.MultipartForm.File["optionDImage"]

		// Process option A image
		if len(optionAImage) > 0 {
			fileHeader := optionAImage[0]
			imagePath := fmt.Sprintf("%sA.png", questionID)
			err := utility.SaveImageToFile(fileHeader, imagePath, questionID, "questions")
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
			imagePath := fmt.Sprintf("%sB.png", questionID)
			err := utility.SaveImageToFile(fileHeader, imagePath, questionID, "questions")
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
			imagePath := fmt.Sprintf("%sC.png", questionID)
			err := utility.SaveImageToFile(fileHeader, imagePath, questionID, "questions")
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
			imagePath := fmt.Sprintf("%sD.png", questionID)
			err := utility.SaveImageToFile(fileHeader, imagePath, questionID, "questions")
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
		"question":       updatedQuestion.Question,
		"images":         updatedQuestion.Images,
		"type":           updatedQuestion.Type,
		"options":        updatedQuestion.Options,
		"correctanswer":  updatedQuestion.CorrectAnswer,
		"subject_tags":   updatedQuestion.Subject_Tags,
		"correctanswers": updatedQuestion.CorrectAnswers,
		"solution":       updatedQuestion.Solution,
		"list1":          updatedQuestion.List1,
		"list2":          updatedQuestion.List2,
	}}

	result, err := productsCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error updating question")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func DeleteManyProducts(w http.ResponseWriter, r *http.Request) {
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
		err := utility.DeleteQuestionImagesByQID(qid)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error deleting questions")
			return
		}
	}

	result, err := productsCollection.DeleteMany(context.Background(), bson.M{"q_id": bson.M{"$in": ids}})
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error deleting questions")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}




// 6-numerical
// 6-single
// 4-Multiple
// 4-list




/*
func CreateQPaper(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "user_id")
	var qpaper models.QPaper
	var createpaperhelper CreateQPaperHelper
	var questions []int

	err := userCollection.FindOne(context.Background(), bson.M{"user_id": userId}).Decode(&user)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&createpaperhelper)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		fmt.Fprintf(w, "Invalid request body")
		return
	}
	questions = createpaperhelper.Questions

	for i := 0; i < len(questions); i++ {
		var pipeline []primitive.M
		if (i >= 0 && i < 6) || (i >= 20 && i < 26) || (i >= 40 && i < 46) {
			pipeline = []bson.M{
				{
					"$match": bson.M{
						"usedby": bson.M{
							"$ne": userId,
						},
						"subject_tags": Topics[questions[i]],
						"type":         "Numerical Answer",
					},
				},
				{
					"$sample": bson.M{
						"size": 1,
					},
				},
			}
		} else if (i >= 6 && i < 12) || (i >= 26 && i < 32) || (i >= 46 && i < 52) {
			pipeline = []bson.M{
				{
					"$match": bson.M{
						"usedby": bson.M{
							"$ne": userId,
						},
						"subject_tags": Topics[questions[i]],
						"type":         "Single Correct",
					},
				},
				{
					"$sample": bson.M{
						"size": 1,
					},
				},
			}
		} else if (i >= 12 && i < 16) || (i >= 32 && i < 36) || (i >= 52 && i < 56) {
			pipeline = []bson.M{
				{
					"$match": bson.M{
						"usedby": bson.M{
							"$ne": userId,
						},
						"subject_tags": Topics[questions[i]],
						"type":         "Multiple Choice",
					},
				},
				{
					"$sample": bson.M{
						"size": 1,
					},
				},
			}
		} else {
			pipeline = []bson.M{
				{
					"$match": bson.M{
						"usedby": bson.M{
							"$ne": userId,
						},
						"subject_tags": Topics[questions[i]],
						"type":         "List-Type",
					},
				},
				{
					"$sample": bson.M{
						"size": 1,
					},
				},
			}
		}

		cursor, err := questionCollection.Aggregate(context.Background(), pipeline)
		if err != nil {
			http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		var question models.Question
		err = cursor.Decode(&question)
		if err != nil {
			http.Error(w, "No such question found for selected topics", http.StatusNotFound)
			return
		}
		question.UsedBy = append(question.UsedBy, userId)
		questionCollection.UpdateOne(context.Background(), bson.M{"q_id": question.Q_id}, bson.M{
			"$set": bson.M{
				"usedby": question.UsedBy},
		})
		qpaper.Questions = append(qpaper.Questions, question.Q_id)
	}

	qpaper.ID = primitive.NewObjectID()
	qpaper.Qpid = qpaper.ID.Hex()
	result, err := qpaperCollection.InsertOne(context.Background(), qpaper)
	if err != nil {
		http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.InsertedID)
}
*/

