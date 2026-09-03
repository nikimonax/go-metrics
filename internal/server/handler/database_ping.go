package handler

import (
	"database/sql"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

type PingDatabaseHandler struct {
	db *sql.DB
}

// ServeHTTP implements [http.Handler].
func (h *PingDatabaseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.db.PingContext(r.Context())

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(httpextra.HDRContentType, httpextra.MIMEText)
	w.WriteHeader(http.StatusOK)
}

func NewPingDatabaseHandler(db *sql.DB) http.Handler {
	return &PingDatabaseHandler{db: db}
}
