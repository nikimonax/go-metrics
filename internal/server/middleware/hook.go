package middleware

import "net/http"

type RequestHook struct {
	beforeRequest []func(*http.Request)
	afterRequest  []func(*http.Request)
}

func (h *RequestHook) BeforeRequest(hook func(*http.Request)) {
	h.beforeRequest = append(h.beforeRequest, hook)
}

func (h *RequestHook) AfterRequest(hook func(*http.Request)) {
	h.afterRequest = append(h.afterRequest, hook)
}

func (h *RequestHook) Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		execHooks(h.beforeRequest, r)
		next.ServeHTTP(w, r)
		execHooks(h.afterRequest, r)
	}
	return http.HandlerFunc(fn)
}

func execHooks(hooks []func(*http.Request), r *http.Request) {
	for _, hook := range hooks {
		hook(r)
	}
}

func NewRequestHook() *RequestHook {
	return &RequestHook{
		beforeRequest: make([]func(*http.Request), 0),
		afterRequest:  make([]func(*http.Request), 0),
	}
}
