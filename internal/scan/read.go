package scan

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/chichex/sdd/internal/events"
	"github.com/chichex/sdd/internal/tracking"
)

// EventsPath es la ubicación del log de un repo.
func EventsPath(r Repo) string { return filepath.Join(r.CommonDir, "sdd", "events.jsonl") }

// artifactDirs devuelve, para grills o specs, el directorio del árbol (del
// checkout principal) y el del common-dir, en ese orden de precedencia.
func artifactDirs(r Repo, kind string) []string {
	dirs := []string{filepath.Join(r.Rutas[0], ".sdd", kind+"s")}
	if r.CommonDir != "" {
		dirs = append(dirs, filepath.Join(r.CommonDir, "sdd", kind+"s"))
	}
	return dirs
}

// Read lee el log, los grills y las specs de un repo. Un directorio ausente
// es vacío; un archivo ilegible va a Errores sin frenar la lectura.
func Read(r Repo) Data {
	d := Data{Repo: r}
	if r.SinGit || len(r.Rutas) == 0 {
		return d
	}
	if f, err := os.Open(EventsPath(r)); err == nil {
		log, rerr := events.ReadLog(f)
		_ = f.Close() // solo lectura: un error al cerrar no invalida lo leído
		d.Events, d.LineasInvalidas = log.Events, log.Invalid+log.Unsupported
		if rerr != nil {
			d.Errores = append(d.Errores, rerr.Error())
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		d.Errores = append(d.Errores, fmt.Sprintf("scan: leer %s: %v", EventsPath(r), err))
	}
	d.Grills = d.readArtifacts(r, "grill")
	d.Specs = d.readArtifacts(r, "spec")
	return d
}

func (d *Data) readArtifacts(r Repo, kind string) []Artifact {
	seen := map[string]bool{}
	var out []Artifact
	for _, dir := range artifactDirs(r, kind) {
		paths, _ := filepath.Glob(filepath.Join(dir, "*.md")) // el patrón es fijo y válido
		slices.Sort(paths)
		for _, path := range paths {
			name := filepath.Base(path)
			if seen[name] {
				continue
			}
			seen[name] = true
			a, err := readArtifact(path, kind)
			if err != nil {
				d.Errores = append(d.Errores, err.Error())
				continue
			}
			out = append(out, a)
		}
	}
	return out
}

func readArtifact(path, kind string) (Artifact, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Artifact{}, fmt.Errorf("scan: leer %s: %w", path, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Artifact{}, fmt.Errorf("scan: leer %s: %w", path, err)
	}
	text := string(data)
	name := filepath.Base(path)
	return Artifact{
		Kind:     kind,
		Name:     name,
		Path:     path,
		Logical:  ".sdd/" + kind + "s/" + name,
		Title:    firstHeading(text),
		ModTime:  info.ModTime().UTC(),
		Tracking: tracking.Parse(text),
	}, nil
}

func firstHeading(text string) string {
	for _, line := range strings.Split(strings.TrimPrefix(text, "\xEF\xBB\xBF"), "\n") {
		if title, ok := strings.CutPrefix(strings.TrimRight(line, "\r"), "# "); ok {
			return strings.TrimSpace(title)
		}
	}
	return ""
}

// Stamp resume el estado en disco de las fuentes de un repo (tamaño y mtime
// del log y de cada artefacto, y de sus directorios): si cambia, hay que releer.
func Stamp(r Repo) string {
	if r.SinGit || len(r.Rutas) == 0 {
		return ""
	}
	var b strings.Builder
	add := func(path string) {
		if info, err := os.Stat(path); err == nil {
			fmt.Fprintf(&b, "%s:%d:%d;", path, info.Size(), info.ModTime().UnixNano())
		}
	}
	add(EventsPath(r))
	for _, kind := range []string{"grill", "spec"} {
		for _, dir := range artifactDirs(r, kind) {
			add(dir)
			paths, _ := filepath.Glob(filepath.Join(dir, "*.md")) // patrón fijo
			for _, p := range paths {
				add(p)
			}
		}
	}
	return b.String()
}
