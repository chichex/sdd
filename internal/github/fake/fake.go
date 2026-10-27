// Package fake es un doble de `gh` para tests: sirve respuestas grabadas de
// testdata/gh/<repo>/graphql.json, cuenta invocaciones, inyecta fallas y
// marca el test como fallido ante una invocación no grabada o que escriba.
package fake

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// Runner implementa github.Runner sin red.
type Runner struct {
	// T recibe los errores de invocaciones no grabadas o prohibidas.
	T testing.TB
	// Dir es la carpeta de respuestas grabadas (testdata/gh).
	Dir string
	// Fail mapea el nombre de un repo al stderr con el que falla `gh`.
	Fail map[string]string
	// NotFound simula `gh` ausente del PATH.
	NotFound bool
	// Gate, si no es nil, hace esperar cada invocación hasta que se cierre.
	Gate chan struct{}

	mu    sync.Mutex
	calls [][]string
}

var forbidden = []string{"edit", "create", "merge", "close", "comment", "delete", "reopen"}

// readOnly es la verificación independiente del doble: ninguna invocación
// escribe en GitHub.
func readOnly(args []string) error {
	for i, a := range args {
		low := strings.ToLower(a)
		if strings.Contains(low, "mutation") {
			return fmt.Errorf("invocación con mutation: %q", args)
		}
		if (a == "-X" || a == "--method") && i+1 < len(args) && strings.ToUpper(args[i+1]) != "GET" {
			return fmt.Errorf("invocación con método %s: %q", args[i+1], args)
		}
		if strings.HasPrefix(a, "-X") && len(a) > 2 && strings.ToUpper(a[2:]) != "GET" {
			return fmt.Errorf("invocación con método %s: %q", a[2:], args)
		}
	}
	if len(args) > 1 {
		for _, f := range forbidden {
			if args[1] == f {
				return fmt.Errorf("subcomando de escritura: %q", args)
			}
		}
	}
	return nil
}

func value(args []string, key string) string {
	for _, a := range args {
		if v, ok := strings.CutPrefix(a, key+"="); ok {
			return v
		}
	}
	return ""
}

// Run atiende una invocación de `gh`.
func (f *Runner) Run(ctx context.Context, args ...string) ([]byte, error) {
	f.mu.Lock()
	f.calls = append(f.calls, append([]string{}, args...))
	f.mu.Unlock()
	if f.Gate != nil {
		select {
		case <-f.Gate:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if err := readOnly(args); err != nil {
		f.T.Errorf("fake gh: %v", err)
		return nil, err
	}
	if f.NotFound {
		return nil, fmt.Errorf("fake gh: %w", exec.ErrNotFound)
	}
	if len(args) < 2 || args[0] != "api" || args[1] != "graphql" {
		f.T.Errorf("fake gh: invocación no grabada: %q", args)
		return nil, errors.New("fake gh: invocación no grabada")
	}
	name := value(args, "name")
	if msg, ok := f.Fail[name]; ok {
		return nil, fmt.Errorf("fake gh: exit status 1: %s", msg)
	}
	data, err := os.ReadFile(filepath.Join(f.Dir, name, "graphql.json"))
	if err != nil {
		f.T.Errorf("fake gh: invocación no grabada para %q: %v", name, err)
		return nil, err
	}
	return data, nil
}

// Calls devuelve las invocaciones recibidas.
func (f *Runner) Calls() [][]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([][]string{}, f.calls...)
}

// CountFor cuenta las invocaciones para el repo de nombre name.
func (f *Runner) CountFor(name string) int {
	n := 0
	for _, c := range f.Calls() {
		if value(c, "name") == name {
			n++
		}
	}
	return n
}
