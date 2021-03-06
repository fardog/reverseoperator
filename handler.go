package reverseoperator

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"

	secop "github.com/fardog/secureoperator"
)

type HandlerOptions struct {
	ContentTypeJSON bool
	ServerHeader    string
}

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
	fail := func(status int, err error) {
		w.WriteHeader(status)
		fmt.Fprint(w, err)
		log.Error(err)
	}

	q, err := urlToDNSQuestion(r.URL)
	if err != nil {
		fail(http.StatusBadRequest, err)
		return
	}

	resp, err := h.provider.Query(*q)
	if err != nil {
		fail(http.StatusServiceUnavailable, err)
		return
	}

	gdns := fromDNStoGDNS(resp)

	// these headers match google's service, as off as they may seem; we allow
	// some override to more standard headers, however
	if h.options.ContentTypeJSON {
		w.Header().Set("content-type", "application/json; charset=UTF-8")
	} else {
		// this is google's content type
		w.Header().Set("content-type", "application/x-javascript; charset=UTF-8")
	}
	w.Header().Set("cache-control", "private")
	w.Header().Set("x-xss-protection", "1; mode=block")
	w.Header().Set("x-frame-options", "SAMEORIGIN")
	if h.options.ServerHeader != "" {
		w.Header().Set("server", h.options.ServerHeader)
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(gdns); err != nil {
		fail(http.StatusInternalServerError, err)
		return
	}

	log.Infof("responded to request %v[%v]", q.Name, q.Type)
}
