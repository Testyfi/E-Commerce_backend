package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	s3 "testify/aws"
	database "testify/database"
	models "testify/internal/models"
	utility "testify/internal/utility"
	httpClient "testify/internal/utility/http"
	"time"

	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var questionCollection *mongo.Collection = database.OpenCollection(database.Client, "question")
var testpaperCollection *mongo.Collection = database.OpenCollection(database.Client, "testpaper")
var usermaxscoreCollection *mongo.Collection = database.OpenCollection(database.Client, "usermaxscore")
var testdetailsCollection *mongo.Collection = database.OpenCollection(database.Client, "testdetails")
var totaluserCollection *mongo.Collection = database.OpenCollection(database.Client, "totaluser")
var qpaperCollection *mongo.Collection = database.OpenCollection(database.Client, "qpaper")
var questionsAWS_S3_API string = "https://testify-jee.s3.ap-south-1.amazonaws.com/assets/questions"

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
	findOptions.SetSort(bson.M{"created_at": -1})

	cur, err := questionCollection.Find(context.Background(), bson.M{}, findOptions)
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
	alreadyExists, err := questionCollection.CountDocuments(context.Background(), bson.M{"question": question.Question})

	if err != nil {
		http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
		return
	}
	if alreadyExists > 0 {
		http.Error(w, "Question already exists!", http.StatusConflict)
		return
	}
	if question.Type != "List-Type" {
		question.List1 = []string{}
		question.List2 = []string{}
	}
	question.ID = primitive.NewObjectID()
	question.Q_id = question.ID.Hex()
	question.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	question.UsedBy = []string{}
	questionImages := r.MultipartForm.File["questionImages"]
	question.Images = []string{}
	question.Solution = ""
	solutionImage := r.MultipartForm.File["solutionImage"]
	if len(solutionImage) > 0 {
		fileHeader := solutionImage[0]
		imagePath := fmt.Sprintf("%ssol.png", question.ID.Hex())
		err := utility.SaveImageToFile(fileHeader, imagePath, question.Q_id, "questions")
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to save Solution Image File", http.StatusInternalServerError)
			return
		}
		question.Solution = imagePath
	}
	for i, fileHeader := range questionImages {
		// Save the uploaded file to the "assets" directory
		imageName := fmt.Sprintf("%s%d.png", question.Q_id, i)
		err := utility.SaveImageToFile(fileHeader, imageName, question.Q_id, "questions")
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}

		// Update the Images array with the path to the saved image
		question.Images = append(question.Images, imageName)
	}

	if question.Type == "Multiple Choice" || question.Type == "Single Correct" {
		optionAImage := r.MultipartForm.File["optionAImage"]
		optionBImage := r.MultipartForm.File["optionBImage"]
		optionCImage := r.MultipartForm.File["optionCImage"]
		optionDImage := r.MultipartForm.File["optionDImage"]

		// Process option A image
		if len(optionAImage) > 0 {
			fileHeader := optionAImage[0]
			imagePath := fmt.Sprintf("%sA.png", question.ID.Hex())
			err := utility.SaveImageToFile(fileHeader, imagePath, question.Q_id, "questions")
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
			imagePath := fmt.Sprintf("%sB.png", question.ID.Hex())
			err := utility.SaveImageToFile(fileHeader, imagePath, question.Q_id, "questions")
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
			imagePath := fmt.Sprintf("%sC.png", question.ID.Hex())
			err := utility.SaveImageToFile(fileHeader, imagePath, question.Q_id, "questions")
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
			imagePath := fmt.Sprintf("%sD.png", question.ID.Hex())
			err := utility.SaveImageToFile(fileHeader, imagePath, question.Q_id, "questions")
			if err != nil {
				http.Error(w, "Failed to save option D image", http.StatusInternalServerError)
				return
			}
			// Update the Option D Image field with the path to the saved image
			question.Options[3].Image = imagePath
		}
	}

	// Create the question in the database
	insertResult, err := questionCollection.InsertOne(context.Background(), question)
	if err != nil {
		http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insertResult)
	w.WriteHeader(http.StatusOK)
	return
}

func GetQuestionByID(w http.ResponseWriter, r *http.Request) {
	questionID := chi.URLParam(r, "id")
	question := models.Question{}
	err := questionCollection.FindOne(
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

func DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	questionID := chi.URLParam(r, "id")

	result, err := questionCollection.DeleteOne(context.Background(), bson.M{"q_id": questionID})
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

	result, err := questionCollection.UpdateOne(context.Background(), filter, update)
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
		err := utility.DeleteQuestionImagesByQID(qid)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error deleting questions")
			return
		}
	}

	result, err := questionCollection.DeleteMany(context.Background(), bson.M{"q_id": bson.M{"$in": ids}})
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

	qFile, _, err := r.FormFile("questionCsvFile")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to retrieve the file", http.StatusBadRequest)
		return
	}
	defer qFile.Close()

	// Parse the CSV file
	reader := csv.NewReader(qFile)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to parse the CSV file", http.StatusBadRequest)
		return
	}

	mid := make(map[string]int)
	oid := make(map[string]int)

	// Insert CSV records in the database
	questions := []models.Question{}

	for _, record := range records {

		question := models.Question{
			Question:       record[0],
			Type:           record[2],
			Subject_Tags:   strings.Split(record[3], ", "),
			Q_id:           record[4],
			ID:             primitive.NewObjectID(),
			CorrectAnswer:  record[6],
			Solution:       record[5],
			Created_at:     time.Now(),
			Options:        make([]models.Option, 4),
			CorrectAnswers: []string{},
			Images:         strings.Split(record[1], ", "),
		}
		question.Options[0].Text = ""
		question.Options[1].Text = ""
		question.Options[2].Text = ""
		question.Options[3].Text = ""
		question.Options[0].Image = ""
		question.Options[1].Image = ""
		question.Options[2].Image = ""
		question.Options[3].Image = ""

		if len(question.Images) > 0 && question.Images[0] == "" {
			question.Images = []string{}
		}

		if question.Type == "Multiple Choice" {
			question.CorrectAnswers = strings.Split(record[6], ", ")
		}

		qid := question.ID.Hex()
		mid[record[4]] = len(questions)
		oid[record[4]] = 0
		question.Q_id = qid
		// Checking if question already exists
		alreadyExists, err := questionCollection.CountDocuments(context.Background(), bson.M{"question": question.Question})

		if err != nil {
			fmt.Println(err)
			http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
			return
		}
		if alreadyExists > 0 {
			fmt.Println(err)
			http.Error(w, "Question already exists!", http.StatusConflict)
			return
		}
		questions = append(questions, question)
	}
	optionFile, _, err := r.FormFile("optionCsvFile")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to retrieve the file", http.StatusBadRequest)
		return
	}
	defer optionFile.Close()
	reader = csv.NewReader(optionFile)
	optionRecords, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to parse the CSV file", http.StatusBadRequest)
		return
	}
	for _, optionRecord := range optionRecords {

		qid := optionRecord[2]
		optionText := optionRecord[0]
		optionImage := optionRecord[1]
		questions[mid[qid]].Options[oid[qid]].Text = optionText
		questions[mid[qid]].Options[oid[qid]].Image = optionImage
		oid[qid]++
	}

	for i := 0; i < len(questions); i++ {

		for j := 0; j < len(questions[i].Images); j++ {
			resp, err := http.Get(questions[i].Images[j])
			if err != nil {
				fmt.Println(err)
				http.Error(w, "Some Images were not able to be added", http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				http.Error(w, "Error adding Images. Please check the image URLs are correct", http.StatusInternalServerError)
				return
			}

			var imageBuffer bytes.Buffer
			_, err = io.Copy(&imageBuffer, resp.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			imageName := fmt.Sprintf("%s%v.png", questions[i].Q_id, j)
			questions[i].Images[j] = imageName
			imageName = fmt.Sprintf("%s/%s/%s/%s", "assets", "questions", questions[i].Q_id, imageName)
			s3.URLImageUpooad("testify-jee", imageName, s3.CreateSession(utility.AwsConfig), utility.AwsConfig, imageBuffer)

		}

		if questions[i].Solution != "" {
			resp, err := http.Get(questions[i].Solution)
			if err != nil {
				fmt.Println(err)
				http.Error(w, "Some Images were not able to be added", http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				http.Error(w, "Error adding Images. Please check the image URLs are correct", http.StatusInternalServerError)
				return
			}

			var imageBuffer bytes.Buffer
			_, err = io.Copy(&imageBuffer, resp.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			imageName := fmt.Sprintf("%ssol.png", questions[i].Q_id)
			questions[i].Solution = imageName
			imageName = fmt.Sprintf("%s/%s/%s/%s", "assets", "questions", questions[i].Q_id, imageName)
			s3.URLImageUpooad("testify-jee", imageName, s3.CreateSession(utility.AwsConfig), utility.AwsConfig, imageBuffer)
		}

		for j := 0; j < 4; j++ {
			if len(questions[i].Options[j].Image) > 0 {

				resp, err := http.Get(questions[i].Options[j].Image)
				if err != nil {
					fmt.Println(err)
					http.Error(w, "Some Images were not able to be added", http.StatusInternalServerError)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					fmt.Println(err)
					http.Error(w, "Error adding Images. Please check the image URLs are correct", http.StatusInternalServerError)
					return
				}

				var imageBuffer bytes.Buffer
				_, err = io.Copy(&imageBuffer, resp.Body)
				if err != nil {
					fmt.Println(err)
					http.Error(w, "Error adding Images.", http.StatusInternalServerError)
					return
				}
				var opt string
				switch j {
				case 0:
					opt = "A"
				case 1:
					opt = "B"
				case 2:
					opt = "C"
				case 3:
					opt = "D"
				}
				imageName := fmt.Sprintf("%s%s.png", questions[i].Q_id, opt)
				questions[i].Options[j].Image = imageName
				imageName = fmt.Sprintf("%s/%s/%s/%s", "assets", "questions", questions[i].Q_id, imageName)
				s3.URLImageUpooad("testify-jee", imageName, s3.CreateSession(utility.AwsConfig), utility.AwsConfig, imageBuffer)
			}
		}

		// Inserting Question
		_, err = questionCollection.InsertOne(context.Background(), questions[i])
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to insert record into the database", http.StatusInternalServerError)
			return
		}
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
}

var Topics = []string{
	"Kinematics", "Newton’s laws of motion",
	"Work Power Energy", "Systems of particles and Rotational Motion", "Gravitation", "Mechanical Properties of Solids", "Mechanical Properties of Fluids", "Thermal Properties of Matter", "Thermodynamics", "Kinetic Theory", "Oscillations", "Waves", "ELECTRIC CHARGES AND FIELDS", "ELECTROSTATIC POTENTIAL AND CAPACITANCE", "CURRENT ELECTRICITY", "MOVING CHARGES AND MAGNETISM", "MAGNETISM AND MATTER", "ELECTROMAGNETIC INDUCTION", "ALTERNATING CURRENT", "ELECTROMAGNETIC WAVES", "RAY OPTICS AND OPTICAL INSTRUMENTS", "WAVE OPTICS", "DUAL NATURE OF RADIATION AND MATTER", "SEMICONDUCTOR", "MODERN PHYSICS", "Some Basic,Mole Concept,Balance Equations", "Gaseous and Liquids States", "Atomic Structure", "Chemical Bonding and Molecular Structure", "Energetics", "Equilibrium", "ElectroChemistry", "Chamical Kinetics", "Solid State", "Solutions", "Nuclear Chemistry", "S and P block Elements", "D and F block Elements", "Metallurgy", "Principles of Qualitative Analysis", "General Organic Chemistry", "Hydrocarbons", "Organic Compunds Containing Halogens", "Alcohols Phenols Ethers", "Aldehyde and Ketones", "Carboxylic Acids and Derivatives", "Organic Compunds Containing Nitrogen", "Practical Organic Chemistry", "Sets, Relation and Function", "Complex Number", "Quadratic Equations", "Gravitation", "Arithmetic and Geometric Progressions", "Logarithms", "Straight Line", "Circle", "Parabola", "Ellipse", "Hyperbola", "Permutation and Combinations", "Binomial Theorem", "Trigonometry", "Probability", "Matrices and Determinant", "Limits, continuity and Differentiability", "Differentiations", "Applications of Differentiations", "Integrals", "Application of Integrals", "Differential equations", "Vectors Algebra", "Three Dimensional Geometry",
}

// 6-numerical
// 6-single
// 4-Multiple
// 4-list

type CreateQPaperHelper struct {
	Questions []int `json:"questions"`
}
func CreateYourQPaperDataBase(w http.ResponseWriter, r *http.Request){
	user, ok := r.Context().Value(models.ContextUser).(models.User)

	if !ok {
		httpClient.RespondError(w, http.StatusBadRequest, "Failed to retrieve user", fmt.Errorf("failed to retrieve user"))
		fmt.Println("some error on user fething statics")
		return 
	}
	var t struct {
		
		Difficulty string `json:"difficulty"`
		Duration string `json:"duration"`
		QuestionId []string `json:"questionid"`

	}


	err := json.NewDecoder(r.Body).Decode(&t)
	//fmt.Println(t)

	if err != nil {
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid test or user", err)
		return
	}
	var qpaper models.QPaper
	qpaper.ID = primitive.NewObjectID()
	qpaper.Qpid = qpaper.ID.Hex()
    qpaper.Difficulty=t.Difficulty
	qpaper.Duration=t.Duration
	qpaper.UserPhone=*user.Phone
	qpaper.Name="Create Your Test "+strconv.Itoa(TotalCreatedTest(*user.Phone)+1)
    qpaper.Questions=t.QuestionId
	qpaperCollection.InsertOne(context.Background(),qpaper)
	httpClient.RespondSuccess(w, "Success")
}
func TotalCreatedTest(phone string)int {


	filter := bson.D{{"userphone", phone}}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	
	// Count the documents in the collection with the specified filter
	totalDocuments, err := qpaperCollection.CountDocuments(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	return int(totalDocuments)
}
func CreateYourTestAdvanced(w http.ResponseWriter, r *http.Request){
	//DistinctSubject_tags()
	//UpdateQuestionCollection()
	user, ok := r.Context().Value(models.ContextUser).(models.User)

	if !ok {
		httpClient.RespondError(w, http.StatusBadRequest, "Failed to retrieve user", fmt.Errorf("failed to retrieve user"))
		fmt.Println("some error on user fething statics")
		return 
	}
	var t struct {
		
		Topics []string `json:"topics"`
		Number int `json:"number"`
	}


	err := json.NewDecoder(r.Body).Decode(&t)
	//fmt.Println(t)

	if err != nil {
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid test or user", err)
		return
	}
	//fmt.Println(Questions(t.Number,t.Topics))
	var questions []models.Question=Questions(t.Number,t.Topics,*user.Phone)
	//questions=
	httpClient.RespondSuccess(w, questions)
	

	for i:=0;i<len(questions);i++{

		addusedby(*user.Phone,questions[i])
	}
	
}
func addusedby(phone string,question models.Question){
question.UsedBy = append(question.UsedBy, phone)

filter := bson.D{{"q_id", question.Q_id}}

	// Specify the update to be applied
	
	update := bson.D{
		{"$set", bson.D{
			{"usedby",question.UsedBy},
			// add more fields to update as needed
		}},
	}

	// Perform the updateMany operation
	result, err := questionCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		fmt.Println("Error updating documents:", err)
		return
	}

	// Print the number of documents updated
	fmt.Printf("Updated %v documents\n", result.ModifiedCount)
}

func UpdateQuestionCollection(){
     //valueTobeupdated:=[]string{"permutation and combination"}
	filter := bson.D{{"subject_tags", "permutation and combination"}}

	// Specify the update to be applied
	valuesToSearch := []string{"Permutation & Combination","Mathematics"}
	update := bson.D{
		{"$set", bson.D{
			{"subject_tags", valuesToSearch},
			// add more fields to update as needed
		}},
	}

	// Perform the updateMany operation
	result, err := questionCollection.UpdateMany(context.Background(), filter, update)
	if err != nil {
		fmt.Println("Error updating documents:", err)
		return
	}

	// Print the number of documents updated
	fmt.Printf("Updated %v documents\n", result.ModifiedCount)
}
func Questions(num int,subject []string,phone string) []models.Question{
	

	// Encode string array to BSON
	
	jeemains:=[]string{"jee mains","JEE MAINS","Jee Mains","jee Mains","jee mains","Jee Main"}
	
	ctx := context.Background()
	pipeline:=bson.A{
		bson.D{{"$match", bson.D{{"usedby", bson.D{{"$ne", phone}}}}}},
		bson.D{
        {"$match",
            bson.D{
                {"subject_tags",
                    bson.D{
                        {"$ne",
                            bson.D{
                                {"$in",
                                    jeemains,
                                },
                            },
                        },
                    },
                },
            },
        },
    },
		bson.D{
			{"$match",
				bson.D{
					{"subject_tags",
						bson.D{
							{"$in",
								subject,
							},
						},
					},
				},
			},
		},
		bson.D{
			{"$group",
				bson.D{
					{"_id",
						bson.A{
							"$type",
						},
					},
					{"questions", bson.D{{"$push", "$$ROOT"}}},
				},
			},
		},
		bson.D{
			{"$match",
				bson.D{
					{"$or",
						bson.A{
							bson.D{{"_id", "Single Correct"}},
							bson.D{{"_id", "Multiple Choice"}},
							bson.D{{"_id", "Numerical Answer"}},
						},
					},
				},
			},
		},
		bson.D{
			{"$project",
				bson.D{
					{"questions",
						bson.D{
							{"$slice",
								bson.A{
									"$questions",
									num,
								},
							},
						},
					},
				},
			},
		},
		bson.D{{"$unwind", "$questions"}},
		bson.D{{"$replaceRoot", bson.D{{"newRoot", "$questions"}}}},
	}
	cursor, err := questionCollection.Aggregate(context.Background(), pipeline)
		if err != nil {
			//http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
			//return
		}
		defer cursor.Close(ctx)
		var questions []models.Question
	for cursor.Next(ctx) {
		var result models.Question
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		questions = append(questions, result)
	}
		// Check for errors from iterating over cursor
		if err := cursor.Err(); err != nil {
			log.Fatal(err)
		}
return questions
}
func FindAllCreatedTest(w http.ResponseWriter, r *http.Request){


	user, ok := r.Context().Value(models.ContextUser).(models.User)

	if !ok {
		httpClient.RespondError(w, http.StatusBadRequest, "Failed to retrieve user", fmt.Errorf("failed to retrieve user"))
		fmt.Println("some error on user fething statics")
		return 
	}
	//var allcreatedtest []models.QPaper
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//err = client.Connect(ctx)
	
	//defer client.Disconnect(ctx)
	filter := bson.D{{"userphone", *user.Phone}}
	options := options.Find().SetSort(bson.D{{"_id", -1}})
	// Find questions in the collection with the specified filter
	cursor, err := qpaperCollection.Find(ctx, filter,options)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Iterate through the result set and print each question
	var qp []models.QPaper
	if err := cursor.All(ctx, &qp); err != nil {
		log.Fatal(err)
	}
    
	httpClient.RespondSuccess(w, qp)
}
func FindCreatedTestQuestions(w http.ResponseWriter, r *http.Request){


	var t struct {
		
		
		Questions []string `json:"questions"`
	}


	err := json.NewDecoder(r.Body).Decode(&t)
	//fmt.Println(t)

	if err != nil {
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid test or user", err)
		return
	}
	//var allcreatedtest []models.QPaper
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//err = client.Connect(ctx)
	
	//defer client.Disconnect(ctx)
	filter := bson.D{{"q_id", bson.D{{"$in", t.Questions}}}}

	// Find questions in the collection with the specified filter
	cursor, err :=questionCollection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Iterate through the result set and print each question
	var qp []models.Question
	if err := cursor.All(ctx, &qp); err != nil {
		log.Fatal(err)
	}

	httpClient.RespondSuccess(w, qp)
}
func CreateYourTestJeeMains(w http.ResponseWriter, r *http.Request){

	user, ok := r.Context().Value(models.ContextUser).(models.User)

	if !ok {
		httpClient.RespondError(w, http.StatusBadRequest, "Failed to retrieve user", fmt.Errorf("failed to retrieve user"))
		fmt.Println("some error on user fething statics")
		return 
	}
	var t struct {
		
		Topics []string `json:"topics"`
		Number int `json:"number"`
	}


	err := json.NewDecoder(r.Body).Decode(&t)
	//fmt.Println(t)

	if err != nil {
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid test or user", err)
		return
	}
	var questions []models.Question=JeeMainsQuestions(t.Topics,t.Number,*user.Phone)
	httpClient.RespondSuccess(w, questions)
	for i:=0;i<len(questions);i++{

		addusedby(*user.Phone,questions[i])
	}
}
func JeeMainsQuestions(topics []string,numberofquestion int,phone string)[]models.Question{
	ctx := context.Background()
	jeemains:=[]string{"Jee mains",
	"jee mains",
	"jee Mains",
	"Jee Mains","JEE mains",
	"Probability,jee mains"}
	
	pipeline:=bson.A{
    bson.D{{"$match", bson.D{{"usedby", bson.D{{"$ne", phone}}}}}},
	bson.D{
        {"$match",
            bson.D{
                {"subject_tags",
                    bson.D{
                        {"$in",
                            jeemains,
                        },
                    },
                },
            },
        },
    },
    bson.D{
        {"$match",
            bson.D{
                {"subject_tags",
                    bson.D{
                        {"$in",
                            topics,
                        },
                    },
                },
            },
        },
    },
    bson.D{{"$sample", bson.D{{"size", numberofquestion}}}},
}
cursor, err := questionCollection.Aggregate(context.Background(), pipeline)
		if err != nil {
			//http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
			//return
		}
		defer cursor.Close(ctx)
		var questions []models.Question
	for cursor.Next(ctx) {
		var result models.Question
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		questions = append(questions, result)
	}
		// Check for errors from iterating over cursor
		if err := cursor.Err(); err != nil {
			log.Fatal(err)
		}
 return questions

}
func DistinctSubject_tags(){

	

	
//fmt.Println("called")
	
    ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	//valuesToSearch := []string{"mathematics","Mathematics"}
	//cursor,err:=questionCollection.Find(ctx,bson.M{"subject_tags":bson.M{"$in": valuesToSearch}})
	cursor,err:=questionCollection.Find(ctx,bson.M{})
    defer cursor.Close(ctx)
	if err !=nil{

		
	}
	var questions []models.Question
	if err =cursor.All(ctx,&questions);err!=nil{
		
	}
    // fmt.Println(questions)
	 
	var str [] string
	for i:=0;i<len(questions);i++{

       var temp=questions[i].Subject_Tags
	   for j:=0;j<len(temp);j++{
        t:=false
		for k:=0;k<len(str);k++{

               if(temp[j]==str[k]){
				t=true;
			   }
		}
		if(!t){
			str = append(str, temp[j])
		}
	   }
	}
	for i:=0;i<len(str);i++{

		fmt.Println(str[i])
	}
	

}
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

