package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/impl"
)

type Server struct {
	config *ServerConfig
	router chi.Router
}

func (s *Server) Run() {
	log.Printf("Starting server on %s\n", s.config.BaseURL)

	err := http.ListenAndServe(s.config.BaseURL, s.router)

	if err != nil {
		log.Println(err)
	}
}

func New(config *ServerConfig) *Server {
	metricRepository := impl.NewInMemoryMetricRepository()

	updateMetricUseCase := app.NewUpdateMetricUseCase(metricRepository)
	getMetricUseCase := app.NewGetMetricUseCase(metricRepository)
	getAllMetricsUseCase := app.NewGetAllMetricsUseCase(metricRepository)

	plainTextErrorPresenter := NewPlainTextErrorPresenter()
	plainTextMetricPresenter := NewPlainTextMetricPresenter()
	htmlTableMetricsPresenter := NewHtmlTableMetricsPresenter()

	updateMetricHandler := NewUpdateMetricHandler(
		updateMetricUseCase,
		plainTextErrorPresenter,
	)
	getMetricHandler := NewGetMetricHandler(
		getMetricUseCase,
		plainTextErrorPresenter,
		plainTextMetricPresenter,
	)
	PreviewMetricsHandler := NewPreviewMetricsHandler(
		getAllMetricsUseCase,
		plainTextErrorPresenter,
		htmlTableMetricsPresenter,
	)

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/", PreviewMetricsHandler.ServeHTTP)
	router.Post(
		"/update/{metricType}/{metricName}/{metricValue}",
		updateMetricHandler.ServeHTTP,
	)
	router.Get(
		"/value/{metricType}/{metricName}",
		getMetricHandler.ServeHTTP,
	)

	return &Server{
		config: config,
		router: router,
	}
}
