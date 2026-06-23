package handler

import (
	"encoding/json"
	"folosy-backend/internal/service"
	"folosy-backend/internal/validation"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var registerRequest RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&registerRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validation.ValidateNewUserData(registerRequest.Email, registerRequest.Username, registerRequest.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userService.Register(registerRequest.Email, registerRequest.Username, registerRequest.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"email":    user.Email,
		"username": user.Username,
	})
}
