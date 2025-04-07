package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	models "testify/internal/models"

	"go.mongodb.org/mongo-driver/bson"
)


func ServerCallBack(w http.ResponseWriter, r *http.Request) {
	var Response struct {
		Success bool   `json:"success"`
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    struct {
			MerchantID            string `json:"merchantId"`
			MerchantTransactionID string `json:"merchantTransactionId"`
			TransactionID         string `json:"transactionId"`
			Amount                int    `json:"amount"`
			State                 string `json:"state"`
			ResponseCode          string `json:"responseCode"`
			PaymentInstrument     struct {
				Type string `json:"type"`
				Utr  string `json:"utr"`
			} `json:"paymentInstrument"`
		} `json:"data"`
	}
	
	var foundUser models.User
	if err := json.NewDecoder(r.Body).Decode(&Response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if(Response.Success){
	err := userCollection.FindOne(context.Background(), bson.M{"phone": Response.Data.MerchantTransactionID[:10]}).Decode(&foundUser)
	if err != nil {
		http.Error(w, "Invalid Phone", http.StatusUnauthorized)
		return
	}
	foundUser.PurchaseDate=time.Now()
	foundUser.Purchased=true
	if(Response.Data.Amount/100<=199){
	foundUser.PurchasePlan=0
	}
	if(Response.Data.Amount/100>199){
		foundUser.PurchasePlan=1
		}
		
			filter := bson.M{"user_id": foundUser.User_id}
			update := bson.M{"$set": bson.M{
				
				"purchasedate":foundUser.PurchaseDate,
				"purchased":true,
				"purchaseplane":foundUser.PurchasePlan,
			}}

			 userCollection.UpdateOne(context.Background(), filter, update)
			

 

}
	

	
}