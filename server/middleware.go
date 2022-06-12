package server

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey string

const ctxKeyID ctxKey = "id"

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w}
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func withUID(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxKeyID, uuid.New().String())
		next(w, r.WithContext(ctx))
	}
}

func withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("(id=%v) Request: %v", r.Context().Value(ctxKeyID), r.RequestURI)
		lw := newLoggingResponseWriter(w)
		next(lw, r)
		log.Printf("(id=%v) Response: %v", r.Context().Value(ctxKeyID), lw.statusCode)
	}
}

func withMiddleware(fs ...middleware) middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		for i := len(fs) - 1; i >= 0; i-- {
			next = fs[i](next)
		}
		return next
	}
}
