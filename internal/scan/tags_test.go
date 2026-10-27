package scan

import (
	"context"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.invalid", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false", "-C", dir}, args...)...)
	cmd.Env = append(cmd.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestTagsContaining_RepoTemporal(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git no está en PATH")
	}
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	git(t, dir, "commit", "-q", "--allow-empty", "-m", "merge del PR 1")
	merged := git(t, dir, "rev-parse", "HEAD")
	git(t, dir, "tag", "api/v1.0.57")
	git(t, dir, "tag", "web/v2.1.0")
	git(t, dir, "commit", "-q", "--allow-empty", "-m", "merge del PR 2")
	untagged := git(t, dir, "rev-parse", "HEAD")

	tags, found, err := TagsContaining(context.Background(), dir, merged)
	if err != nil || !found || !slices.Equal(tags, []string{"api/v1.0.57", "web/v2.1.0"}) {
		t.Fatalf("TagsContaining(merged) = %v, %v, %v; want los dos tags del monorepo", tags, found, err)
	}
	tags, found, err = TagsContaining(context.Background(), dir, untagged)
	if err != nil || !found || len(tags) != 0 {
		t.Fatalf("TagsContaining(untagged) = %v, %v, %v; want sin tags y encontrado", tags, found, err)
	}
	_, found, err = TagsContaining(context.Background(), dir, strings.Repeat("ab", 20))
	if err != nil || found {
		t.Fatalf("SHA ausente = found %v err %v; want no encontrado (sin dato local)", found, err)
	}
	if _, _, err := TagsContaining(context.Background(), dir, "--output=/tmp/x"); err == nil {
		t.Fatalf("un SHA inválido debería rechazarse antes de llamar a git")
	}
}
