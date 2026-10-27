package github

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/chichex/sdd/internal/github/fake"
)

// fakeClock entrega un canal por cada After y deja que el test lo dispare.
type fakeClock struct {
	mu     sync.Mutex
	asked  []time.Duration
	timers []chan time.Time
}

func (c *fakeClock) Now() time.Time { return now }

func (c *fakeClock) After(d time.Duration) <-chan time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	ch := make(chan time.Time, 1)
	c.asked = append(c.asked, d)
	c.timers = append(c.timers, ch)
	return ch
}

func (c *fakeClock) fire(t *testing.T, i int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		c.mu.Lock()
		if len(c.timers) > i {
			c.timers[i] <- now
			c.mu.Unlock()
			return
		}
		c.mu.Unlock()
		if time.Now().After(deadline) {
			t.Fatalf("el poller no pidió el timer %d", i)
		}
		time.Sleep(time.Millisecond)
	}
}

func waitResult(t *testing.T, ch <-chan Result) Result {
	t.Helper()
	select {
	case r := <-ch:
		return r
	case <-time.After(5 * time.Second):
		t.Fatalf("el poll no terminó")
	}
	return Result{}
}

func TestPoller_IntervaloYUnaConsultaPorRepo(t *testing.T) {
	gh := &fake.Runner{T: t, Dir: "../../testdata/gh"}
	clock := &fakeClock{}
	results := make(chan Result, 4)
	p := &Poller{Runner: gh, Clock: clock, Interval: DefaultInterval,
		Repos:    func() []string { return []string{"chichex/waica", "pramaestudio/platform"} },
		OnResult: func(r Result) { results <- r }}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); p.Run(ctx) }()
	t.Cleanup(func() { cancel(); <-done })

	first := waitResult(t, results)
	if len(first.Repos) != 2 || first.Repos[0].Snapshot == nil || first.Repos[1].Err != nil {
		t.Fatalf("primer poll = %+v", first)
	}
	if gh.CountFor("waica") != 1 || gh.CountFor("platform") != 1 || len(gh.Calls()) != 2 {
		t.Fatalf("invocaciones del primer poll = %d, want una por repo", len(gh.Calls()))
	}
	clock.fire(t, 0)
	waitResult(t, results)
	if gh.CountFor("waica") != 2 || gh.CountFor("platform") != 2 {
		t.Fatalf("tras el intervalo: waica %d platform %d, want 2 y 2", gh.CountFor("waica"), gh.CountFor("platform"))
	}
	clock.mu.Lock()
	asked := clock.asked[0]
	clock.mu.Unlock()
	if asked != 2*time.Minute {
		t.Fatalf("intervalo pedido = %v, want 2m", asked)
	}
}

func TestPoller_RefreshConcurrenteSeFusiona(t *testing.T) {
	gate := make(chan struct{})
	gh := &fake.Runner{T: t, Dir: "../../testdata/gh", Gate: gate}
	p := &Poller{Runner: gh, Clock: &fakeClock{}, Interval: DefaultInterval,
		Repos: func() []string { return []string{"chichex/waica", "pramaestudio/platform"} }}
	var wg sync.WaitGroup
	out := make([]Result, 2)
	for i := range 2 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			out[i] = p.Refresh(context.Background())
		}(i)
		if i == 0 {
			deadline := time.Now().Add(5 * time.Second)
			for len(gh.Calls()) < 2 {
				if time.Now().After(deadline) {
					close(gate)
					t.Fatalf("el primer refresh no consultó los dos repos")
				}
				time.Sleep(time.Millisecond)
			}
		}
	}
	time.Sleep(50 * time.Millisecond) // el segundo pedido ya está esperando la consulta en vuelo
	close(gate)
	wg.Wait()
	if len(gh.Calls()) != 2 {
		t.Fatalf("invocaciones = %d, want 2 (una consulta en vuelo para los dos pedidos)", len(gh.Calls()))
	}
	if !out[0].FetchedAt.Equal(out[1].FetchedAt) || len(out[1].Repos) != 2 {
		t.Fatalf("los dos pedidos no recibieron el mismo resultado: %+v / %+v", out[0], out[1])
	}
}

func TestPoller_FallaDeUnRepoNoAfectaAOtros(t *testing.T) {
	gh := &fake.Runner{T: t, Dir: "../../testdata/gh", Fail: map[string]string{"waica": "dial tcp: i/o timeout"}}
	p := &Poller{Runner: gh, Clock: &fakeClock{}, Interval: DefaultInterval,
		Repos: func() []string { return []string{"chichex/waica", "pramaestudio/platform"} }}
	res := p.Refresh(context.Background())
	if len(res.Repos) != 2 || res.Repos[0].Reason != ReasonRed || res.Repos[0].Snapshot != nil || res.Repos[1].Snapshot == nil || res.Repos[1].Reason != "" {
		t.Fatalf("resultado = %+v, want waica red y platform con datos", res.Repos)
	}
}

func TestCache_RutaYRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if got := CachePath(dir, "chichex/waica"); got != filepath.Join(dir, "sdd", "chichex_waica.json") {
		t.Fatalf("CachePath = %q", got)
	}
	gh := &fake.Runner{T: t, Dir: "../../testdata/gh"}
	snap, err := Fetch(context.Background(), gh, "chichex/waica", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveCache(dir, snap); err != nil {
		t.Fatalf("SaveCache: %v", err)
	}
	back, err := LoadCache(dir, "chichex/waica")
	if err != nil || back.FetchedAt != snap.FetchedAt || len(back.PRs) != len(snap.PRs) || back.Issues[0].Title != snap.Issues[0].Title {
		t.Fatalf("LoadCache = %+v, %v", back, err)
	}
	if _, err := LoadCache(dir, "o/ausente"); err == nil {
		t.Fatalf("LoadCache de un repo sin caché no falló")
	}
}
