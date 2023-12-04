package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	httpClient "testify/internal/utility/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func GetLiveTest(w http.ResponseWriter, r *http.Request) {

	var t struct {
		Tag string `json:"tag"`
	}

	err := json.NewDecoder(r.Body).Decode(&t)

	if err != nil {
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
		return
	}
    ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=questionCollection.Find(ctx,bson.M{"subject_tags":t.Tag})
    defer cursor.Close(ctx)
	if err !=nil{

		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
		return
	}
	var questions []bson.M
	if err =cursor.All(ctx,&questions);err!=nil{
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
		return
	}
    // fmt.Println(questions)
	 
	httpClient.RespondSuccess(w, questions)
}
func GetRank(w http.ResponseWriter, r *http.Request){

	var t struct {
		Testname string `json:"testname"`
		User string `json:"user"`
		Number int `json:"number"`
	}


	err := json.NewDecoder(r.Body).Decode(&t)

	if err != nil {
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid test or user", err)
		return
	}
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err:=testpaperCollection.Find(ctx,bson.M{"number":bson.M{"$gt":t.Number}})
    defer cursor.Close(ctx)
	if err !=nil{
                
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid test or user", err)
		return

		
	}
	var users []bson.M
	if err =cursor.All(ctx,&users);err!=nil{
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
		return
	}
	//cursor,err=testpaperCollection.Find(ctx,bson.M{"test_name":t.Testname,"User":t.User})
	
	httpClient.RespondSuccess(w, len(users)+1)

}
func InsertTestData(w http.ResponseWriter, r *http.Request ){

	var t struct {
		Testname string `json:"testname"`
		User string `json:"user"`
		Number int `json:"number"`
	}
	err := json.NewDecoder(r.Body).Decode(&t)

	if err != nil {
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid test or user", err)
		return
	}
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second) 
		testpaperCollection.InsertOne(ctx,bson.M{"test_name":t.Testname,"user":t.User,"number":t.Number})

		httpClient.RespondSuccess(w, "success")
	
	

}
func UpdateTestData(w http.ResponseWriter, r *http.Request ){

	var t struct {
		Testname string `json:"testname"`
		User string `json:"user"`
		Number int `json:"number"`
	}
	err := json.NewDecoder(r.Body).Decode(&t)

	if err != nil {
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid test or user", err)
		return
	}
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second) 
		testpaperCollection.UpdateOne(ctx,bson.M{"test_name":t.Testname,"user":t.User},bson.M{"$set":bson.M{"number":t.Number}})

		httpClient.RespondSuccess(w, "success")
	
	

}
