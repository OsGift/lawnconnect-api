package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"lawnconnect-api/internal/api/handlers"
	coreServices "lawnconnect-api/internal/core/services"
	"lawnconnect-api/internal/infrastructure/database"
	"lawnconnect-api/internal/infrastructure/database/repositories"
	infrastructureServices "lawnconnect-api/internal/infrastructure/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")
	if mongoURI == "" || dbName == "" {
		log.Fatal("MONGO_URI and MONGO_DB_NAME must be set in .env")
	}
	mongoClient, err := database.NewMongoClient(ctx, mongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())
	db := mongoClient.Database(dbName)

	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		log.Printf("Could not initialize Cloudinary, uploads will not work: %v", err)
	}
	uploadService := infrastructureServices.NewUploadService(cld)
	_ = uploadService

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("FROM_EMAIL")
	templatesPath := os.Getenv("TEMPLATES_PATH")
	loginURL := os.Getenv("LOGIN_URL")
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		log.Fatalf("Invalid SMTP_PORT: %v", err)
	}
	emailService := infrastructureServices.NewEmailService(smtpHost, smtpPort, smtpUser, smtpPass, fromEmail, templatesPath, loginURL)
	_ = emailService

	userRepo := repositories.NewUserRepository(db)
	bookingRepo := repositories.NewBookingRepository(db)

	authService := coreServices.NewAuthService(userRepo, emailService)
	bookingService := coreServices.NewBookingService(bookingRepo)

	authHandler := handlers.NewAuthHandler(authService)
	bookingHandler := handlers.NewBookingHandler(bookingService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "web"))
	r.Handle("/*", http.FileServer(filesDir))

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes for authentication
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)
		})

		// Protected routes for customers
		r.Group(func(r chi.Router) {
			r.Use(handlers.AuthMiddleware)
			r.Use(handlers.RoleMiddleware("customer"))

			// Customer-specific booking routes
			r.Post("/bookings", bookingHandler.CreateBooking)
			r.Get("/bookings", bookingHandler.ListBookings)
			r.Get("/bookings/{bookingID}", bookingHandler.GetBookingByID)
			r.Put("/bookings/{bookingID}/cancel", bookingHandler.CancelBooking)
		})

		// Protected routes for mowers
		r.Group(func(r chi.Router) {
			r.Use(handlers.AuthMiddleware)
			r.Use(handlers.RoleMiddleware("mower"))

			// Mower-specific booking routes
			r.Get("/bookings", bookingHandler.ListBookings)
			r.Get("/bookings/pending", bookingHandler.ListPendingBookings)
			r.Get("/bookings/{bookingID}", bookingHandler.GetBookingByID)
			r.Put("/bookings/{bookingID}/accept", bookingHandler.AcceptBooking)
			r.Put("/bookings/{bookingID}/complete", bookingHandler.CompleteBooking)
			r.Put("/bookings/{bookingID}/reject", bookingHandler.RejectBooking)
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
