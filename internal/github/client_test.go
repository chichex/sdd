package github

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/chichex/sdd/internal/github/fake"
)

var now = time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)

func TestFetch_RespuestaGrabada(t *testing.T) {
	gh := &fake.Runner{T: t, Dir: "../../testdata/gh"}
	snap, err := Fetch(context.Background(), gh, "pramaestudio/platform", now)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if snap.Repo != "pramaestudio/platform" || snap.FetchedAt != "2026-10-27T12:00:00Z" || len(snap.Issues) != 4 || !snap.IssuesComplete {
		t.Fatalf("snapshot = repo %q fetchedAt %q issues %d completa %v", snap.Repo, snap.FetchedAt, len(snap.Issues), snap.IssuesComplete)
	}
	if len(snap.PRs) != 16 {
		t.Fatalf("PRs = %d, want 16 (abiertos y recientes sin duplicar)", len(snap.PRs))
	}
	p113, _ := snap.PR(113)
	p111, _ := snap.PR(111)
	p97, _ := snap.PR(97)
	if p113.UnresolvedThreads != 1 || len(p111.Checks) != 2 || p111.Checks[0].WorkflowName != "CI" || p97.MergeCommit == "" || p97.MergedAt == "" {
		t.Fatalf("PRs convertidos = 113 %+v · 111 %+v · 97 %+v", p113, p111.Checks, p97)
	}
	if SubEstado(p111) != "checks rojos" {
		t.Fatalf("SubEstado(111) = %q, want checks rojos", SubEstado(p111))
	}
	calls := gh.Calls()
	if len(calls) != 1 || calls[0][0] != "api" || calls[0][1] != "graphql" {
		t.Fatalf("invocaciones = %q, want una sola consulta GraphQL", calls)
	}
}

func TestReadOnly_RechazaEscrituras(t *testing.T) {
	tests := []struct {
		name string
		args []string
		ok   bool
	}{
		{"consulta GraphQL", []string{"api", "graphql", "-f", "query=query { viewer { login } }"}, true},
		{"mutation", []string{"api", "graphql", "-f", "query=mutation { addComment }"}, false},
		{"método POST", []string{"api", "-X", "POST", "repos/o/r/issues"}, false},
		{"método DELETE pegado", []string{"api", "-XDELETE", "repos/o/r/git/refs/heads/x"}, false},
		{"método PATCH largo", []string{"api", "--method", "PATCH", "repos/o/r"}, false},
		{"pr merge", []string{"pr", "merge", "1"}, false},
		{"issue comment", []string{"issue", "comment", "1", "-b", "x"}, false},
		{"pr edit", []string{"pr", "edit", "1"}, false},
		{"issue create", []string{"issue", "create"}, false},
		{"issue close", []string{"issue", "close", "1"}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := ReadOnly(tc.args); (err == nil) != tc.ok {
				t.Fatalf("ReadOnly(%q) = %v, want ok %v", tc.args, err, tc.ok)
			}
		})
	}
}

func TestFetch_NoEjecutaEscrituras(t *testing.T) {
	r := &recording{}
	if _, err := Fetch(context.Background(), r, "o/r", now); err == nil {
		t.Fatalf("Fetch con respuesta vacía no falló")
	}
	for _, args := range r.calls {
		if err := ReadOnly(args); err != nil || strings.Contains(strings.Join(args, " "), "mutation") {
			t.Fatalf("Fetch invocó %q", args)
		}
	}
}

type recording struct{ calls [][]string }

func (r *recording) Run(_ context.Context, args ...string) ([]byte, error) {
	r.calls = append(r.calls, args)
	return []byte(`{}`), nil
}

func TestReason_Degradado(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"gh ausente", fmt.Errorf("x: %w", exec.ErrNotFound), ReasonSinGH},
		{"sin login", errors.New("exit status 4: To get started with GitHub CLI, please run:  gh auth login"), ReasonSinLogin},
		{"credenciales", errors.New("HTTP 401: Bad credentials (https://api.github.com/graphql)"), ReasonSinLogin},
		{"rate limit", errors.New("GraphQL: API rate limit exceeded for user ID 1."), ReasonRateLimit},
		{"repo inexistente", errors.New("GraphQL: Could not resolve to a Repository with the name 'o/r'. (repository)"), ReasonSinRemote},
		{"sin red", errors.New(`error connecting to api.github.com: dial tcp: lookup api.github.com: no such host`), ReasonRed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Reason(tc.err); got != tc.want {
				t.Fatalf("Reason = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFetch_ErrorDeGhSeClasifica(t *testing.T) {
	gh := &fake.Runner{T: t, Dir: "../../testdata/gh", Fail: map[string]string{"waica": "GraphQL: API rate limit exceeded"}}
	_, err := Fetch(context.Background(), gh, "chichex/waica", now)
	if err == nil || Reason(err) != ReasonRateLimit {
		t.Fatalf("Fetch = %v (%q), want rate-limit", err, Reason(err))
	}
}
