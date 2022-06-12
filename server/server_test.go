package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/x-color/docdb-in-go/docdb"
)

func TestServer_AddDocumentHandler(t *testing.T) {
	tests := []struct {
		name    string
		server  Server
		reqBody string
		code    int
		doc     map[string]any
	}{
		{
			name: "Create document",
			server: Server{
				docdb: docdb.NewDocDB(),
			},
			reqBody: `{"greeting":"hello"}`,
			code:    http.StatusCreated,
			doc: map[string]any{
				"greeting": "hello",
			},
		},
		{
			name: "Create invalid document",
			server: Server{
				docdb: docdb.NewDocDB(),
			},
			reqBody: `{"greeting":"hello"`,
			code:    http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/docs", bytes.NewBufferString(tt.reqBody))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/docs", tt.server.AddDocumentHandler)
			router.ServeHTTP(rr, req)

			if rr.Code != tt.code {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.code)
			}

			if rr.Code != http.StatusCreated {
				return
			}

			res := struct {
				ID string
			}{}
			if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
				t.Errorf("handler returned invalid body: got %v", rr.Body.String())
			}
			v, err := tt.server.docdb.Get(res.ID)
			if err != nil {
				t.Errorf("can not get document: %v", err)
			}

			if diff := cmp.Diff(tt.doc, v); diff != "" {
				t.Errorf("document mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestServer_GetDocumentHandler(t *testing.T) {
	tests := []struct {
		name       string
		server     Server
		overrideID string
		code       int
		doc        map[string]any
	}{
		{
			name: "Get document",
			server: Server{
				docdb: docdb.NewDocDB(),
			},
			code: http.StatusOK,
			doc: map[string]any{
				"greeting": "hello",
			},
		},
		{
			name: "Not found document",
			server: Server{
				docdb: docdb.NewDocDB(),
			},
			overrideID: "not-found",
			code:       http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testdata := map[string]any{
				"greeting": "hello",
			}
			id, err := tt.server.docdb.Add(testdata)
			if err != nil {
				t.Fatalf("failed to add data to DB for preparing test: %v", err)
			}

			if tt.overrideID != "" {
				id = tt.overrideID
			}
			req, err := http.NewRequest("GET", "/docs/"+id, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/docs/{id}", tt.server.GetDocumentHandler)
			router.ServeHTTP(rr, req)

			if rr.Code != tt.code {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.code)
			}

			if rr.Code != http.StatusOK {
				return
			}

			res := make(map[string]any)
			if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
				t.Errorf("handler returned invalid body: got %v", rr.Body.String())
			}

			if diff := cmp.Diff(tt.doc, res); diff != "" {
				t.Errorf("document mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestServer_SearchDocumentHandler(t *testing.T) {
	tests := []struct {
		name   string
		server Server
		q      string
		code   int
		doc    map[string]any
	}{
		{
			name: "Search document",
			server: Server{
				docdb: docdb.NewDocDB(),
			},
			q:    "greeting:hello",
			code: http.StatusOK,
			doc: map[string]any{
				"document": map[string]any{
					"greeting": "hello",
				},
			},
		},
		{
			name: "Not found document",
			server: Server{
				docdb: docdb.NewDocDB(),
			},
			q:    "a.b.:1",
			code: http.StatusNotFound,
			doc:  make(map[string]any),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testdata := map[string]any{
				"greeting": "hello",
			}
			id, err := tt.server.docdb.Add(testdata)
			if err != nil {
				t.Fatalf("failed to add data to DB for preparing test: %v", err)
			}
			tt.doc["id"] = id

			req, err := http.NewRequest("GET", "/docs?q="+tt.q, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/docs", tt.server.SearchDocumentsHandler)
			router.ServeHTTP(rr, req)

			if rr.Code != tt.code {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.code)
			}

			if rr.Code != http.StatusOK {
				return
			}

			res := make(map[string]any)
			if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
				t.Errorf("handler returned invalid body: got %v", rr.Body.String())
			}

			if diff := cmp.Diff(tt.doc, res); diff != "" {
				t.Errorf("document mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
