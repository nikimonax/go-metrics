package middleware

import "github.com/go-chi/chi/v5/middleware"

// some re-exports
var (
	Compress             = middleware.Compress
	CleanPath            = middleware.CleanPath
	AllowContentType     = middleware.AllowContentType
	AllowContentEncoding = middleware.AllowContentEncoding
)
