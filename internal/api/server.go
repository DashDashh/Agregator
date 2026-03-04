package api

import (
	"net/http"
	"strings"
)

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", h.Health)

	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListOrders(w, r)
		case http.MethodPost:
			h.CreateOrder(w, r)
		default:
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/status") {
			h.UpdateOrderStatus(w, r)
			return
		}
		if r.Method == http.MethodGet {
			h.GetOrder(w, r)
		} else {
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	// Эксплуатанты
	mux.HandleFunc("/operators", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.RegisterOperator(w, r)
		} else {
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	// Заказчики
	mux.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.RegisterCustomer(w, r)
		} else {
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
