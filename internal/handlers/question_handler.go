package handlers

// import (
// 	"encoding/json"
// 	"net/http"

// 	db "testify/database"
// )

// func GetQuestionsHandler(w http.ResponseWriter, r *http.Request) {
// 	questions, err := db.GetQuestions()
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(questions)
// }
