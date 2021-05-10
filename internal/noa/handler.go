package dnsinjector

import (
	"net/http"
	"os"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Handler handles incoming admission requests.
type Handler struct {
	logger     *zap.SugaredLogger
	kubeClient *kubernetes.Clientset
}

// NewHandler creates a new Handler with given configuration.
func NewHandler(logger *zap.SugaredLogger) (*Handler, error) {
	// If KUBECONFIG is not set BuildConfigFromFlags will use InClusterConfig()
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Handler{logger: logger, kubeClient: kubeClient}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()

	// TODO: check http method in handleFunc
	mux.HandleFunc("/mutate", h.handleMutate())
	mux.HandleFunc("/healthz", h.handleHealthz())

	mux.ServeHTTP(w, r)
}

// handleMutate handles mutate requests.
func (h *Handler) handleMutate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *Handler) handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	}
}
