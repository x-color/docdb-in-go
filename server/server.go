package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

type Server struct {
	db     *cache.Cache
	server *http.Server
	wait   time.Duration
}

func (s Server) defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s Server) AddDocumentHandler(w http.ResponseWriter, r *http.Request) {
	doc := make(map[string]any)
	dc := json.NewDecoder(r.Body)
	if err := dc.Decode(&doc); err != nil {
		errResponse(w, http.StatusBadRequest, err)
		return
	}

	id := uuid.New().String()
	b, err := json.Marshal(doc)
	if err != nil {
		errResponse(w, http.StatusInternalServerError, nil)
		return
	}

	s.db.Set(id, b, 0)

	response(w, http.StatusCreated, map[string]any{
		"id": id,
	})
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
		db: cache.New(30*time.Minute, 10*time.Minute),
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", addr, port),
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		wait: 15 * time.Second,
	}
	r := mux.NewRouter()
	r.HandleFunc("/docs", s.AddDocumentHandler).Methods("POST")
	r.HandleFunc("/docs", s.SearchDocumentsHandler).Methods("GET")
	r.HandleFunc("/docs/{id}", s.GetDocumentHandler).Methods("GET")
	r.HandleFunc("/", s.defaultHandler)
	s.server.Handler = r

	return s
}

func response(w http.ResponseWriter, code int, body map[string]any) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(body)
	if err != nil {
		return
	}
	w.Write(b)
}

func errResponse(w http.ResponseWriter, code int, err error) {
	body := map[string]any{}
	if err != nil {
		body["error"] = err.Error()
	}
	response(w, code, body)
}
