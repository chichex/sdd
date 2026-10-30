package main

import (
	"os"
	"strings"
	"testing"

	"github.com/chichex/sdd/internal/semver"
)

// TestVersionFile valida el VERSION real de la raíz: una sola línea con
// MAJOR.MINOR.PATCH y salto de línea final. No fija el valor: cambia en cada PR.
func TestVersionFile(t *testing.T) {
	b, err := os.ReadFile("../../VERSION")
	if err != nil {
		t.Fatalf("leer VERSION de la raíz: %v", err)
	}
	content := string(b)
	if strings.Count(content, "\n") != 1 || !strings.HasSuffix(content, "\n") {
		t.Fatalf("VERSION = %q, want una sola línea con salto de línea final", content)
	}
	if _, err := semver.Parse(strings.TrimSuffix(content, "\n")); err != nil {
		t.Fatalf("VERSION: %v", err)
	}
}
