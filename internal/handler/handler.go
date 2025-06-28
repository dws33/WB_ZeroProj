package handler

import (
	"context"
	"encoding/json"
	"github.com/dws33/WB_ZeroProj/internal/model"
	"log"
	"net/http"
	//"order-service/internal/storage"
)

type storage interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrder(ctx context.Context, uid string) (*model.Order, error)
}

// Handler хранит зависимости: БД и кэш.
type Handler struct {
	storage storage
}

// New создает новый Handler.
func New(storage storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

// GetOrder — HTTP-обработчик GET /order/{order_uid}
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderUID := r.PathValue("order_uid")
	if orderUID == "" {
		http.Error(w, "order_uid is required", http.StatusBadRequest)
		return
	}

	order, err := h.storage.GetOrder(r.Context(), orderUID)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(order); err != nil {
		log.Println("failed to encode response:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
}
