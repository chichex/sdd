// Package workspace mantiene en vivo el estado del visor: hace el escaneo
// inicial de forma progresiva, observa por polling de mtime los logs y
// artefactos de cada repo, rearma los trabajos y avisa a los suscriptores.
package workspace

import (
	"context"
	"maps"
	"sync"
	"time"

	"github.com/chichex/sdd/internal/github"
	"github.com/chichex/sdd/internal/model"
	"github.com/chichex/sdd/internal/scan"
)

// DefaultWatchInterval es el default de --watch-interval.
const DefaultWatchInterval = 2 * time.Second

// Config configura el servicio. ReadRepo es opcional: nil usa scan.Read.
type Config struct {
	Root           string
	WatchInterval  time.Duration
	ReadRepo       func(scan.Repo) scan.Data
	GitHub         github.Runner
	GitHubInterval time.Duration
	CacheDir       string
}

// GitHubStatus es el estado de la fuente GitHub; FetchedAt vacío es sin dato.
type GitHubStatus struct {
	OK        bool
	Reason    string
	FetchedAt string
}

// Snapshot es una foto consistente del estado: GitHub, RepoGitHub y Trabajos
// salen del mismo rebuild().
type Snapshot struct {
	Root       string
	ScannedAt  string
	Scanning   bool
	GitHub     GitHubStatus
	RepoGitHub map[string]GitHubStatus
	Repos      []scan.Data
	Trabajos   []model.Trabajo
}

// Service es el dueño del estado en vivo.
type Service struct {
	mu       sync.RWMutex
	cfg      Config
	now      func() time.Time
	repos    []scan.Repo
	data     map[string]scan.Data
	stamps   map[string]string
	scanning bool
	subs     map[chan struct{}]struct{}
	poller   *github.Poller
	gh       map[string]*github.Snapshot
	ghStatus map[string]repoStatus
	tags     map[string]model.Tags
	polled   bool

	// Vista derivada del estado de arriba: la reescribe rebuild() entera y
	// bajo un solo lock, y Snapshot() solo la lee. ghPerRepo no se muta una
	// vez guardado.
	trabajos  []model.Trabajo
	scannedAt string
	ghGlobal  GitHubStatus
	ghPerRepo map[string]GitHubStatus
}

// New crea el servicio; Run lo pone en marcha.
func New(cfg Config) *Service {
	if cfg.ReadRepo == nil {
		cfg.ReadRepo = scan.Read
	}
	if cfg.WatchInterval <= 0 {
		cfg.WatchInterval = DefaultWatchInterval
	}
	s := &Service{cfg: cfg, now: time.Now, data: map[string]scan.Data{}, stamps: map[string]string{},
		scanning: true, subs: map[chan struct{}]struct{}{}, gh: map[string]*github.Snapshot{},
		ghStatus: map[string]repoStatus{}, tags: map[string]model.Tags{}}
	if cfg.GitHub != nil {
		s.poller = &github.Poller{Runner: cfg.GitHub, Interval: cfg.GitHubInterval, Repos: s.githubRepos, OnResult: s.applyGitHub}
	}
	s.ghGlobal, s.ghPerRepo = s.githubStatus()
	return s
}

// Run hace el escaneo inicial y después observa los repos hasta que ctx se
// cancela. Es un proceso long-lived: su dueño es quien cancela ctx.
func (s *Service) Run(ctx context.Context) error {
	s.scanAll(ctx)
	var wg sync.WaitGroup
	defer wg.Wait()
	if s.poller != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.poller.Run(ctx) // termina al cancelar ctx
		}()
	}
	ticker := time.NewTicker(s.cfg.WatchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.watchTick()
		}
	}
}

// scanAll descubre los repos y los lee uno por uno, publicando cada avance.
func (s *Service) scanAll(ctx context.Context) {
	repos, err := scan.Discover(ctx, s.cfg.Root)
	if err != nil {
		repos = nil // carpeta ilegible: workspace vacío, el server sigue sirviendo
	}
	s.mu.Lock()
	s.repos, s.scanning = repos, true
	s.ghGlobal, s.ghPerRepo = s.githubStatus() // todos los repos tienen estado desde el descubrimiento
	s.mu.Unlock()
	s.loadCaches()
	s.publish()
	for _, r := range repos {
		if ctx.Err() != nil {
			return
		}
		s.store(r, s.cfg.ReadRepo(r), scan.Stamp(r))
	}
	s.mu.Lock()
	s.scanning = false
	s.mu.Unlock()
	s.rebuild()
}

func (s *Service) store(r scan.Repo, d scan.Data, stamp string) {
	s.mu.Lock()
	s.data[r.ID], s.stamps[r.ID] = d, stamp
	s.mu.Unlock()
	s.rebuild()
}

// watchTick relee los repos cuya huella en disco cambió.
func (s *Service) watchTick() {
	s.mu.RLock()
	repos := s.repos
	s.mu.RUnlock()
	for _, r := range repos {
		stamp := scan.Stamp(r)
		s.mu.RLock()
		changed := stamp != s.stamps[r.ID]
		s.mu.RUnlock()
		if changed {
			s.store(r, s.cfg.ReadRepo(r), stamp)
		}
	}
}

// rebuild rearma los trabajos y el estado de GitHub con lo leído hasta ahora,
// en una sola sección crítica, y avisa.
func (s *Service) rebuild() {
	s.mu.Lock()
	inputs := make([]model.RepoInput, 0, len(s.repos))
	for _, r := range s.repos {
		if d, ok := s.data[r.ID]; ok {
			inputs = append(inputs, s.inputFor(d))
		}
	}
	s.trabajos = model.Build(inputs)
	s.ghGlobal, s.ghPerRepo = s.githubStatus()
	s.scannedAt = s.now().UTC().Format("2006-01-02T15:04:05.000Z")
	s.mu.Unlock()
	s.publish()
}

// Snapshot devuelve el estado actual. Los repos todavía sin leer aparecen
// con lo que se sabe del descubrimiento. El estado de GitHub y los trabajos
// son los que guardó el mismo rebuild(): no se recalcula nada acá.
func (s *Service) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	repos := make([]scan.Data, 0, len(s.repos))
	for _, r := range s.repos {
		d, ok := s.data[r.ID]
		if !ok {
			d = scan.Data{Repo: r}
		}
		repos = append(repos, d)
	}
	return Snapshot{
		Root:       s.cfg.Root,
		ScannedAt:  s.scannedAt,
		Scanning:   s.scanning,
		GitHub:     s.ghGlobal,
		RepoGitHub: maps.Clone(s.ghPerRepo), // el mapa guardado es de rebuild(), no del llamador
		Repos:      repos,
		Trabajos:   s.trabajos,
	}
}

// Subscribe devuelve un canal que recibe una señal (sin bloquear al emisor)
// cada vez que cambia el estado, y la función para darse de baja.
func (s *Service) Subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	return ch, func() {
		s.mu.Lock()
		delete(s.subs, ch)
		s.mu.Unlock()
	}
}

func (s *Service) publish() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.subs {
		select {
		case ch <- struct{}{}:
		default: // ya tiene un aviso pendiente
		}
	}
}
