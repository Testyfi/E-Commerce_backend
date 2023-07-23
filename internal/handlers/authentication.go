package handlers

import (
	"net/http"
	"strings"
	utility "testify/internal/utility"
)

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenString = strings.Split(tokenString, " ")[1]

		// Validate the JWT token
		claims, errMsg := utility.ValidateToken(tokenString)
		if errMsg != "" {
			claims, errMsg := utility.ValidateAdminToken(tokenString)
			if errMsg != "" {
				http.Error(w, errMsg, http.StatusUnauthorized)
				return
			}
			println(claims)
			next.ServeHTTP(w, r)
			return
		}
		println(claims)
		next.ServeHTTP(w, r)
	})
}

func AdminAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenString = strings.Split(tokenString, " ")[1]
		// Validate Admin Token
		claims, errMsg := utility.ValidateAdminToken(tokenString)
		if errMsg != "" {
			http.Error(w, errMsg, http.StatusUnauthorized)
			return
		}
		println(claims)
		next.ServeHTTP(w, r)
	})
}
