// Package model arma los trabajos del visor a partir de lo que se leyó de
// cada repo: agrupa eventos y artefactos por ancla, deriva etapas, ramas y
// sub-estados, y calcula el Gantt y los links del detalle. Es puro: no hace I/O.
package model

import (
	"github.com/chichex/sdd/internal/github"
	"github.com/chichex/sdd/internal/scan"
)

// Etapa es una de las siete etapas del flujo SDD.
type Etapa string

// Etapas del flujo, en orden.
const (
	Idea      Etapa = "idea"
	Grill     Etapa = "grill"
	Spec      Etapa = "spec"
	Run       Etapa = "run"
	PR        Etapa = "pr"
	Mergeado  Etapa = "mergeado"
	Publicado Etapa = "publicado"
)

// Ancla identifica un trabajo: un issue, un grill o una spec.
type Ancla struct {
	Tipo string `json:"tipo"`
	Ref  string `json:"ref"`
}

// Refs son las referencias de un nodo; un campo nil no aplica.
type Refs struct {
	Issue    *string `json:"issue"`
	Grill    *string `json:"grill"`
	Spec     *string `json:"spec"`
	PR       *int    `json:"pr"`
	MergedAt *string `json:"mergedAt,omitempty"`
}

// Nodo es un paso de una rama en una etapa.
type Nodo struct {
	Etapa     Etapa    `json:"etapa"`
	SubEstado string   `json:"subEstado"`
	TS        string   `json:"ts"`
	Refs      Refs     `json:"refs"`
	URL       *string  `json:"url"`
	Ruta      string   `json:"ruta,omitempty"`
	Capa      string   `json:"capa,omitempty"`
	Stale     bool     `json:"stale,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	// inicio es el primer evento de la etapa en su rama; vacío si el nodo no
	// sale de un evento (artefacto sin historia o etapa inferida por refs).
	inicio string
}

// Rama es una línea del grafo: la viva o una cerrada.
type Rama struct {
	ID     string `json:"id"`
	Estado string `json:"estado"`
	Nodos  []Nodo `json:"nodos"`
	// fin cierra una rama fallada o reemplazada en el Gantt.
	fin string
}

// Trabajo es la unidad raíz.
type Trabajo struct {
	ID              string  `json:"id"`
	Repo            string  `json:"repo"`
	Titulo          string  `json:"titulo"`
	Ancla           Ancla   `json:"ancla"`
	Etapa           Etapa   `json:"etapa"`
	SubEstado       string  `json:"subEstado"`
	UltimaActividad string  `json:"ultimaActividad"`
	Terminado       bool    `json:"terminado"`
	SinHistoria     bool    `json:"sinHistoria"`
	Ramas           []Rama  `json:"ramas"`
	Gantt           []Barra `json:"-"`
	Links           Links   `json:"-"`
}

// Barra es una fila del Gantt.
type Barra struct {
	Rama   string  `json:"rama"`
	Etapa  Etapa   `json:"etapa"`
	Inicio string  `json:"inicio"`
	Fin    *string `json:"fin"`
	Estado string  `json:"estado"`
}

// PRLink es un PR con su URL.
type PRLink struct {
	Numero int     `json:"numero"`
	URL    *string `json:"url"`
}

// Links del detalle.
type Links struct {
	Issue  *string  `json:"issue"`
	PRs    []PRLink `json:"prs"`
	Grills []string `json:"grills"`
	Specs  []string `json:"specs"`
}

// RepoInput es lo que Build recibe por repo.
type RepoInput struct {
	Data   scan.Data
	GitHub *github.Snapshot
	Stale  bool
	Tags   Tags
}

// Tags dice qué tags contienen cada merge commit.
type Tags struct {
	BySHA   map[string][]string
	Missing map[string]bool
}

func ptr[T any](v T) *T { return &v }
