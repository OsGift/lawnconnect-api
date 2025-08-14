package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	httpresponse "lawnconnect-api/internal/api/http"
	"lawnconnect-api/internal/core/apperror"
	"lawnconnect-api/internal/core/services"
	infrastructureServices "lawnconnect-api/internal/infrastructure/services"

	"github.com/go-chi/chi/v5"
)

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	AuthService   services.AuthService
	UploadService infrastructureServices.UploadService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authSrv services.AuthService) *AuthHandler {
	return &AuthHandler{
		AuthService: authSrv,
		// Note: The UploadService is not yet included here as it's not a dependency
		// for the core auth methods, but it's part of the struct for future use.
	}
}

// Register handles user registration.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if reqBody.Role != "customer" && reqBody.Role != "mower" {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid role specified")
		return
	}

	user, err := h.AuthService.Register(r.Context(), reqBody.Name, reqBody.Email, reqBody.Password, reqBody.Role)
	if err != nil {
		if _, ok := err.(apperror.DuplicateError); ok {
			httpresponse.JSONError(w, http.StatusConflict, err.Error())
			return
		}
		httpresponse.JSONError(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	httpresponse.JSONSuccess(w, http.StatusCreated, "User registered successfully", user)
}

// Login handles user login and JWT token generation.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, token, err := h.AuthService.Login(r.Context(), reqBody.Email, reqBody.Password)
	if err != nil {
		if _, ok := err.(apperror.InvalidLoginCredentials); ok {
			httpresponse.JSONError(w, http.StatusUnauthorized, "Invalid login credentials")
			return
		}
		httpresponse.JSONError(w, http.StatusInternalServerError, "Failed to log in")
		return
	}

	response := map[string]interface{}{
		"user":  user,
		"token": token,
	}

	httpresponse.JSONSuccess(w, http.StatusOK, "Login successful", response)
}

// ForgotPassword handles the forgot password request.
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.AuthService.ForgotPassword(r.Context(), reqBody.Email)
	if err != nil {
		log.Printf("ForgotPassword failed for %s: %v", reqBody.Email, err)
		// We return a success message even if the user doesn't exist to prevent
		// email enumeration attacks.
	}

	httpresponse.JSONSuccess(w, http.StatusOK, "If a user with that email exists, a password reset link has been sent.", nil)
}

// ResetPassword handles the password reset with a token.
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.AuthService.ResetPassword(r.Context(), reqBody.Token, reqBody.NewPassword)
	if err != nil {
		httpresponse.JSONError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	httpresponse.JSONSuccess(w, http.StatusOK, "Password reset successfully", nil)
}

// Routes mounts the authentication routes.
func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/forgot-password", h.ForgotPassword)
	r.Post("/reset-password", h.ResetPassword)
	return r
}
