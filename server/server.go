package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/x-color/docdb-in-go/docdb"
	"github.com/x-color/docdb-in-go/query"
)

type middleware func(http.HandlerFunc) http.HandlerFunc

type Server struct {
	docdb  *docdb.DocDB
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

	id, err := s.docdb.Add(doc)
	if err != nil {
		errResponse(w, http.StatusInternalServerError, nil)
		return
	}

	response(w, http.StatusCreated, map[string]any{
		"id": id,
	})
}

func (s Server) SearchDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	q, err := query.ParseQuery(r.URL.Query().Get("q"))
	if err != nil {
		log.Printf("(id=%v) Not found document: %v", r.Context().Value(ctxKeyID), err)
		errResponse(w, http.StatusBadRequest, err)
		return
	}

	docs, err := s.docdb.Search(q)
	if err != nil {
		errResponse(w, http.StatusInternalServerError, nil)
		return
	}
	if len(docs) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response(w, http.StatusOK, map[string]any{
		"documents": docs,
		"count":     len(docs),
	})
}

func (s Server) GetDocumentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	doc, err := s.docdb.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, docdb.ErrNotFound):
			errResponse(w, http.StatusNotFound, nil)
		default:
			errResponse(w, http.StatusInternalServerError, nil)
		}
		return
	}

	response(w, http.StatusOK, doc)
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
		docdb: docdb.NewDocDB(),
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", addr, port),
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		wait: 15 * time.Second,
	}

	with := withMiddleware(withUID, withLogging)

	r := mux.NewRouter()
	r.HandleFunc("/docs", with(s.AddDocumentHandler)).Methods("POST")
	r.HandleFunc("/docs", with(s.SearchDocumentsHandler)).Methods("GET")
	r.HandleFunc("/docs/{id}", with(s.GetDocumentHandler)).Methods("GET")
	r.HandleFunc("/", with(s.defaultHandler))
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
