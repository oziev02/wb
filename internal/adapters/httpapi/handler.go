package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/oziev02/wb/internal/usecase"
)

type Handler struct{ uc *usecase.OrderService }

func NewHandler(uc *usecase.OrderService) *Handler { return &Handler{uc: uc} }

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/order/", h.getOrder)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil { /* ignore */
		}
	})
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/order/")
	if id == "" {
		http.Error(w, "order id required", http.StatusBadRequest)
		return
	}
	o, ok, err := h.uc.Get(id)
	if err != nil && !ok {
		var status = http.StatusInternalServerError
		if errors.Is(err, contextCanceled(r.Context())) {
			status = http.StatusRequestTimeout
		}
		http.Error(w, "server error", status)
		return
	}
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(o); err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
	}
}

// утилита, чтобы не тянуть context сюда.
func contextCanceled(_ interface{}) error { return nil }
