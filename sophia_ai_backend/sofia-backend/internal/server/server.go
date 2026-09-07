package server

import (
	"net/http"

	"github.com/armandoalvarado/sofia-backend/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func New(cfg config.Config, routes Routes) *Server {
	handler := routes.Handler()
	handler = failedRequestLoggingMiddleware(handler)
	handler = recoverMiddleware(handler)
	handler = corsMiddleware(cfg.CORSAllowedOrigins, handler)

	return &Server{
		httpServer: &http.Server{
			Addr:              cfg.Addr(),
			Handler:           handler,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		},
	}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}
