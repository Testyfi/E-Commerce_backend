package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	database "testify/database"
	models "testify/internal/models"
	utility "testify/internal/utility"
	httpClient "testify/internal/utility/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Data struct {
	Token     string `json:"token"`
	AdminName string `json:"adminName"`
}

var adminCollection *mongo.Collection = database.OpenCollection(database.Client, "admins")

func GetAdmins(w http.ResponseWriter, r *http.Request) {
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

	cur, err := adminCollection.Find(context.Background(), bson.M{}, findOptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())

	admins := []models.Admin{}
	for cur.Next(context.Background()) {
		var admin models.Admin
		err := cur.Decode(&admin)
		if err != nil {
			http.Error(w, "Interal Server Error"+err.Error(), http.StatusInternalServerError)
			return
		}
		admins = append(admins, admin)
	}
	if err := cur.Err(); err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}

	totalAdmins, err := adminCollection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}

	// Respond with the questions and total number of questions
	response := map[string]interface{}{
		"admins":      admins,
		"totalAdmins": totalAdmins,
	}
	// Set response headers and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreateAdmin(w http.ResponseWriter, r *http.Request) {
	var admin models.Admin
    
	if err := json.NewDecoder(r.Body).Decode(&admin); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		print(err.Error())
		return
	}
    
	// Password Hashing
	password := HashPassword(*admin.Password)
	admin.Password = &password

	// Checking if admin already exists
	alreadyExists, err := adminCollection.CountDocuments(context.Background(), bson.M{"email": *admin.Email})

	if err != nil {
		print(err.Error())
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
	insertResult, err := adminCollection.InsertOne(context.Background(), admin)
	if err != nil {
		http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insertResult)
	w.WriteHeader(http.StatusOK)
	return
}

func AdminLogin(w http.ResponseWriter, r *http.Request) {

	var foundAdmin models.Admin
	var admin models.Admin
	if err := json.NewDecoder(r.Body).Decode(&admin); err != nil {
		//println("yes")
		http.Error(w, "Interal Server Error "+err.Error(), http.StatusInternalServerError)
		return
	}
    //println(*admin.Email)
	err := adminCollection.FindOne(context.Background(), bson.M{"email": admin.Email}).Decode(&foundAdmin)
    //println(err.Error())
	if err != nil {
		http.Error(w, "Email or Password is incorrect", http.StatusUnauthorized)
		return
	}
	passwordIsValid, msg := VerifyPassword(*admin.Password, *foundAdmin.Password)

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
func CreateTest(w http.ResponseWriter, r *http.Request){

	var PaperDetails struct {
		Name       string `json:"Name"`
		Start      string `json:"Start"`
		StartAt    string `json:"StartAt"`
		Difficulty string `json:"Difficulty"`
		Topics     string `json:"Topics"`
		Duration   string `json:"Duration"`
		Prize      string `json:"Prize"`
	}
	var Insert struct {
		Name       string `json:"Name"`
		Start      string `json:"Start"`
		StartAt    string `json:"StartAt"`
		StartDate  time.Time `json:"StartDate"`
		Difficulty string `json:"Difficulty"`
		Topics     string `json:"Topics"`
		Duration   string `json:"Duration"`
		Prize      string `json:"Prize"`
	}
	
	
		err := json.NewDecoder(r.Body).Decode(&PaperDetails)
	  
		if err != nil {
			httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid testname", err)
			return
		}
		Insert.Name=PaperDetails.Name
		Insert.Start=PaperDetails.Start
		Insert.StartAt=PaperDetails.StartAt
		Insert.Difficulty=PaperDetails.Difficulty
		Insert.Topics=PaperDetails.Topics
		Insert.Duration=PaperDetails.Duration
		Insert.Prize=PaperDetails.Prize
		indianTimeZone, err := time.LoadLocation("Asia/Kolkata")
		if err != nil {
			fmt.Println("Error loading Indian time zone:", err)
			
		}
		
		s:=strings.Split(Insert.Start, "/")
		//fmt.Println(s)
		Insert.StartDate= time.Date(StringtoInt(s[0]), time.Month(StringtoInt(s[1])), StringtoInt(s[2]), StringtoInt(s[3]), StringtoInt(s[4]), StringtoInt(s[5]), 0,indianTimeZone)
		fmt.Println(Insert.StartDate)
	testdetailsCollection.InsertOne(context.TODO(),Insert)
	httpClient.RespondSuccess(w,"Success")
}
func GetAllTestDetails(w http.ResponseWriter, r *http.Request){
	//indianTimeZone, err := time.LoadLocation("Asia/Kolkata")
	type PaperDetails struct {
		Name       string `json:"Name"`
		Start      string `json:"Start"`
		StartAt    string `json:"StartAt"`
		StartDate  time.Time `json:"StartDate"`
		Difficulty string `json:"Difficulty"`
		Topics     string `json:"Topics"`
		Duration   string `json:"Duration"`
		Prize      string `json:"Prize"`
	}
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=testdetailsCollection.Find(ctx,bson.M{})
	defer cursor.Close(ctx)
	if err !=nil{

		httpClient.RespondError(w, http.StatusBadRequest, "Some Error", err)
		return
	}
	var papers []PaperDetails
	if err =cursor.All(ctx,&papers);err!=nil{
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
		return
	}
    
	 //fmt.Println(papers)
	httpClient.RespondSuccess(w, papers)



}

