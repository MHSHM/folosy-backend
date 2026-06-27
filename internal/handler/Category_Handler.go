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

type CategoryHandler struct {
	categoryService *service.CategoryService
}

func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// maxAuthBodyBytes caps the request body for auth endpoints
const maxCategoryBodyBytes = 2 << 10

type CategoryRequest struct {
	Name             string `json:"name"`
	Icon             string `json:"icon"`
	BudgetLimitMinor int64  `json:"budget_limit_minor"`
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	var categoryRequest CategoryRequest

	r.Body = http.MaxBytesReader(w, r.Body, maxCategoryBodyBytes)
	err := json.NewDecoder(r.Body).Decode(&categoryRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validation.ValidateCategory(categoryRequest.Name, categoryRequest.BudgetLimitMinor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category, err := h.categoryService.Create(r.Context(), userID, categoryRequest.Name, categoryRequest.Icon, categoryRequest.BudgetLimitMinor)
	switch {
	case errors.Is(err, domain.ErrCategoryNameExists):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case err != nil:
		log.Printf("create category: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	categories, err := h.categoryService.List(r.Context(), userID)
	if err != nil {
		log.Printf("list categories: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(categories)
}

func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")

	category, err := h.categoryService.Get(r.Context(), id, userID)
	switch {
	case errors.Is(err, domain.ErrCategoryNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		log.Printf("get category: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")

	var categoryRequest CategoryRequest

	r.Body = http.MaxBytesReader(w, r.Body, maxCategoryBodyBytes)
	err := json.NewDecoder(r.Body).Decode(&categoryRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validation.ValidateCategory(categoryRequest.Name, categoryRequest.BudgetLimitMinor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category, err := h.categoryService.Update(r.Context(), id, userID, categoryRequest.Name, categoryRequest.Icon, categoryRequest.BudgetLimitMinor)
	switch {
	case errors.Is(err, domain.ErrCategoryNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case errors.Is(err, domain.ErrCategoryNameExists):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case err != nil:
		log.Printf("update category: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")

	err := h.categoryService.Delete(r.Context(), id, userID)
	switch {
	case errors.Is(err, domain.ErrCategoryNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		log.Printf("delete category: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
