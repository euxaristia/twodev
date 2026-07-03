package api

import "net/http"

func (h *Handler) route(mux *http.ServeMux, pattern string, fn http.HandlerFunc) {
	if h.guard != nil && h.guard.Enabled() {
		mux.Handle(pattern, h.guard.Middleware(http.HandlerFunc(fn)))
		return
	}
	mux.HandleFunc(pattern, fn)
}