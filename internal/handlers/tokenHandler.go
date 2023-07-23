package handlers

import (
	"net/http"
	"strings"
	utility "testify/internal/utility"
)

func VerifyAdminToken(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenString = strings.Split(tokenString, " ")[1]
	claims, errMsg := utility.ValidateAdminToken(tokenString)
	if errMsg != "" {
		http.Error(w, errMsg, http.StatusUnauthorized)
		return
	}
	email := claims.Email
	if email == "" {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
