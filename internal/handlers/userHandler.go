package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	database "testify/database"
	models "testify/internal/models"
	utility "testify/internal/utility"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()
var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
var user models.User
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

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

	token, refreshToken, _ := utility.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, user.User_id)
	user.Token = &token
	user.Refresh_token = &refreshToken

	// Create the user in the database
	insertResult, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer cancel()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insertResult)
	w.WriteHeader(http.StatusOK)
	return
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
	json.NewEncoder(w).Encode(token)
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
	println("Email is " + email)

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
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	cur, err := userCollection.Find(ctx, bson.M{})
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
	// Serialize users to JSON
	data, err := json.Marshal(users)
	if err != nil {
		http.Error(w, "Failed to serialize users", http.StatusInternalServerError)
		return
	}
	// Set response headers and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	return
}
