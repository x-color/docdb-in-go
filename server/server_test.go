package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/gorilla/mux"
	"github.com/x-color/docdb-in-go/docdb"
)

func TestServer_AddDocumentHandler(t *testing.T) {
	tests := []struct {
		name     string
		server   Server
		reqBody  string
		wantCode int
		wantDoc  map[string]any
	}{
		{
			name: "Create document",
			server: Server{
				docdb: docdb.NewDocDB(),
			},
			reqBody:  `{"greeting":"hello"}`,
			wantCode: http.StatusCreated,
			wantDoc: map[string]any{
				"greeting": "hello",
			},
		},
		{
			name: "Create invalid document",
			server: Server{
				docdb: docdb.NewDocDB(),
			},
			reqBody:  `{"greeting":"hello"`,
			wantCode: http.StatusBadRequest,
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

			if rr.Code != tt.wantCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantCode)
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

			if diff := cmp.Diff(tt.wantDoc, v); diff != "" {
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
		wantCode   int
		wantDoc    map[string]any
	}{
		{
			name: "Get document",
			server: Server{
				docdb: docdb.NewDocDB(),
			},
			wantCode: http.StatusOK,
			wantDoc: map[string]any{
				"greeting": "hello",
			},
		},
		{
			name: "Not found document",
			server: Server{
				docdb: docdb.NewDocDB(),
			},
			overrideID: "not-found",
			wantCode:   http.StatusNotFound,
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

			if rr.Code != tt.wantCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantCode)
			}

			if rr.Code != http.StatusOK {
				return
			}

			res := make(map[string]any)
			if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
				t.Errorf("handler returned invalid body: got %v", rr.Body.String())
			}

			if diff := cmp.Diff(tt.wantDoc, res); diff != "" {
				t.Errorf("document mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestServer_SearchDocumentHandler(t *testing.T) {
	tests := []struct {
		name     string
		server   Server
		docs     []map[string]any
		q        string
		wantCode int
		wantRes  map[string]any
	}{
		{
			name: "Search document",
			server: Server{
				docdb: docdb.NewDocDB(),
			},
			docs: []map[string]any{
				{
					"greeting": "hello",
					"obj": map[string]any{
						"num": "1",
					},
				},
				{
					"greeting": "hello",
				},
			},
			q:        "obj.num:1",
			wantCode: http.StatusOK,
			wantRes: map[string]any{
				"documents": []any{
					map[string]any{
						"document": map[string]any{
							"greeting": "hello",
							"obj": map[string]any{
								"num": "1",
							},
						},
					},
				},
				"count": float64(1),
			},
		},
		{
			name: "Search documents",
			server: Server{
				docdb: docdb.NewDocDB(),
			},
			docs: []map[string]any{
				{
					"greeting": "hello",
					"num":      1,
				},
				{
					"greeting": "hello",
					"num":      2,
				},
			},
			q:        "greeting:hello",
			wantCode: http.StatusOK,
			wantRes: map[string]any{
				"documents": []any{
					map[string]any{
						"document": map[string]any{
							"greeting": "hello",
							"num":      float64(1),
						},
					},
					map[string]any{
						"document": map[string]any{
							"greeting": "hello",
							"num":      float64(2),
						},
					},
				},
				"count": float64(2),
			},
		},
		{
			name: "Not found document",
			server: Server{
				docdb: docdb.NewDocDB(),
			},
			docs: []map[string]any{
				{
					"greeting": "hello",
					"num":      1,
				},
				{
					"greeting": "hello",
					"num":      2,
				},
			},
			q:        "a.b.:1",
			wantCode: http.StatusNotFound,
			wantRes:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, doc := range tt.docs {
				_, err := tt.server.docdb.Add(doc)
				if err != nil {
					t.Fatalf("failed to add data to DB for preparing test: %v", err)
				}
			}

			req, err := http.NewRequest("GET", "/docs?q="+tt.q, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/docs", tt.server.SearchDocumentsHandler)
			router.ServeHTTP(rr, req)

			if rr.Code != tt.wantCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantCode)
			}

			if rr.Code != http.StatusOK {
				return
			}

			res := make(map[string]any)
			if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
				t.Errorf("handler returned invalid body: got %v", rr.Body.String())
			}

			sortOpt := cmp.Transformer("sort", func(in map[string]any) map[string]any {
				docs, ok := in["documents"].([]any)
				if !ok {
					return in
				}
				sort.Slice(docs, func(i, j int) bool {
					hi := fmt.Sprintf("%s", sha256.Sum256([]byte(fmt.Sprintf("%v", docs[i]))))
					hj := fmt.Sprintf("%s", sha256.Sum256([]byte(fmt.Sprintf("%v", docs[j]))))
					return hi < hj
				})
				return map[string]any{
					"documents": docs,
					"count":     in["count"],
				}
			})
			ignoreOpt := cmpopts.IgnoreMapEntries(func(k, v any) bool {
				return k == "id"
			})

			if diff := cmp.Diff(tt.wantRes, res, sortOpt, ignoreOpt); diff != "" {
				t.Errorf("document mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
