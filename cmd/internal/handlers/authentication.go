package handlers

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
	"testify/internal/models"
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
		authDetails, errMsg := utility.ValidateToken(tokenString)
		if errMsg != "" {
			claims, errMsg := utility.ValidateAdminToken(tokenString)
			if errMsg != "" {
				println(claims)
				http.Error(w, errMsg, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		err := userCollection.FindOne(context.Background(), bson.M{"user_id": authDetails.Uid}).Decode(&user)
		if err != nil {
			// Ignore at the moment
			// fmt.Println(err)
			// http2.RespondError(w, http.StatusUnauthorized, err.Error(), err)
			// return
		}

		ctx := context.WithValue(r.Context(), models.ContextUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
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
