package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/johnmikee/cuebert/handlers/middlewares"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Reporting is an interface for the reporting information.
type Reporting interface {
	MakePie() []PieInfo
	MakeLines() *LineInfo
}

// Handler represents the http handlers.
type Handler struct {
	Status    *StatusHandler
	Reporting bool
	Grahps    Reporting
	Log       logger.Logger
}

// Server represents the client handling the http requests.
type Server struct {
	log    logger.Logger
	port   string
	router *chi.Mux
	status *StatusHandler
	graphs Reporting
}

// StartHandlers starts the http handlers.
func StartHandlers(h *Handler) {
	m := chi.NewRouter()

	s := &Server{
		log:    h.Log,
		router: m,
		port:   "8888",
		status: h.Status,
		graphs: h.Grahps,
	}

	s.registerMiddlewares()

	s.registerRoutes()

	s.serve()
}

func (s *Server) registerMiddlewares() {
	s.router.Use(
		middlewares.Logger(&s.log), // Logger is our custom logging middleware.
		middleware.Recoverer,       // Recoverer is a middleware that recovers from panics, logs the panic (and a backtrace), and returns a HTTP 500 (Internal Server Error) status if possible.
		middleware.RealIP,          // RealIP is a middleware that sets a http.Request's RemoteAddr to the results of parsing either the X-Forwarded-For or X-Real-IP headers (in that order).
		middleware.RequestID,       // RequestID is a middleware that injects a request ID into the context of each request.
	)
}

func (s *Server) registerRoutes() {
	s.router.Get("/health", s.healthHandler)
	s.router.Get("/healthcheck", s.healthChecker)
	s.router.Get("/reporting", s.reportHandler)
}

func (s *Server) serve() {
	s.log.Info().Str("listenAddr", s.port).Msg("starting server")
	server := &http.Server{
		Addr:              ":" + s.port,
		Handler:           s.router, // set the chi router as the handler
		ReadHeaderTimeout: 3 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		s.log.Debug().AnErr("error", err).Msg("failed to listen and serve")
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	_, err := w.Write([]byte("200 OK"))
	if err != nil {
		s.log.Debug().AnErr("error", err).Msg("failed to write response")
	}
}

func (s *Server) healthChecker(w http.ResponseWriter, r *http.Request) {
	status := s.status.GetStatus() // Read the status from the handler

	jsonData, err := json.Marshal(status)
	if err != nil {
		s.log.Debug().AnErr("error", err).Msg("failed to marshal json")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(jsonData)
	if err != nil {
		s.log.Debug().AnErr("error", err).Msg("failed to write response")
		return
	}
}

func (s *Server) reportHandler(w http.ResponseWriter, r *http.Request) {
	for _, p := range s.graphs.MakePie() {
		err := makePie(p, w)
		if err != nil {
			s.log.Debug().AnErr("error", err).Msg("failed to make pie")
		}
	}
	err := makeLines(s.graphs.MakeLines(), w)
	if err != nil {
		s.log.Debug().AnErr("error", err).Msg("failed to make lines")
	}
}
