package handler

import (
	"encoding/json"
	"errors"
	"folosy-backend/internal/auth"
	"folosy-backend/internal/domain"
	"folosy-backend/internal/service"
	"folosy-backend/internal/validation"
	"log"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// maxAuthBodyBytes caps the request body for auth endpoints makes a "huge body" memory-exhaustion attempt harmless.
const maxAuthBodyBytes = 2 << 10 // 4 KB

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var registerRequest RegisterRequest

	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)
	err := json.NewDecoder(r.Body).Decode(&registerRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validation.ValidateRegister(registerRequest.Email, registerRequest.Username, registerRequest.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userService.Register(r.Context(), registerRequest.Email, registerRequest.Username, registerRequest.Password)
	switch {
	case errors.Is(err, domain.ErrEmailExists):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case err != nil:
		log.Printf("register user: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"email":    user.Email,
		"username": user.Username,
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest LoginRequest

	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validation.ValidateLogin(loginRequest.Email, loginRequest.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.userService.Login(r.Context(), loginRequest.Email, loginRequest.Password)
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	case err != nil:
		log.Printf("login user: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}

// Me returns the authenticated user's own profile. It runs only behind
// RequireAuth, so the user ID is already verified and waiting in the context.
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	case err != nil:
		log.Printf("get me: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
	})
}

type GoogleLoginRequest struct {
	IDToken string `json:"id_token"`
}

// GoogleLogin exchanges a Google ID token for our own access + refresh pair.
func (h *UserHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var googleLoginRequest GoogleLoginRequest

	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)
	err := json.NewDecoder(r.Body).Decode(&googleLoginRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validation.ValidateGoogle(googleLoginRequest.IDToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.userService.GoogleLogin(r.Context(), googleLoginRequest.IDToken)
	switch {
	case errors.Is(err, domain.ErrInvalidGoogleToken):
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	case err != nil:
		log.Printf("google login: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var refreshRequest RefreshRequest

	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)
	err := json.NewDecoder(r.Body).Decode(&refreshRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validation.ValidateRefresh(refreshRequest.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.userService.Refresh(r.Context(), refreshRequest.RefreshToken)
	switch {
	case errors.Is(err, domain.ErrInvalidRefreshToken):
		// One generic 401 for unknown / expired / revoked-reuse alike.
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	case err != nil:
		log.Printf("refresh token: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}
