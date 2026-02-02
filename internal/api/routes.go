package api

import "net/http"

func (h *Handler) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/order", h.GetOrder)
	mux.HandleFunc("POST /api/order", h.CreateOrder)
	mux.HandleFunc("GET /", h.ServeStatic)
	mux.HandleFunc("GET /script.js", h.ServeJS)
	mux.HandleFunc("GET /styles.css", h.ServeCSS)
}
