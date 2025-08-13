package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	httpresponse "lawnconnect-api/internal/api/http"
	"lawnconnect-api/internal/core/apperror"
	"lawnconnect-api/internal/core/domain"
	"lawnconnect-api/internal/core/services"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	AuthService services.AuthService
}

func NewAuthHandler(authSrv services.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authSrv}
}

// RegisterUser handles user registration.
func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.AuthService.Register(r.Context(), reqBody.Name, reqBody.Email, reqBody.Password, reqBody.Role)
	if err != nil {
		if _, ok := err.(apperror.DuplicateError); ok {
			httpresponse.JSONError(w, http.StatusConflict, err.Error())
			return
		}
		httpresponse.JSONError(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	log.Printf("User registration for %s successful", user.Email)
	httpresponse.JSONSuccess(w, http.StatusCreated, "User registered successfully", user)
}

// LoginUser handles user login.
func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, token, err := h.AuthService.Login(r.Context(), credentials.Email, credentials.Password)
	if err != nil {
		if _, ok := err.(apperror.InvalidLoginCredentials); ok {
			httpresponse.JSONError(w, http.StatusUnauthorized, err.Error())
			return
		}
		httpresponse.JSONError(w, http.StatusInternalServerError, "Failed to log in")
		return
	}

	log.Printf("User login for %s successful", user.Email)
	response := struct {
		User  *domain.User `json:"user"`
		Token string       `json:"token"`
	}{
		User:  user,
		Token: token,
	}

	httpresponse.JSONSuccess(w, http.StatusOK, "Login successful", response)
}

// Routes mounts the authentication routes to a chi router.
func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.RegisterUser)
	r.Post("/login", h.LoginUser)
	return r
}
