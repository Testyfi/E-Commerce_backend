package handlers

import (
	"net/http"
	utility "testify/internal/utility"
)

func VerifyAdminToken(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
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
