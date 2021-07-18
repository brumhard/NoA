package dnsinjector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Handler handles incoming admission requests.
type Handler struct {
	mux         *http.ServeMux
	logger      *zap.Logger
	annotations []string
}

// NewHandler creates a new Handler with given configuration.
func NewHandler(logger *zap.Logger, annotations []string) *Handler {
	h := Handler{mux: http.NewServeMux(), logger: logger, annotations: annotations}
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

var rfc6901Encoder = strings.NewReplacer("~", "~0", "/", "~1")

// handleMutate handles mutate requests.
func (h *Handler) handleMutate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.logger.Info("called /mutate")

		var admissionReview admissionv1.AdmissionReview
		if err := json.NewDecoder(r.Body).Decode(&admissionReview); err != nil {
			h.writeError(w, err, http.StatusInternalServerError)
			return
		}

		if admissionReview.Request == nil {
			return
		}

		var someObjectWithAnnotations struct {
			metav1.TypeMeta   `json:",inline"`
			metav1.ObjectMeta `json:"metadata,omitempty"`
		}

		if err := json.Unmarshal(admissionReview.Request.Object.Raw, &someObjectWithAnnotations); err != nil {
			h.writeError(w, err, http.StatusInternalServerError)
		}

		var patches []map[string]interface{}

		for _, a := range h.annotations {
			if _, ok := someObjectWithAnnotations.Annotations[a]; ok {
				patches = append(patches, map[string]interface{}{
					"op":   "remove",
					"path": fmt.Sprintf("/metadata/annotations/%s", rfc6901Encoder.Replace(a)),
				})
			}
		}

		patchesBytes, err := json.Marshal(patches)
		if err != nil {
			h.writeError(w, err, http.StatusInternalServerError)
			return
		}

		h.logger.Info(
			"applying patches",
			zap.String("secret", admissionReview.Request.Name),
			zap.String("namespace", admissionReview.Request.Namespace),
			zap.ByteString("patches", patchesBytes),
		)

		jsonPatch := admissionv1.PatchTypeJSONPatch
		admissionReview.Response = &admissionv1.AdmissionResponse{
			UID:       admissionReview.Request.UID,
			Allowed:   true,
			Patch:     patchesBytes,
			PatchType: &jsonPatch,
		}

		if err := json.NewEncoder(w).Encode(admissionReview); err != nil {
			h.writeError(w, err, http.StatusInternalServerError)
			return
		}
	}
}

func (h *Handler) handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "OK")
	}
}

func (h *Handler) writeError(w http.ResponseWriter, err error, code int) {
	h.logger.Error(err.Error())
	http.Error(w, err.Error(), code)
}
