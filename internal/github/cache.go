package github

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CachePath es `<dir>/sdd/<owner>_<repo>.json`.
func CachePath(dir, repo string) string {
	return filepath.Join(dir, "sdd", strings.ReplaceAll(repo, "/", "_")+".json")
}

// LoadCache lee el último snapshot cacheado de un repo.
func LoadCache(dir, repo string) (*Snapshot, error) {
	data, err := os.ReadFile(CachePath(dir, repo))
	if err != nil {
		return nil, fmt.Errorf("github: leer caché de %s: %w", repo, err)
	}
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("github: caché de %s: %w", repo, err)
	}
	return &s, nil
}

// SaveCache persiste el snapshot con escritura atómica (temporal + rename).
func SaveCache(dir string, s *Snapshot) error {
	path := CachePath(dir, s.Repo)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("github: crear caché: %w", err)
	}
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("github: serializar caché de %s: %w", s.Repo, err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return fmt.Errorf("github: caché temporal: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close() // ya hay un error que reportar
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("github: escribir caché de %s: %w", s.Repo, err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name()) // ya hay un error que reportar
		return fmt.Errorf("github: cerrar caché de %s: %w", s.Repo, err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("github: guardar caché de %s: %w", s.Repo, err)
	}
	return nil
}
