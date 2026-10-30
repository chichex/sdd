package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"
)

// Runner ejecuta `gh` con los argumentos dados y devuelve su stdout.
type Runner interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
}

// ExecRunner ejecuta el `gh` del PATH con un timeout por invocación.
type ExecRunner struct {
	Timeout time.Duration
}

// Run ejecuta `gh`; el error incluye el stderr para poder clasificarlo.
func (e ExecRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	path, err := exec.LookPath("gh")
	if err != nil {
		return nil, fmt.Errorf("github: buscar gh: %w", err)
	}
	timeout := e.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return stdout.Bytes(), fmt.Errorf("github: gh %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

var writeSubcommands = []string{"edit", "create", "merge", "close", "comment", "delete", "reopen"}

// ErrWrite es el rechazo de una invocación que escribiría en GitHub.
var ErrWrite = errors.New("github: el visor es solo lectura")

// ReadOnly rechaza toda invocación que pueda escribir: mutations, métodos
// HTTP distintos de GET y subcomandos de escritura.
func ReadOnly(args []string) error {
	for i, a := range args {
		if strings.Contains(strings.ToLower(a), "mutation") {
			return fmt.Errorf("%w: mutation", ErrWrite)
		}
		method := ""
		switch {
		case (a == "-X" || a == "--method") && i+1 < len(args):
			method = args[i+1]
		case strings.HasPrefix(a, "-X") && len(a) > 2:
			method = a[2:]
		case strings.HasPrefix(a, "--method="):
			method = strings.TrimPrefix(a, "--method=")
		}
		if method != "" && !strings.EqualFold(method, "GET") {
			return fmt.Errorf("%w: método %s", ErrWrite, method)
		}
	}
	if len(args) > 1 && slices.Contains(writeSubcommands, args[1]) {
		return fmt.Errorf("%w: %s %s", ErrWrite, args[0], args[1])
	}
	return nil
}

// Reason clasifica un error de GitHub en una razón de degradado.
func Reason(err error) string {
	if errors.Is(err, exec.ErrNotFound) {
		return ReasonSinGH
	}
	text := strings.ToLower(err.Error())
	switch {
	case strings.Contains(text, "rate limit"):
		return ReasonRateLimit
	case strings.Contains(text, "auth login"), strings.Contains(text, "not logged"),
		strings.Contains(text, "bad credentials"), strings.Contains(text, "http 401"):
		return ReasonSinLogin
	case strings.Contains(text, "could not resolve to a repository"):
		return ReasonSinRemote
	}
	return ReasonRed
}

// Fetch hace la única consulta GraphQL del poll para owner/repo y la convierte.
func Fetch(ctx context.Context, r Runner, repo string, now time.Time) (*Snapshot, error) {
	owner, name, ok := strings.Cut(repo, "/")
	if !ok {
		return nil, fmt.Errorf("github: repo inválido %q", repo)
	}
	args := []string{"api", "graphql", "-f", "query=" + repoQuery, "-f", "owner=" + owner, "-f", "name=" + name}
	if err := ReadOnly(args); err != nil {
		return nil, err
	}
	out, err := r.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("github: consultar %s: %w", repo, err)
	}
	var resp gqlResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("github: respuesta de %s: %w", repo, err)
	}
	if resp.Data.Repository == nil {
		msgs := []string{"sin repository"}
		for _, e := range resp.Errors {
			msgs = append(msgs, e.Message)
		}
		return nil, fmt.Errorf("github: consultar %s: %s", repo, strings.Join(msgs, "; "))
	}
	return resp.Data.Repository.snapshot(repo, now), nil
}
