package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"lawnconnect-api/internal/core/apperror"
	"lawnconnect-api/internal/core/domain"
	"lawnconnect-api/internal/infrastructure/database/repositories"
	infrastructureServices "lawnconnect-api/internal/infrastructure/services"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// Claims represents the JWT claims.
type Claims struct {
	UserID primitive.ObjectID `json:"userId"`
	Role   string             `json:"role"`
	jwt.RegisteredClaims
}

// AuthService defines the business logic for authentication.
type AuthService interface {
	Register(ctx context.Context, name, email, password, role string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, string, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

type authService struct {
	userRepo     repositories.UserRepository
	emailService infrastructureServices.EmailService
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(userRepo repositories.UserRepository, emailService infrastructureServices.EmailService) AuthService {
	return &authService{userRepo: userRepo, emailService: emailService}
}

// Register handles user registration logic.
func (s *authService) Register(ctx context.Context, name, email, password, role string) (*domain.User, error) {
	_, err := s.userRepo.FindUserByEmail(ctx, email)
	if err == nil {
		return nil, apperror.DuplicateError{Resource: "User with this email"}
	}
	if _, ok := err.(apperror.NotFound); !ok {
		return nil, fmt.Errorf("error checking for existing user: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		ID:        primitive.NewObjectID(),
		Name:      name,
		Email:     email,
		Password:  string(hashedPassword),
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to save user to database: %w", err)
	}

	return user, nil
}

// Login handles user login and JWT token generation.
func (s *authService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, "", apperror.InvalidLoginCredentials{}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", apperror.InvalidLoginCredentials{}
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return user, tokenString, nil
}

// ForgotPassword handles the logic for a user requesting a password reset.
func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		if _, ok := err.(apperror.NotFound); ok {
			// Fail silently to prevent email enumeration attacks.
			log.Printf("Password reset request for non-existent email: %s", email)
			return nil
		}
		return fmt.Errorf("error finding user: %w", err)
	}

	resetToken, err := generateRandomToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	resetTokenExpiresAt := time.Now().Add(1 * time.Hour)
	update := bson.M{
		"$set": bson.M{
			"resetToken":          resetToken,
			"resetTokenExpiresAt": resetTokenExpiresAt,
			"updatedAt":           time.Now(),
		},
	}
	err = s.userRepo.UpdateUser(ctx, user.ID, update)
	if err != nil {
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	resetURL := fmt.Sprintf("%s?token=%s", os.Getenv("LOGIN_URL"), resetToken)
	templateData := map[string]interface{}{
		"Name":     user.Name,
		"ResetURL": resetURL,
	}

	log.Printf("Password reset email sent to %s with reset link: %s", user.Email, resetURL)
	err = s.emailService.SendEmail(ctx, user.Email, "Password Reset Request", "password-reset.html", templateData)
	if err != nil {
		log.Printf("Failed to send password reset email to %s: %v", user.Email, err)
	}

	return nil
}

// ResetPassword handles the logic for a user resetting their password with a valid token.
func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	user, err := s.userRepo.FindUserByResetToken(ctx, token)
	if err != nil {
		return apperror.CustomError{Message: "Invalid or expired token"}
	}

	if time.Now().After(user.ResetTokenExpiresAt) {
		return apperror.CustomError{Message: "Invalid or expired token"}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"password":            string(hashedPassword),
			"resetToken":          "", // Clear the token after use
			"resetTokenExpiresAt": time.Time{},
			"updatedAt":           time.Now(),
		},
	}

	err = s.userRepo.UpdateUser(ctx, user.ID, update)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// generateRandomToken creates a cryptographically secure random string.
func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
