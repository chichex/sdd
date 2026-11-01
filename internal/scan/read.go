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

// brands son las dos marcas de la suite, en orden de precedencia: chiche (la
// actual) gana sobre sdd (la anterior, que se sigue leyendo).
var brands = []string{"chiche", "sdd"}

// EventsPaths son los logs de un repo: `<common>/sdd/events.jsonl` primero,
// así un empate de ts deja lo anterior antes que lo nuevo.
func EventsPaths(r Repo) []string {
	return []string{filepath.Join(r.CommonDir, "sdd", "events.jsonl"), filepath.Join(r.CommonDir, "chiche", "events.jsonl")}
}

// artifactDir es un directorio de grills o specs y la marca de su ruta lógica.
type artifactDir struct{ path, brand string }

// artifactDirs devuelve, para grills o specs, los cuatro directorios en orden
// de precedencia: `.chiche/` y `.sdd/` del árbol (del checkout principal) y
// después `chiche/` y `sdd/` del common-dir.
func artifactDirs(r Repo, kind string) []artifactDir {
	var dirs []artifactDir
	for _, b := range brands {
		dirs = append(dirs, artifactDir{filepath.Join(r.Rutas[0], "."+b, kind+"s"), b})
	}
	if r.CommonDir != "" {
		for _, b := range brands {
			dirs = append(dirs, artifactDir{filepath.Join(r.CommonDir, b, kind+"s"), b})
		}
	}
	return dirs
}

// Read lee los logs, los grills y las specs de un repo. Un directorio ausente
// es vacío; un archivo ilegible va a Errores sin frenar la lectura.
func Read(r Repo) Data {
	d := Data{Repo: r}
	if r.SinGit || len(r.Rutas) == 0 {
		return d
	}
	logs := 0
	for _, path := range EventsPaths(r) {
		if d.readLog(path) {
			logs++
		}
	}
	// Con eventos en los dos logs quedan en un solo orden por ts; el orden
	// estable conserva el de cada log y, ante un empate, sdd antes que chiche.
	// Un solo log queda en su orden.
	if logs > 1 {
		slices.SortStableFunc(d.Events, func(a, b events.Event) int { return strings.Compare(a.TS, b.TS) })
	}
	d.Grills = d.readArtifacts(r, "grill")
	d.Specs = d.readArtifacts(r, "spec")
	return d
}

// readLog suma los eventos de un log y dice si aportó alguno.
func (d *Data) readLog(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			d.Errores = append(d.Errores, fmt.Sprintf("scan: leer %s: %v", path, err))
		}
		return false
	}
	log, rerr := events.ReadLog(f)
	_ = f.Close() // solo lectura: un error al cerrar no invalida lo leído
	d.Events = append(d.Events, log.Events...)
	d.LineasInvalidas += log.Invalid + log.Unsupported
	if rerr != nil {
		d.Errores = append(d.Errores, rerr.Error())
	}
	return len(log.Events) > 0
}

func (d *Data) readArtifacts(r Repo, kind string) []Artifact {
	seen := map[string]bool{}
	var out []Artifact
	for _, dir := range artifactDirs(r, kind) {
		paths, _ := filepath.Glob(filepath.Join(dir.path, "*.md")) // el patrón es fijo y válido
		slices.Sort(paths)
		for _, path := range paths {
			name := filepath.Base(path)
			if seen[name] {
				continue
			}
			seen[name] = true
			a, err := readArtifact(path, kind, dir.brand)
			if err != nil {
				d.Errores = append(d.Errores, err.Error())
				continue
			}
			out = append(out, a)
		}
	}
	return out
}

func readArtifact(path, kind, brand string) (Artifact, error) {
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
		Logical:  "." + brand + "/" + kind + "s/" + name,
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
	for _, path := range EventsPaths(r) {
		add(path)
	}
	for _, kind := range []string{"grill", "spec"} {
		for _, dir := range artifactDirs(r, kind) {
			add(dir.path)
			paths, _ := filepath.Glob(filepath.Join(dir.path, "*.md")) // patrón fijo
			for _, p := range paths {
				add(p)
			}
		}
	}
	return b.String()
}
