package api

import (
	"net/http"
	"strings"

	"github.com/kirilltahmazidi/aggregator/internal/api/handlers"
)

func NewRouter(h *handlers.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", h.Health)

	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.Login(w, r)
		} else {
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

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
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/offer") {
			h.OfferPrice(w, r)
			return
		}
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/confirm-price") {
			h.ConfirmPrice(w, r)
			return
		}
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/confirm-completion") {
			h.ConfirmCompletion(w, r)
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

	// Получение эксплуатанта по ID
	mux.HandleFunc("/operators/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetOperator(w, r)
			return
		}
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
	})

	// Заказчики
	mux.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.RegisterCustomer(w, r)
		} else {
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	// Получение заказчика по ID
	mux.HandleFunc("/customers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetCustomer(w, r)
			return
		}
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
	})

	// Отдача статичного фронта для удобства))))
	mux.Handle("/", http.FileServer(http.Dir("./frontend")))

	return enableCORS(mux)
}
