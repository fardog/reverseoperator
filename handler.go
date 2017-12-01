package reverseoperator

import (
	"encoding/json"
	"fmt"
	"net/http"

	secop "github.com/fardog/secureoperator"
)

type HandlerOptions struct{}

func NewHandler(provider secop.Provider, options *HandlerOptions) *Handler {
	return &Handler{
		options:  options,
		provider: provider,
	}
}

type Handler struct {
	options  *HandlerOptions
	provider secop.Provider
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	q, err := urlToDNSQuestion(r.URL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	resp, err := h.provider.Query(*q)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, err)
		return
	}

	gdns := fromDNStoGDNS(resp)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(gdns); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
}
