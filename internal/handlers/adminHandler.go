package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	database "testify/database"
	models "testify/internal/models"
	utility "testify/internal/utility"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Data struct {
	Token     string `json:"token"`
	AdminName string `json:"adminName"`
}

var adminCollection *mongo.Collection = database.OpenCollection(database.Client, "admin")

func GetAdmins(w http.ResponseWriter, r *http.Request) {
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	cur, err := adminCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	admins := []models.Admin{}
	for cur.Next(ctx) {
		var admin models.Admin
		err := cur.Decode(&admin)
		if err != nil {
			http.Error(w, "Interal Server Error", http.StatusInternalServerError)
			return
		}
		admins = append(admins, admin)
	}
	// Serialize admins to JSON
	data, err := json.Marshal(admins)
	if err != nil {
		http.Error(w, "Failed to serialize admins", http.StatusInternalServerError)
		return
	}
	// Set response headers and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	return
}

func CreateAdmin(w http.ResponseWriter, r *http.Request) {
	var admin models.Admin
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	if err := json.NewDecoder(r.Body).Decode(&admin); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		print(err.Error())
		return
	}

	// Password Hashing
	password := HashPassword(*admin.Password)
	admin.Password = &password

	// Checking if admin already exists
	alreadyExists, err := adminCollection.CountDocuments(ctx, bson.M{"email": admin.Email})
	defer cancel()
	if err != nil {
		http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
		return
	}
	if alreadyExists > 0 {
		http.Error(w, "admin already exists!", http.StatusConflict)
		return
	}

	admin.ID = primitive.NewObjectID()
	admin.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	admin.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	admin.Admin_ID = admin.ID.Hex()
	token, refreshToken, _ := utility.GenerateAllAdminTokens(*admin.Email, *admin.AdminName, admin.Admin_ID)
	admin.Token = &token
	admin.Refresh_token = &refreshToken

	// Create the admin in the database
	insertResult, err := adminCollection.InsertOne(ctx, admin)
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

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var foundAdmin models.Admin
	var admin models.Admin
	if err := json.NewDecoder(r.Body).Decode(&admin); err != nil {
		http.Error(w, "Interal Server Error "+err.Error(), http.StatusInternalServerError)
		return
	}
	err := adminCollection.FindOne(ctx, bson.M{"email": admin.Email}).Decode(&foundAdmin)
	defer cancel()
	if err != nil {
		http.Error(w, "Email or Password is incorrect", http.StatusUnauthorized)
		return
	}
	passwordIsValid, msg := VerifyPassword(*admin.Password, *foundAdmin.Password)
	defer cancel()
	if passwordIsValid != true {
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	token, refreshToken, _ := utility.GenerateAllAdminTokens(*foundAdmin.Email, *foundAdmin.AdminName, foundAdmin.Admin_ID)
	utility.UpdateAllAdminTokens(token, refreshToken, foundAdmin.Admin_ID)

	w.Header().Set("Content-Type", "application/json")
	data := Data{
		Token:     token,
		AdminName: *foundAdmin.AdminName,
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
