package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	database "testify/database"
	models "testify/internal/models"
	utility "testify/internal/utility"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()
var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
var user models.User
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

type QuestionStat struct {
	QuestionID    string `json:"qid"`
	QuestionText  string `json:"question"`
	UserAnswer    string `json:"user_answer"`
	CorrectAnswer string `json:"correct_answer"`
	MarksObtained int    `json:"marks_obtained"`
}

type SignInData struct {
	User_ID        string `json:"user_id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Token          string `json:"token"`
	ProfilePicture string `json:"profile_picture"`
	Wallet         int    `json:"wallet"`
	Purchased      bool   `json:"purchased"`
}

type PasswordChange struct {
	ExistingPassword string `json:"existing_password"`
	NewPassword      string `json:"new_password"`
}

// HashPassword is used to encrypt the password before it is stored in the DB
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

// VerifyPassword checks the input password while verifying it with the passward in the DB.
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("login or passowrd is incorrect")
		check = false
	}

	return check, msg
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// validation and password hashing
	validationErr := validate.Struct(user)
	if validationErr != nil {
		http.Error(w, "Fields not valid", http.StatusBadRequest)
		return
	}

	// Password Hashing
	password := HashPassword(*user.Password)
	user.Password = &password

	// Checking if user already exists
	alreadyExists, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	defer cancel()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if alreadyExists > 0 {
		http.Error(w, "User already exists!", http.StatusConflict)
		return
	}

	user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.ID = primitive.NewObjectID()
	user.User_id = user.ID.Hex()
	user.QuestionPapers = make(map[string]map[string]string, 0)
	user.Profile = "https://static.vecteezy.com/system/resources/previews/005/544/718/original/profile-icon-design-free-vector.jpg"
	user.Purchased = false

	token, refreshToken, _ := utility.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, user.User_id)
	user.Token = &token
	user.Refresh_token = &refreshToken
	user.Wallet = 0

	referral := user.ReferralCode
	if len(referral) == 18 && referral[:8] == "testify@" {
		number := referral[8:]
		var referer models.User
		err := userCollection.FindOne(ctx, bson.M{"phone": number}).Decode(&referer)
		defer cancel()
		if err != nil {
			http.Error(w, "Invalid Referral Code", http.StatusNotFound)
			return
		}
		referer.Wallet += 200
		filter := bson.M{"user_id": referer.User_id}
		update := bson.M{"$set": bson.M{
			"wallet": referer.Wallet,
		}}

		result, err := userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error updating Referer Wallet")
			return
		}
		fmt.Println(result)
	} else if referral != "" {
		http.Error(w, "Invalid Referral Code", http.StatusNotFound)
		return
	}

	// Create the user in the database
	insertResult, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer cancel()
	println(insertResult)

	w.Header().Set("Content-Type", "application/json")
	data := SignInData{
		Token:          token,
		User_ID:        user.User_id,
		FirstName:      *user.First_name,
		LastName:       *user.Last_name,
		Email:          *user.Email,
		Phone:          *user.Phone,
		ProfilePicture: user.Profile,
		Wallet:         user.Wallet,
		Purchased:      user.Purchased,
	}
	jsonResp, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonResp)
	w.WriteHeader(http.StatusOK)

}

func Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var foundUser models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Interal Server Error", http.StatusInternalServerError)
		return
	}
	err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
	defer cancel()
	if err != nil {
		http.Error(w, "Email or Password is incorrect", http.StatusUnauthorized)
		return
	}
	passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
	defer cancel()
	if passwordIsValid != true {
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	token, refreshToken, _ := utility.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
	utility.UpdateAllTokens(token, refreshToken, foundUser.User_id)

	w.Header().Set("Content-Type", "application/json")
	data := SignInData{
		Token:          token,
		User_ID:        foundUser.User_id,
		FirstName:      *foundUser.First_name,
		LastName:       *foundUser.Last_name,
		Email:          *foundUser.Email,
		Phone:          *foundUser.Phone,
		ProfilePicture: foundUser.Profile,
		Wallet:         foundUser.Wallet,
		Purchased:      foundUser.Purchased,
	}
	jsonResp, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonResp)
	w.WriteHeader(http.StatusOK)
	return
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {

	// Extract the JWT token from the Authorization header
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate the JWT token
	claims, errMsg := utility.ValidateToken(tokenString)
	if errMsg != "" {
		http.Error(w, errMsg, http.StatusUnauthorized)
		return
	}

	// Delete the user with the email from the token
	email := claims.Email
	if email == "" {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	deleteResult, err := userCollection.DeleteOne(ctx, bson.M{"email": email})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if deleteResult.DeletedCount == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User deleted successfully")
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
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
	cur, err := userCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	users := []models.User{}
	for cur.Next(ctx) {
		var user models.User
		err := cur.Decode(&user)
		if err != nil {
			http.Error(w, "Interal Server Error", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}
	if err := cur.Err(); err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}

	totalUsers, err := userCollection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}

	// Respond with the questions and total number of questions
	response := map[string]interface{}{
		"users":      users,
		"totalUsers": totalUsers,
	}
	// Set response headers and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func SubmitQPaper(w http.ResponseWriter, r *http.Request) {
	paperID := chi.URLParam(r, "paper_id")
	userID := chi.URLParam(r, "user_id")

	var user models.User
	err := userCollection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	m := make(map[string]string)
	err = json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, "Error decoding JSON data", http.StatusBadRequest)
		return
	}
	user.QuestionPapers[paperID] = m
	fmt.Println(user)

	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": bson.M{
		"questionPapers": user.QuestionPapers,
	}}

	result, err := userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error updating question")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func GetPaperStats(w http.ResponseWriter, r *http.Request) {
	paperID := chi.URLParam(r, "paper_id")
	userID := chi.URLParam(r, "user_id")

	var user models.User
	err := userCollection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	userChoices, found := user.QuestionPapers[paperID]
	if found {
		var PaperStats []QuestionStat
		for qid, userAns := range userChoices {
			var qStat QuestionStat
			var question models.Question
			qStat.QuestionID = qid
			qStat.UserAnswer = userAns
			err := questionCollection.FindOne(
				ctx,
				bson.M{"q_id": qid},
			).Decode(&question)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprintf(w, "Question not found"+err.Error())
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Error retrieving question")
					fmt.Println(err)
				}
				return
			}
			qStat.CorrectAnswer = question.CorrectAnswer
			qStat.QuestionText = question.Question
			qStat.MarksObtained = 0
			if qStat.CorrectAnswer == qStat.UserAnswer {
				qStat.MarksObtained = 4
			}
			PaperStats = append(PaperStats, qStat)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PaperStats)
	} else {
		http.Error(w, "You have not attempted this Question Paper", http.StatusNotFound)
		return
	}
}

func UpdateProfilePic(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "user_id")
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	profileImage := r.MultipartForm.File["profileImage"][0]
	imageName := fmt.Sprintf("profile%s.jpg", userId)
	err = utility.SaveImageToFile(profileImage, imageName)
	if err != nil {
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}
	imageURL := fmt.Sprintf("%s/image/%s", os.Getenv("BACKEND_URL"), imageName)

	filter := bson.M{"user_id": userId}
	update := bson.M{"$set": bson.M{
		"profile": imageURL,
	}}

	result, err := userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error updating profile image")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

}

func PurchaseCourse(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "user_id")
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	fmt.Println(user.Wallet)
	if user.Wallet < 199 {
		http.Error(w, "Insufficent Wallet Balance. Please add money to your wallet", http.StatusNotAcceptable)
		return
	}
	user.Wallet -= 199
	user.Purchased = true
	result, err := userCollection.UpdateOne(ctx, bson.M{"user_id": userId}, bson.M{"$set": bson.M{
		"wallet":    user.Wallet,
		"purchased": user.Purchased,
	}})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error Purchasing :(")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "user_id")
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var passwordChanger PasswordChange
	if err := json.NewDecoder(r.Body).Decode(&passwordChanger); err != nil {
		http.Error(w, "Interal Server Error", http.StatusInternalServerError)
		return
	}
	err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	passwordIsValid, msg := VerifyPassword(passwordChanger.ExistingPassword, *user.Password)
	defer cancel()
	if passwordIsValid != true {
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}
	newPass := HashPassword(passwordChanger.NewPassword)
	user.Password = &newPass

	result, err := userCollection.UpdateOne(ctx, bson.M{"user_id": userId}, bson.M{"$set": bson.M{
		"password": user.Password,
	}})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
