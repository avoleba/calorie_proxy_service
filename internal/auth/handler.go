package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"calorie-proxy/internal/models"
	"calorie-proxy/internal/store"
)

type Handler struct {
	store *store.Store
	secret string
}

func NewHandler(store *store.Store, jwtSecret string) *Handler {
	return &Handler{store: store, secret: jwtSecret}
}

// RegisterHandler POST /api/v1/auth/register
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Email == "" || req.Password == "" {
		sendErr(w, http.StatusBadRequest, "email and password required")
		return
	}
	if len(req.Password) < 6 {
		sendErr(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}
	user, err := h.store.CreateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		if isUniqueErr(err) {
			sendErr(w, http.StatusConflict, "email already registered")
			return
		}
		sendErr(w, http.StatusInternalServerError, "registration failed")
		return
	}
	token, exp, err := NewToken(user.ID, user.Email, h.secret)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, "token creation failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AuthResponse{
		Token:     token,
		ExpiresAt: exp,
		User:      *user,
	})
}

// LoginHandler POST /api/v1/auth/login
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErr(w, http.StatusBadRequest, "invalid body")
		return
	}

	if req.Email == "" || req.Password == "" {
		sendErr(w, http.StatusBadRequest, "email and password required")
		return
	}
	user, err := h.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		sendErr(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if !store.CheckPassword(user.PasswordHash, req.Password) {
		sendErr(w, http.StatusUnauthorized, "invalid password")
		return
	}
	token, exp, err := NewToken(user.ID, user.Email, h.secret)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, "token creation failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AuthResponse{
		Token:     token,
		ExpiresAt: exp,
		User:      *user,
	})
}

func sendErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{Error: http.StatusText(status), Status: status, Message: msg})
}

func isUniqueErr(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "UNIQUE") || strings.Contains(s, "unique") || strings.Contains(s, "Duplicate")
}
