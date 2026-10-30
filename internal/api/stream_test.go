package api

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chichex/sdd/internal/workspace"
)

func TestStream_PingPeriodico(t *testing.T) {
	if DefaultPing != 30*time.Second {
		t.Fatalf("DefaultPing = %v, want 30s", DefaultPing)
	}
	svc := workspace.New(workspace.Config{Root: t.TempDir(), WatchInterval: time.Hour})
	api := New(svc, "test")
	api.ping = 20 * time.Millisecond
	mux := http.NewServeMux()
	api.Register(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/stream", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	sc := bufio.NewScanner(res.Body)
	for sc.Scan() {
		if sc.Text() == ": ping" {
			return
		}
	}
	t.Fatalf("no llegó ningún ping: %v", sc.Err())
}
