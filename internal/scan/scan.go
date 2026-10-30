// Package scan descubre los repos SDD bajo una carpeta y lee sus fuentes
// locales: el log events-v1 del common-dir y los grills y specs del árbol y
// del common-dir. Solo lee: nunca escribe en los clones.
package scan

import (
	"time"

	"github.com/chichex/sdd/internal/events"
	"github.com/chichex/sdd/internal/tracking"
)

// Repo es un repo descubierto: una o más rutas (worktrees) que comparten
// common-dir. Una carpeta con .sdd/project.md sin git es SinGit y no se lee.
type Repo struct {
	ID        string
	Nombre    string
	Rutas     []string
	SinGit    bool
	Remote    string
	CommonDir string
	Errores   []string
}

// GitHubRepo devuelve owner/repo si el remote apunta a github.com.
func (r Repo) GitHubRepo() (string, bool) {
	if r.Remote == "" || !isGitHubRemote(r.Remote) {
		return "", false
	}
	return r.ID, true
}

// Artifact es un grill o una spec leídos del disco. Logical es la ruta lógica
// `.sdd/<kind>s/<name>`, la misma para la copia del árbol y la del common-dir.
type Artifact struct {
	Kind     string
	Name     string
	Path     string
	Logical  string
	Title    string
	ModTime  time.Time
	Tracking tracking.Result
}

// Data es lo que se leyó de un repo.
type Data struct {
	Repo            Repo
	Events          []events.Event
	LineasInvalidas int
	Grills          []Artifact
	Specs           []Artifact
	Errores         []string
}
