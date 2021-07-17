package dnsinjector

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// Handler handles incoming admission requests.
type Handler struct {
	mux    *http.ServeMux
	logger *zap.Logger
}

// NewHandler creates a new Handler with given configuration.
func NewHandler(logger *zap.Logger) *Handler {
	h := Handler{mux: http.NewServeMux(), logger: logger}
	h.routes()

	return &h
}

func (h *Handler) routes() {
	h.mux.HandleFunc("/mutate", h.handleMutate())
	h.mux.HandleFunc("/healthz", h.handleHealthz())
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// handleMutate handles mutate requests.
func (h *Handler) handleMutate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.logger.Info("called /mutate")

	}
}

func (h *Handler) handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "OK")
	}
}
