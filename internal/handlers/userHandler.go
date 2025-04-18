package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	database "testify/database"
	models "testify/internal/models"
	utility "testify/internal/utility"
	httpClient "testify/internal/utility/http"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()
var user models.User
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var otpcollection *mongo.Collection = database.OpenCollection(database.Client, "otp")

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

func UserVerification(w http.ResponseWriter, r *http.Request) {
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
	alreadyExists, err := userCollection.CountDocuments(context.Background(), bson.M{"phone": user.Phone})
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
	user.Verified = false
	token, refreshToken, _ := utility.GenerateAllTokens( *user.First_name, *user.Last_name, user.User_id)
	user.Token = &token
	user.Refresh_token = &refreshToken
	user.Wallet = 0

	var otp models.OTP
	err = otpcollection.FindOne(context.Background(), bson.M{"phone": user.Phone}).Decode(&otp)
	if err != nil {
		http.Error(w, "Invalid OTP", http.StatusUnauthorized)
		return
	}

	if user.Verified == false && user.SecretCode == otp.SecretCode && user.Otp == otp.Otp {
		user.Verified = true

		referral := user.ReferralCode
		if len(referral) == 18 && referral[:8] == "testify@" {
			number := referral[8:]
			var referer models.User
			err := userCollection.FindOne(context.Background(), bson.M{"phone": number}).Decode(&referer)

			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusNotFound)
				return
			}
			referer.Wallet += 200
			filter := bson.M{"user_id": referer.User_id}
			update := bson.M{"$set": bson.M{
				"wallet": referer.Wallet,
			}}

			result, err := userCollection.UpdateOne(context.Background(), filter, update)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			fmt.Println(result)
		}

		result, err := userCollection.InsertOne(context.Background(), user)
		if err != nil {
			http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
			return
		}
		println(result)
		httpClient.RespondSuccess(w, "Verification Success.Please Login now")
		return
	}

	http.Error(w, "404 Not Found", http.StatusNotFound)
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	var otp models.OTP
	if err := json.NewDecoder(r.Body).Decode(&otp); err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Checking if user already exists
	
	// Checking if phone already exists
	alreadyExists, err := userCollection.CountDocuments(context.Background(), bson.M{"phone": otp.Phone})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if alreadyExists > 0 {
		http.Error(w, "Phone number already exists!", http.StatusConflict)
		return
	}

	alreadyExists, err = otpcollection.CountDocuments(context.Background(), bson.M{"phone": otp.Phone})
	fmt.Println(alreadyExists)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if alreadyExists > 0 {
		http.Error(w, "An OTP has already been sent. Please verify or wait for sometime to request OTP again", http.StatusConflict)
		return
	}

	//otp.SecretCode = utility.GenerateRandomCode()
	otp.Otp = utility.GenerateRandomCode()
	//verificationLink := fmt.Sprintf("Your OTP for Email verification is %s", otp.SecretCode)
	//err = utility.SendMail(verificationLink, *otp.Email, "Email Verification")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	utility.SMS("160066", []string{*otp.First_name, otp.Otp}, []string{*otp.Phone})

	otp.CreatedAt = time.Now()

	_, err = otpcollection.InsertOne(context.Background(), otp)
	if err != nil {
		http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
		return
	}
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "createdat", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(60), // 1 minutes in seconds
	}
	_, err = otpcollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		// Handle the error if creating the TTL index fails
		http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var foundUser models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err := userCollection.FindOne(context.Background(), bson.M{"phone": user.Phone}).Decode(&foundUser)
	if err != nil {
		http.Error(w, "Email or Password is incorrect", http.StatusUnauthorized)
		return
	}
	passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
	if passwordIsValid != true {
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}
	

	token, refreshToken, _ := utility.GenerateAllTokens( *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
	utility.UpdateAllTokens(token, refreshToken, foundUser.User_id)

	w.Header().Set("Content-Type", "application/json")
	data := SignInData{
		Token:          token,
		User_ID:        foundUser.User_id,
		FirstName:      *foundUser.First_name,
		LastName:       *foundUser.Last_name,
		
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

	deleteResult, err := userCollection.DeleteOne(context.Background(), bson.M{"email": email})
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

	cur, err := userCollection.Find(context.Background(), bson.M{}, findOptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())

	users := []models.User{}
	for cur.Next(context.Background()) {
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

	result, err := userCollection.UpdateOne(context.Background(), filter, update)
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
				context.Background(),
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

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	profileImage := r.MultipartForm.File["profileImage"][0]
	imageName := fmt.Sprintf("profile%s.jpg", userId)
	err = utility.SaveImageToFile(profileImage, imageName, userId, "users")
	if err != nil {
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}
	imageURL := fmt.Sprintf("%s/users/%s/%s", "https://testify-jee.s3.ap-south-1.amazonaws.com/assets", userId, imageName)

	filter := bson.M{"user_id": userId}
	update := bson.M{"$set": bson.M{
		"profile": imageURL,
	}}

	result, err := userCollection.UpdateOne(context.Background(), filter, update)
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

	err := userCollection.FindOne(context.Background(), bson.M{"user_id": userId}).Decode(&user)
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
	result, err := userCollection.UpdateOne(context.Background(), bson.M{"user_id": userId}, bson.M{"$set": bson.M{
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

	var passwordChanger PasswordChange
	if err := json.NewDecoder(r.Body).Decode(&passwordChanger); err != nil {
		http.Error(w, "Interal Server Error", http.StatusInternalServerError)
		return
	}
	err := userCollection.FindOne(context.Background(), bson.M{"user_id": userId}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	passwordIsValid, msg := VerifyPassword(passwordChanger.ExistingPassword, *user.Password)

	if passwordIsValid != true {
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}
	newPass := HashPassword(passwordChanger.NewPassword)
	user.Password = &newPass

	result, err := userCollection.UpdateOne(context.Background(), bson.M{"user_id": userId}, bson.M{"$set": bson.M{
		"password": user.Password,
	}})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")
	secret := r.URL.Query().Get("secret")

	err := userCollection.FindOne(context.Background(), bson.M{"user_id": userId}).Decode(&user)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if user.ResetCode == secret {
		pass := HashPassword(secret)
		user.Password = &pass
		user.ResetCode = "Not applicable"
		result, err := userCollection.UpdateOne(context.Background(), bson.M{"user_id": userId}, bson.M{"$set": bson.M{
			"password":  user.Password,
			"resetcode": user.ResetCode,
		}})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err)
			fmt.Fprintf(w, "Error resetting Password")
			return
		}
		fmt.Println(result)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Password Reset Successful. Your new Password is " + secret))
		return
	}
	http.Error(w, "404 Not Found. Probably you have already reset your password using this Link.", http.StatusNotFound)
}

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	type ForgotPass struct {
		Phone string `json:"phone"`
	}

	var helper ForgotPass
	if err := json.NewDecoder(r.Body).Decode(&helper); err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
    
	err := userCollection.FindOne(context.Background(), bson.M{"phone": helper.Phone}).Decode(&user)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	str:=generateRandomString(8)
	newPass := HashPassword(str)
	user.Password = &newPass

	result, err := userCollection.UpdateOne(context.Background(), bson.M{"phone": user.Phone}, bson.M{"$set": bson.M{
		"password": user.Password,
	}})
	if err != nil {
		//fmt.Println("some error")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		fmt.Fprintf(w, "Error Resetting Password. Please try again later")
		return
	}
	utility.SMS("163562", []string{*user.First_name, str}, []string{*user.Phone})
	fmt.Println(result)

	//fmt.Println(user)
/*
	user.ResetCode = utility.GenerateRandomCode()
	result, err := userCollection.UpdateOne(context.Background(), bson.M{"user_id": user.User_id}, bson.M{"$set": bson.M{
		"resetcode": user.ResetCode,
	}})
	if err != nil {
		//fmt.Println("some error")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		fmt.Fprintf(w, "Error Resetting Password. Please try again later")
		return
	}
	fmt.Println(result)
	
	resetLink := fmt.Sprint(os.Getenv("BACKEND_URL") + "/reset" + "?user_id=" + user.User_id + "&secret=" + user.ResetCode)
	err = utility.SendMail("click on this link to reset your password. "+resetLink+" Your new Password will be "+user.ResetCode+". If this was not requested by you, kindly ignore.", helper.Email, "Testify Password Reset Link")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		fmt.Fprintf(w, "Internal Server Error. Please try again later")
		return
	}
	//fmt.Println("Working")
	*/
	httpClient.RespondSuccess(w, "Your new password has been successfully sent to your mobile. Please check")
	//w.Write([]byte("Password Reset Link has been sent to your email."))
}
const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitBytes  = "0123456789"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateRandomString(length int) string {
	allBytes := letterBytes + digitBytes
	result := make([]byte, length)

	for i := range result {
		result[i] = allBytes[rand.Intn(len(allBytes))]
	}

	return string(result)
}
