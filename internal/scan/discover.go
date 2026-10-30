package scan

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

// MaxDepth es la profundidad máxima bajo la carpeta raíz donde se buscan repos.
const MaxDepth = 3

var skipDirs = map[string]bool{"node_modules": true, ".git": true, "vendor": true, "dist": true, "build": true}

const gitTimeout = 10 * time.Second

func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		return "", fmt.Errorf("scan: git %s en %s: %w", strings.Join(args, " "), dir, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// candidates recorre root hasta MaxDepth y devuelve las carpetas con
// .sdd/project.md. No sigue symlinks ni entra en skipDirs.
func candidates(root string) []string {
	var found []string
	var walk func(dir string, depth int)
	walk = func(dir string, depth int) {
		if info, err := os.Stat(filepath.Join(dir, ".sdd", "project.md")); err == nil && info.Mode().IsRegular() {
			found = append(found, dir)
		}
		if depth == MaxDepth {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() && !skipDirs[e.Name()] {
				walk(filepath.Join(dir, e.Name()), depth+1)
			}
		}
	}
	walk(root, 0)
	return found
}

// Discover busca repos con .sdd/project.md hasta MaxDepth bajo root y agrupa
// los worktrees que comparten common-dir en un solo repo.
func Discover(ctx context.Context, root string) ([]Repo, error) {
	if _, err := os.Stat(root); err != nil {
		return nil, fmt.Errorf("scan: discover: %w", err)
	}
	byCommon := map[string]*Repo{}
	var repos []*Repo
	for _, dir := range candidates(root) {
		if _, err := os.Stat(filepath.Join(dir, ".git")); errors.Is(err, os.ErrNotExist) {
			name := filepath.Base(dir)
			repos = append(repos, &Repo{ID: name, Nombre: name, Rutas: []string{dir}, SinGit: true})
			continue
		}
		common, err := runGit(ctx, dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
		if err != nil {
			name := filepath.Base(dir)
			repos = append(repos, &Repo{ID: name, Nombre: name, Rutas: []string{dir}, Errores: []string{err.Error()}})
			continue
		}
		if r, ok := byCommon[common]; ok {
			r.Rutas = append(r.Rutas, dir)
			continue
		}
		remote, _ := runGit(ctx, dir, "config", "--get", "remote.origin.url") // sin remote: id por carpeta
		r := &Repo{Rutas: []string{dir}, Remote: remote, CommonDir: common}
		byCommon[common] = r
		repos = append(repos, r)
	}
	out := make([]Repo, 0, len(repos))
	for _, r := range repos {
		finishRepo(r)
		out = append(out, *r)
	}
	slices.SortFunc(out, func(a, b Repo) int { return strings.Compare(a.ID, b.ID) })
	return out, nil
}

// finishRepo ordena las rutas (el checkout principal, dueño del common-dir,
// primero) y deriva id y nombre.
func finishRepo(r *Repo) {
	if r.SinGit || r.CommonDir == "" {
		return
	}
	slices.SortFunc(r.Rutas, func(a, b string) int {
		am, bm := filepath.Join(a, ".git") == r.CommonDir, filepath.Join(b, ".git") == r.CommonDir
		switch {
		case am && !bm:
			return -1
		case bm && !am:
			return 1
		}
		return strings.Compare(a, b)
	})
	r.ID = RepoID(r.Remote)
	if r.ID == "" {
		r.ID = filepath.Base(r.Rutas[0])
	}
	r.Nombre = r.ID[strings.LastIndex(r.ID, "/")+1:]
}

var remotePattern = regexp.MustCompile(`^(?:[a-z+]+://)?(?:[^@/]+@)?([^/:]+)(?::[0-9]+)?[:/]([^/]+)/([^/]+?)(?:\.git)?/?$`)

// RepoID deriva owner/repo de un remote https, `git@host:` o `ssh://`.
func RepoID(remote string) string {
	m := remotePattern.FindStringSubmatch(strings.TrimSpace(remote))
	if m == nil {
		return ""
	}
	return m[2] + "/" + m[3]
}

func isGitHubRemote(remote string) bool {
	m := remotePattern.FindStringSubmatch(strings.TrimSpace(remote))
	return m != nil && strings.EqualFold(m[1], "github.com")
}
