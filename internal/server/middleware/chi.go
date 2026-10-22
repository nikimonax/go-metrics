package middleware

import "github.com/go-chi/chi/v5/middleware"

// some re-exports
var (
	NewCompressor          = middleware.NewCompressor
	CleanPath              = middleware.CleanPath
	AllowContentType       = middleware.AllowContentType
	AllowContentEncoding   = middleware.AllowContentEncoding
	ClientIPFromRemoteAddr = middleware.ClientIPFromRemoteAddr
	GetClientIP            = middleware.GetClientIP
)
