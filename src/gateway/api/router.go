package api

import (
	"net/http"
	"strings"

	contractsapi "github.com/kirilltahmazidi/aggregator/src/contracts_component/httpapi"
	ordersapi "github.com/kirilltahmazidi/aggregator/src/orders_component/httpapi"
	registryapi "github.com/kirilltahmazidi/aggregator/src/registry_component/httpapi"
	"github.com/kirilltahmazidi/aggregator/src/shared/httpx"
)

type Handlers struct {
	Registry  *registryapi.Handler
	Orders    *ordersapi.Handler
	Contracts *contractsapi.Handler
}

func NewRouter(h Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		httpx.Respond(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.Registry.Login(w, r)
		} else {
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.Orders.ListOrders(w, r)
		case http.MethodPost:
			h.Orders.CreateOrder(w, r)
		default:
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/offer") {
			h.Contracts.OfferPrice(w, r)
			return
		}
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/confirm-price") {
			h.Contracts.ConfirmPrice(w, r)
			return
		}
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/confirm-completion") {
			h.Contracts.ConfirmCompletion(w, r)
			return
		}
		if r.Method == http.MethodGet {
			h.Orders.GetOrder(w, r)
		} else {
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	// Эксплуатанты
	mux.HandleFunc("/operators", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.Registry.RegisterOperator(w, r)
		} else {
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	// Получение эксплуатанта по ID
	mux.HandleFunc("/operators/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.Registry.GetOperator(w, r)
			return
		}
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
	})

	// Заказчики
	mux.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.Registry.RegisterCustomer(w, r)
		} else {
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	// Получение заказчика по ID
	mux.HandleFunc("/customers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.Registry.GetCustomer(w, r)
			return
		}
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
	})

	// Отдача статичного фронта для удобства))))
	mux.Handle("/", http.FileServer(http.Dir("./frontend")))

	return enableCORS(mux)
}
