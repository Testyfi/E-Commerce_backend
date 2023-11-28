package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	httpClient "testify/internal/utility/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func GetPastTest(w http.ResponseWriter, r *http.Request) {

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
