package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"testify/internal/handlers"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{os.Getenv("STUDENT_FRONTEND_URL"), os.Getenv("ADMIN_FRONTEND_URL"), "http://localhost:4200"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// User Routes
	r.Post("/userlogin", handlers.Login)
	r.Post("/usersignup", handlers.SignUp)
	r.Post("/userverify", handlers.UserVerification)
	r.Post("/forgotpass", handlers.ForgotPassword)
	r.Get("/reset", handlers.ResetPassword)
	r.Route("/users", func(r chi.Router) {
		r.Use(handlers.AuthenticationMiddleware)
		r.Post("/delete", handlers.DeleteUser)
		r.Get("/stats/{user_id}/{paper_id}", handlers.GetPaperStats)
		r.Post("/submit/user/{user_id}/paper/{paper_id}", handlers.SubmitQPaper)
		r.Put("/{user_id}/profile", handlers.UpdateProfilePic)
		r.Get("/{user_id}/purchase", handlers.PurchaseCourse)
		r.Put("/{user_id}/passwordchange", handlers.ChangePassword)
		r.Post("/{user_id}/createTest", handlers.CreateQPaper)
	})

	r.Get("/image/{image}", handlers.ServeImage)

	// Question routes
	r.Route("/questions", func(r chi.Router) {
		r.Use(handlers.AuthenticationMiddleware)
		r.Post("/", handlers.CreateQuestion)
		r.Get("/", handlers.GetQuestions)
		r.Get("/id/{id}", handlers.GetQuestionByID)
		r.Put("/id/{id}", handlers.EditQuestion)
		r.Delete("/{id}", handlers.DeleteQuestion)
		r.Post("/delete", handlers.DeleteMany)
		r.Post("/upload", handlers.UploadCSV)
	})

	r.Post("/adminlogin", handlers.AdminLogin)
	r.Route("/admins", func(r chi.Router) {
		r.Use(handlers.AdminAuthenticationMiddleware)
		r.Get("/", handlers.GetAdmins)
		r.Get("/users", handlers.GetUsers)
		r.Post("/create", handlers.CreateAdmin)
		r.Post("/testinfo", handlers.GetAllTestDetails)
		r.Post("/createtest", handlers.CreateTest)
		r.Post("/deletetestinfo", handlers.DeleteTestInfo)
		r.Get("/verify", handlers.VerifyAdminToken)
	})

	r.Route("/payment", func(r chi.Router) {
		r.Use(handlers.AuthenticationMiddleware)
		r.Post("/phonepay/request", handlers.GetPaymentRequestUrl)
	})
	r.Route("/rankbooster", func(r chi.Router) {
		r.Use(handlers.AuthenticationMiddleware)
		r.Post("/pasttest", handlers.GetPastTest)
		r.Post("/livetest", handlers.GetLiveTestQuestion)
		r.Post("/livetest/delete/userdata", handlers.DeleteLiveTestAllUserData)
		r.Post("/livetest/response", handlers.LiveTestResponse)
	})
	

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
