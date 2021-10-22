// Package api sets up the HTTP server, configures middleware and routes
package api

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"

	"github.com/gp42/pf-demo/pkg/api/handler"
	"github.com/gp42/pf-demo/pkg/db"
	m "github.com/gp42/pf-demo/pkg/mailer"
)

// HTTP server object
type Server struct {
	httpServer     *http.Server
	router         *mux.Router
	handler        *handler.Handler
	log            *logr.Logger
	tlsCertPath    *string
	tlsCertKeyPath *string
}

// Create a new HTTP server
// listenAddr must be in format: "<INTERFACE_ADDRESS>:<PORT>"
func NewServer(listenAddr string, dbConn *db.DBConnection, mailer *m.Mailer, logger *logr.Logger, certPath, certKeyPath *string) *Server {
	log := logger.WithValues("server_addr", listenAddr)
	requestMiddleware := &Middleware{
		DB:  dbConn,
		Log: &log,
	}
	handler := handler.Handler{DB: dbConn, Mailer: mailer}

	router := mux.NewRouter()
	// Healthz is not using main middleware to avoid extra logging and request blocking
	// Another approach might be to use custom healthchecker headers to verify request source
	rHealth := router.PathPrefix("/healthz").Subrouter()
	rHealth.Path("").Methods("GET", "HEAD").HandlerFunc(handler.Healthz)
	rApi := router.PathPrefix("").Subrouter()
	rApi.Use(requestMiddleware.RequestMiddleware)
	initRoutes(rApi, &handler)

	httpServer := &http.Server{
		Addr:    listenAddr,
		Handler: router,
	}

	s := &Server{
		httpServer:     httpServer,
		router:         router,
		handler:        &handler,
		log:            &log,
		tlsCertPath:    certPath,
		tlsCertKeyPath: certKeyPath,
	}

	return s
}

// Start an instance of HTTP server
//
// example usage:
//
//		ch := server.Start()
//		err = <-ch
//		if err != nil {
//			panic(err.Error())
//		}
//
func (s *Server) Start() chan error {
	ch := make(chan error)

	s.log.Info("Starting http server")
	go func() {
		if *s.tlsCertPath != "" && *s.tlsCertKeyPath != "" {
			err := s.httpServer.ListenAndServeTLS(*s.tlsCertPath, *s.tlsCertKeyPath)
			ch <- err
		} else {
			err := s.httpServer.ListenAndServe()
			ch <- err
		}
	}()
	return ch
}

// Initialize HTTP Server routes
func initRoutes(r *mux.Router, h *handler.Handler) {

	// Square a number
	r.Path("/").Queries("n", "{n:[-]?[0-9]+}").Methods("POST").HandlerFunc(h.Square)

	// Blacklist
	r.Path("/blacklisted").Methods("GET", "HEAD").HandlerFunc(h.Blacklisted)
}
