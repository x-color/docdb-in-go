package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	server *http.Server
	wait   time.Duration
}

func (s Server) defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s Server) AddDocumentHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (s Server) SearchDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (s Server) GetDocumentHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (s Server) Start() error {
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	s.waitSignal(os.Interrupt)
	return s.Shutdown()
}

func (s Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.wait)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s Server) waitSignal(sigs ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, sigs...)
	<-c
}

func NewServer(addr string, port int) Server {
	s := Server{
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", addr, port),
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
		},
		wait: time.Second * 15,
	}
	r := mux.NewRouter()
	r.HandleFunc("/docs", s.AddDocumentHandler).Methods("POST")
	r.HandleFunc("/docs", s.SearchDocumentsHandler).Methods("GET")
	r.HandleFunc("/docs/{id}", s.GetDocumentHandler).Methods("GET")
	r.HandleFunc("/", s.defaultHandler)
	s.server.Handler = r

	return s
}
