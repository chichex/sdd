package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestSpaHandlerFallsBackToIndex(t *testing.T) {
	dist := fstest.MapFS{
		"index.html":    {Data: []byte("<html>app</html>")},
		"assets/app.js": {Data: []byte("console.log(1)")},
	}
	h := spaHandler(dist)

	for _, route := range []string{"/", "/trabajos/42", "/index.html"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, route, nil))
		if rec.Code != http.StatusOK || rec.Body.String() != "<html>app</html>" {
			t.Fatalf("%s: esperaba index.html, obtuve %d %q", route, rec.Code, rec.Body.String())
		}
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "console.log(1)" {
		t.Fatalf("asset: esperaba el archivo, obtuve %d %q", rec.Code, rec.Body.String())
	}
}

func TestSpaHandlerWithoutBuild(t *testing.T) {
	rec := httptest.NewRecorder()
	spaHandler(fstest.MapFS{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("sin index.html esperaba 503, obtuve %d", rec.Code)
	}
}
