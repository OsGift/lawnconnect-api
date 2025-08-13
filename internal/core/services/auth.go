package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"lawnconnect-api/internal/core/apperror"
	"lawnconnect-api/internal/core/domain"
	"lawnconnect-api/internal/infrastructure/database/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jwtKey = []byte("your_very_secure_secret_key") // Use a secret key from your .env file

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
}

type authService struct {
	userRepo repositories.UserRepository
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(userRepo repositories.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
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
		IsVerified: false,
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
	// Find the user by email
	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, "", apperror.InvalidLoginCredentials{}
	}

	// Compare the password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", apperror.InvalidLoginCredentials{}
	}

	// Generate a JWT token
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
