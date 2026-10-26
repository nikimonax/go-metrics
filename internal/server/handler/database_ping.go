package handler

import (
	"database/sql"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

type PingDatabaseHandler struct {
	db *sql.DB
}

// PingDatabaseHandler godoc
//
//	@ID				pingDatabase
//	@Summary		Проверка доступности БД
//	@Description	Возвращает 200, если доступ к базе данных работает.
//	@Tags			health
//	@Produce		text/plain
//	@Success		200	{string}	string	"OK"
//	@Failure		500	{string}	string	"Ошибка подключения к БД"
//	@Router			/ping [get]
//
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

var _ http.Handler = (*PingDatabaseHandler)(nil)

func NewPingDatabaseHandler(db *sql.DB) *PingDatabaseHandler {
	return &PingDatabaseHandler{db: db}
}
