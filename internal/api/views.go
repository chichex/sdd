package api

import (
	"github.com/chichex/sdd/internal/model"
	"github.com/chichex/sdd/internal/workspace"
)

type githubView struct {
	OK        bool    `json:"ok"`
	Reason    *string `json:"reason"`
	FetchedAt *string `json:"fetchedAt"`
}

type repoView struct {
	ID              string     `json:"id"`
	Nombre          string     `json:"nombre"`
	Rutas           []string   `json:"rutas"`
	SinGit          bool       `json:"sinGit"`
	Remote          *string    `json:"remote"`
	LineasInvalidas int        `json:"lineasInvalidas"`
	Errores         []string   `json:"errores"`
	GitHub          githubView `json:"github"`
}

type workspaceJSON struct {
	Root      string          `json:"root"`
	ScannedAt *string         `json:"scannedAt"`
	Scanning  bool            `json:"scanning"`
	GitHub    githubView      `json:"github"`
	Repos     []repoView      `json:"repos"`
	Trabajos  []model.Trabajo `json:"trabajos"`
}

// detalleJSON es el trabajo con su Gantt y sus links, que el listado omite.
type detalleJSON struct {
	model.Trabajo
	Gantt []model.Barra `json:"gantt"`
	Links model.Links   `json:"links"`
}

func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func workspaceView(snap workspace.Snapshot, trabajos []model.Trabajo) workspaceJSON {
	repos := make([]repoView, 0, len(snap.Repos))
	for _, d := range snap.Repos {
		errores := append(append([]string{}, d.Repo.Errores...), d.Errores...)
		gh := snap.RepoGitHub[d.Repo.ID]
		repos = append(repos, repoView{ID: d.Repo.ID, Nombre: d.Repo.Nombre, Rutas: d.Repo.Rutas, SinGit: d.Repo.SinGit,
			Remote: optional(d.Repo.Remote), LineasInvalidas: d.LineasInvalidas, Errores: errores,
			GitHub: githubView{OK: gh.OK, Reason: optional(gh.Reason), FetchedAt: optional(gh.FetchedAt)}})
	}
	return workspaceJSON{
		Root:      snap.Root,
		ScannedAt: optional(snap.ScannedAt),
		Scanning:  snap.Scanning,
		GitHub:    githubView{OK: snap.GitHub.OK, Reason: optional(snap.GitHub.Reason), FetchedAt: optional(snap.GitHub.FetchedAt)},
		Repos:     repos,
		Trabajos:  trabajos,
	}
}

func detalleView(tr model.Trabajo) detalleJSON {
	gantt := tr.Gantt
	if gantt == nil {
		gantt = []model.Barra{}
	}
	return detalleJSON{Trabajo: tr, Gantt: gantt, Links: tr.Links}
}
