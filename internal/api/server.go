package api

import "net/http"

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", h.Health)

	// Заказы: /orders (без ID) и /orders/{id} (с ID) — разные patterns
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
