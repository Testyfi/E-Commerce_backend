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
	cursor,err:=qpaperCollection.Find(ctx,bson.M{})

	if err !=nil{

		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
		return
	}
	var questions []bson.M
	if err =cursor.All(ctx,&questions);err!=nil{
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid tag", err)
		return
	}

	httpClient.RespondSuccess(w, questions)
}
