package main

import (
	"fmt"
	"log"
	"net/http"

	"testify/internal/handlers"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:4200"},
		// AllowedOrigins: []string{"https://testify-admin.onrender.com", "https://testify-preview.onrender.com"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// User Routes
	r.Route("/users", func(r chi.Router) {
		r.Get("/", handlers.GetUsers)
		r.Post("/signup", handlers.SignUp)
		r.Post("/login", handlers.Login)
		r.Post("/delete", handlers.DeleteUser)
	})

	// Question routes
	r.Route("/questions", func(r chi.Router) {

		r.Post("/", handlers.CreateQuestion)
		r.Get("/", handlers.GetQuestions)
		r.Get("/{id}", handlers.GetQuestionByID)
		r.Put("/{id}", handlers.EditQuestion)
		r.Delete("/{id}", handlers.DeleteQuestion)
	})

	r.Route("/admins", func(r chi.Router) {
		r.Get("/", handlers.GetAdmins)
		r.Post("/create", handlers.CreateAdmin)
		r.Post("/login", handlers.AdminLogin)
		r.Post("/verify", handlers.VerifyAdminToken)
	})

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
