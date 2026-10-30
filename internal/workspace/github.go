package workspace

import (
	"context"
	"time"

	"github.com/chichex/sdd/internal/github"
	"github.com/chichex/sdd/internal/model"
	"github.com/chichex/sdd/internal/scan"
)

const tagsTimeout = 30 * time.Second

// repoStatus es el estado de GitHub de un repo tras el último poll.
type repoStatus struct {
	ok     bool
	reason string
}

// githubRepos son los repos con remote en github.com: los que se consultan.
func (s *Service) githubRepos() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []string
	for _, r := range s.repos {
		if id, ok := r.GitHubRepo(); ok && !r.SinGit {
			out = append(out, id)
		}
	}
	return out
}

// loadCaches sirve el último snapshot cacheado de cada repo mientras corre la
// primera consulta.
func (s *Service) loadCaches() {
	if s.cfg.CacheDir == "" || s.poller == nil {
		return
	}
	for _, id := range s.githubRepos() {
		if snap, err := github.LoadCache(s.cfg.CacheDir, id); err == nil {
			s.mu.Lock()
			s.gh[id] = snap
			s.mu.Unlock()
		}
	}
}

// applyGitHub guarda el resultado de un poll: el snapshot fresco (y su caché
// y sus tags locales) o el motivo del degradado, conservando lo anterior.
func (s *Service) applyGitHub(res github.Result) {
	for _, r := range res.Repos {
		if r.Snapshot == nil {
			s.mu.Lock()
			s.ghStatus[r.Repo] = repoStatus{reason: r.Reason}
			s.mu.Unlock()
			continue
		}
		if s.cfg.CacheDir != "" {
			_ = github.SaveCache(s.cfg.CacheDir, r.Snapshot) // sin caché solo se pierde el arranque rápido
		}
		tags := s.tagsFor(r.Repo, r.Snapshot)
		s.mu.Lock()
		s.gh[r.Repo], s.ghStatus[r.Repo], s.tags[r.Repo] = r.Snapshot, repoStatus{ok: true}, tags
		s.mu.Unlock()
	}
	s.mu.Lock()
	s.polled = true
	s.mu.Unlock()
	s.rebuild()
}

// tagsFor busca en el clon local los tags que contienen el merge commit de
// cada PR mergeado que el log del repo nombra.
func (s *Service) tagsFor(id string, snap *github.Snapshot) model.Tags {
	tags := model.Tags{BySHA: map[string][]string{}, Missing: map[string]bool{}}
	s.mu.RLock()
	d, ok := s.data[id]
	s.mu.RUnlock()
	if !ok || len(d.Repo.Rutas) == 0 {
		return tags
	}
	named := map[int]bool{}
	for _, ev := range d.Events {
		for _, n := range ev.PR {
			named[n] = true
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), tagsTimeout)
	defer cancel()
	for _, pr := range snap.PRs {
		if pr.MergeCommit == "" || !named[pr.Number] {
			continue
		}
		list, found, err := scan.TagsContaining(ctx, d.Repo.Rutas[0], pr.MergeCommit)
		switch {
		case err != nil:
		case !found:
			tags.Missing[pr.MergeCommit] = true
		default:
			tags.BySHA[pr.MergeCommit] = list
		}
	}
	return tags
}

// inputFor arma la entrada del modelo de un repo con lo que se sabe de GitHub.
// Antes del primer poll la caché se muestra como vigente; después, un repo
// sin dato fresco queda stale.
func (s *Service) inputFor(d scan.Data) model.RepoInput {
	in := model.RepoInput{Data: d, Stale: true}
	id, onGitHub := d.Repo.GitHubRepo()
	if s.poller == nil || !onGitHub {
		return in
	}
	in.GitHub, in.Tags = s.gh[id], s.tags[id]
	in.Stale = s.polled && !s.ghStatus[id].ok
	return in
}

// githubStatus resume el estado global y el de cada repo.
func (s *Service) githubStatus() (GitHubStatus, map[string]GitHubStatus) {
	perRepo := map[string]GitHubStatus{}
	if s.poller == nil {
		for _, r := range s.repos {
			perRepo[r.ID] = GitHubStatus{Reason: "desactivado"}
		}
		return GitHubStatus{Reason: "desactivado"}, perRepo
	}
	global := GitHubStatus{OK: true}
	queried, okCount, firstReason := 0, 0, ""
	for _, r := range s.repos {
		id, onGitHub := r.GitHubRepo()
		if !onGitHub || r.SinGit {
			perRepo[r.ID] = GitHubStatus{Reason: github.ReasonSinRemote}
			continue
		}
		queried++
		st := GitHubStatus{OK: !s.polled || s.ghStatus[id].ok, Reason: s.ghStatus[id].reason}
		if snap := s.gh[id]; snap != nil {
			st.FetchedAt = snap.FetchedAt
			global.FetchedAt = max(global.FetchedAt, snap.FetchedAt)
		}
		if st.OK {
			okCount++
		} else if firstReason == "" {
			firstReason = st.Reason
		}
		perRepo[r.ID] = st
	}
	switch {
	case len(s.repos) > 0 && queried == 0:
		global = GitHubStatus{Reason: github.ReasonSinRemote}
	case queried > 0 && okCount == 0:
		global.OK, global.Reason = false, firstReason
	}
	return global, perRepo
}

// Refresh relee lo local y consulta GitHub ya (o espera la consulta en
// vuelo); devuelve el fetchedAt resultante.
func (s *Service) Refresh(ctx context.Context) string {
	s.watchTick()
	if s.poller != nil {
		s.poller.Refresh(ctx)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	global, _ := s.githubStatus()
	return global.FetchedAt
}
