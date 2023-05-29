package handlers

import (
	"context"
	"net/http"

	utility "testify/internal/utility"
)

func Authentication() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientToken := r.Header.Get("token")
			if clientToken == "" {
				http.Error(w, "No Authorization header provided", http.StatusInternalServerError)
				return
			}

			claims, err := utility.ValidateToken(clientToken)
			if err != "" {
				http.Error(w, err, http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), "email", claims.Email)
			ctx = context.WithValue(ctx, "first_name", claims.First_name)
			ctx = context.WithValue(ctx, "last_name", claims.Last_name)
			ctx = context.WithValue(ctx, "uid", claims.Uid)

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
