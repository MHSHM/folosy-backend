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
	"time"
)

type TransactionHandler struct {
	transactionService *service.TransactionService
}

func NewTransactionHandler(transactionService *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{transactionService: transactionService}
}

const maxTransactionBodyBytes = 2 << 10

type TransactionRequest struct {
	CategoryID  domain.NullString `json:"category_id"` // null/omitted = uncategorized
	AmountMinor int64             `json:"amount_minor"`
	Direction   domain.Direction  `json:"direction"`
	Merchant    string            `json:"merchant"`
	OccurredAt  time.Time         `json:"occurred_at"`
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	var req TransactionRequest

	r.Body = http.MaxBytesReader(w, r.Body, maxTransactionBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validation.ValidateTransaction(req.AmountMinor, req.Direction, req.Merchant, req.OccurredAt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	transaction, err := h.transactionService.Create(r.Context(), userID, req.CategoryID, req.AmountMinor, req.Direction, req.Merchant, req.OccurredAt)
	switch {
	case errors.Is(err, domain.ErrCategoryNotFound):
		// The body referenced a category that doesn't exist or isn't this user's.
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		log.Printf("create transaction: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	transactions, err := h.transactionService.List(r.Context(), userID)
	if err != nil {
		log.Printf("list transactions: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(transactions)
}

func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")

	transaction, err := h.transactionService.Get(r.Context(), id, userID)
	switch {
	case errors.Is(err, domain.ErrTransactionNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		log.Printf("get transaction: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")

	var req TransactionRequest

	r.Body = http.MaxBytesReader(w, r.Body, maxTransactionBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validation.ValidateTransaction(req.AmountMinor, req.Direction, req.Merchant, req.OccurredAt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	transaction, err := h.transactionService.Update(r.Context(), id, userID, req.CategoryID, req.AmountMinor, req.Direction, req.Merchant, req.OccurredAt)
	switch {
	case errors.Is(err, domain.ErrTransactionNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case errors.Is(err, domain.ErrCategoryNotFound):
		// The body referenced a category that doesn't exist or isn't this user's.
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		log.Printf("update transaction: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")

	err := h.transactionService.Delete(r.Context(), id, userID)
	switch {
	case errors.Is(err, domain.ErrTransactionNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		log.Printf("delete transaction: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TransactionHandler) TopExpenses(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	from, err := time.Parse(time.RFC3339, r.URL.Query().Get("from"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	to, err := time.Parse(time.RFC3339, r.URL.Query().Get("to"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !from.Before(to) {
		http.Error(w, "'from' must be before 'to'", http.StatusBadRequest)
		return
	}

	topExpenses, err := h.transactionService.TopExpenses(r.Context(), userID, from, to)
	if err != nil {
		log.Printf("top expenses: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(topExpenses)
}
