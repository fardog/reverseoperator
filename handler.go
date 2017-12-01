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

	// these headers match google's service, as off as they may seem
	w.Header().Set("content-type", "application/x-javascript; charset=UTF-8")
	w.Header().Set("cache-control", "private")
	w.Header().Set("x-xss-protection", "1; mode=block")
	w.Header().Set("x-frame-options", "SAMEORIGIN")

	enc := json.NewEncoder(w)
	if err := enc.Encode(gdns); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
}
