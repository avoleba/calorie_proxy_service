package cart

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"calorie-proxy/internal/auth"
	"calorie-proxy/internal/models"
	"calorie-proxy/internal/store"

	"database/sql"
)

// Handler — обработчики корзины (требуется авторизация)
type Handler struct {
	store *store.Store
}

func NewHandler(store *store.Store) *Handler {
	return &Handler{store: store}
}

// AddHandler POST /api/v1/cart — добавить продукт в корзину (в граммах)
func (h *Handler) AddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		sendErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req models.AddCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Grams <= 0 {
		sendErr(w, http.StatusBadRequest, "grams must be positive")
		return
	}
	item, err := h.store.AddCartItem(r.Context(), userID, req.Product, req.Grams)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, "failed to add to cart")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// ListHandler GET /api/v1/cart — список корзины с итогами по калориям/БЖУ
func (h *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		sendErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := h.store.GetCartByUserID(r.Context(), userID)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, "failed to load cart")
		return
	}
	var totalCal, totalPr, totalFat, totalCarb float64
	for i := range items {
		totalCal += items[i].TotalCalories
		totalPr += items[i].TotalProtein
		totalFat += items[i].TotalFat
		totalCarb += items[i].TotalCarbohydrates
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.CartListResponse{
		Items:              items,
		TotalCalories:      totalCal,
		TotalProtein:       totalPr,
		TotalFat:           totalFat,
		TotalCarbohydrates: totalCarb,
	})
}

// UpdateHandler PATCH /api/v1/cart?id=... — изменить граммовку позиции
func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		sendErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		sendErr(w, http.StatusBadRequest, "query id required")
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		sendErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req models.UpdateCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Grams <= 0 {
		sendErr(w, http.StatusBadRequest, "grams must be positive")
		return
	}
	item, err := h.store.UpdateCartItemGrams(r.Context(), id, userID, req.Grams)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, "update failed")
		return
	}
	if item == nil {
		sendErr(w, http.StatusNotFound, "cart item not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// DeleteHandler DELETE /api/v1/cart?id=... — удалить позицию из корзины
func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		sendErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		sendErr(w, http.StatusBadRequest, "query id required")
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		sendErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.store.DeleteCartItem(r.Context(), id, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendErr(w, http.StatusNotFound, "cart item not found")
			return
		}
		sendErr(w, http.StatusInternalServerError, "delete failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func sendErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{Error: http.StatusText(status), Status: status, Message: msg})
}
