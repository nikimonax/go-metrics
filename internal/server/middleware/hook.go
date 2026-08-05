package middleware

import "net/http"

type RequestHook struct {
	beforeRequest []func(*http.Request) error
	afterRequest  []func(*http.Request) error
	OnError       func(error)
}

func (h *RequestHook) BeforeRequest(hook func(*http.Request) error) {
	h.beforeRequest = append(h.beforeRequest, hook)
}

func (h *RequestHook) AfterRequest(hook func(*http.Request) error) {
	h.afterRequest = append(h.afterRequest, hook)
}

func (h *RequestHook) Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		h.execHooks(h.beforeRequest, r)
		next.ServeHTTP(w, r)
		h.execHooks(h.afterRequest, r)
	}
	return http.HandlerFunc(fn)
}

func (h *RequestHook) execHooks(hooks []func(*http.Request) error, r *http.Request) {
	for _, hook := range hooks {
		if err := hook(r); h.OnError != nil && err != nil {
			h.OnError(err)
		}
	}
}

func NewRequestHook() *RequestHook {
	return &RequestHook{
		beforeRequest: make([]func(*http.Request) error, 0),
		afterRequest:  make([]func(*http.Request) error, 0),
	}
}
