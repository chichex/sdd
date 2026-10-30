package github

import (
	"context"
	"sync"
	"time"
)

// DefaultInterval es el default de --github-interval.
const DefaultInterval = 2 * time.Minute

// maxParallel acota las consultas simultáneas a GitHub dentro de un poll.
const maxParallel = 4

// Clock es el reloj del poller; en tests se reemplaza por uno manual.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type realClock struct{}

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// RepoResult es el resultado de un repo en un poll: un snapshot fresco o el
// error con su razón de degradado.
type RepoResult struct {
	Repo     string
	Snapshot *Snapshot
	Err      error
	Reason   string
}

// Result es un poll completo, con un resultado por repo en el orden pedido.
type Result struct {
	FetchedAt time.Time
	Repos     []RepoResult
}

type call struct {
	done chan struct{}
	res  Result
}

// Poller consulta GitHub para los repos que devuelve Repos: uno al arrancar,
// uno cada Interval y uno por cada Refresh. Los Refresh concurrentes se
// fusionan en la consulta que esté en vuelo.
type Poller struct {
	Runner   Runner
	Clock    Clock // opcional; nil usa el reloj real
	Interval time.Duration
	Repos    func() []string
	OnResult func(Result) // opcional; se llama antes de liberar a quienes esperan

	mu       sync.Mutex
	inflight *call
	// applyMu serializa OnResult: los resultados se aplican en el orden en que
	// terminaron sus consultas.
	applyMu sync.Mutex
}

func (p *Poller) clock() Clock {
	if p.Clock == nil {
		return realClock{}
	}
	return p.Clock
}

// Run consulta al arrancar y después cada Interval hasta que ctx se cancela.
// Es un proceso long-lived: su dueño es quien cancela ctx.
func (p *Poller) Run(ctx context.Context) {
	interval := p.Interval
	if interval <= 0 {
		interval = DefaultInterval
	}
	for {
		p.Refresh(ctx)
		select {
		case <-ctx.Done():
			return
		case <-p.clock().After(interval):
		}
	}
}

// Refresh consulta ya, o espera la consulta en vuelo y devuelve su resultado.
// Una consulta deja de estar en vuelo cuando termina de consultar: un pedido
// que llega mientras se aplican sus resultados dispara una consulta nueva.
func (p *Poller) Refresh(ctx context.Context) Result {
	p.mu.Lock()
	if c := p.inflight; c != nil {
		p.mu.Unlock()
		select {
		case <-c.done:
			return c.res
		case <-ctx.Done():
			return Result{}
		}
	}
	c := &call{done: make(chan struct{})}
	p.inflight = c
	p.mu.Unlock()

	c.res = p.poll(ctx)
	p.applyMu.Lock()
	p.mu.Lock()
	p.inflight = nil
	p.mu.Unlock()
	if p.OnResult != nil {
		p.OnResult(c.res)
	}
	p.applyMu.Unlock()
	close(c.done)
	return c.res
}

func (p *Poller) poll(ctx context.Context) Result {
	repos := p.Repos()
	res := Result{FetchedAt: p.clock().Now(), Repos: make([]RepoResult, len(repos))}
	sem := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup
	for i, repo := range repos {
		wg.Add(1)
		go func(i int, repo string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			snap, err := Fetch(ctx, p.Runner, repo, res.FetchedAt)
			r := RepoResult{Repo: repo, Snapshot: snap, Err: err}
			if err != nil {
				r.Reason = Reason(err)
			}
			res.Repos[i] = r
		}(i, repo)
	}
	wg.Wait()
	return res
}
