package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/chichex/sdd/internal/semver"
)

func TestCheck_VersionContraTags(t *testing.T) {
	tests := []struct {
		name    string
		version string
		tags    []string
		wantMsg string
		wantErr error
		errHas  string
	}{
		{name: "mayor que el último tag", version: "0.2.0", tags: []string{"v0.1.0", "v0.1.1"}, wantMsg: "VERSION 0.2.0 > último tag v0.1.1"},
		{name: "sin tags previos", version: "0.1.0", wantMsg: "VERSION 0.1.0, sin tags previos"},
		{name: "solo tags que no son v<semver>", version: "0.1.0", tags: []string{"latest", "v1.2", "release-1"}, wantMsg: "VERSION 0.1.0, sin tags previos"},
		{name: "igual al último tag", version: "0.1.0", tags: []string{"v0.1.0"}, wantErr: errNotGreater, errHas: "0.1.0 no es mayor que v0.1.0"},
		{name: "menor que el último tag", version: "0.9.0", tags: []string{"v0.10.0"}, wantErr: errNotGreater, errHas: "0.9.0 no es mayor que v0.10.0"},
		{name: "VERSION inválida", version: "v0.2.0", wantErr: semver.ErrInvalid, errHas: `"v0.2.0"`},
		{name: "VERSION vacía", version: "", tags: []string{"v0.1.0"}, wantErr: semver.ErrInvalid, errHas: "VERSION"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := check(tc.version, tc.tags)
			if tc.wantErr == nil {
				if err != nil || msg != tc.wantMsg {
					t.Fatalf("check(%q, %q) = %q, %v; want %q, nil", tc.version, tc.tags, msg, err, tc.wantMsg)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("check(%q, %q) error = %v; want errors.Is %v", tc.version, tc.tags, err, tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.errHas) {
				t.Fatalf("check(%q, %q) error = %q; want que diga %q", tc.version, tc.tags, err, tc.errHas)
			}
		})
	}
}

// git corre git sobre dir con identidad y config aisladas de la máquina.
func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.invalid", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false", "-C", dir}, args...)...)
	cmd.Env = append(cmd.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// repoConTags arma un repo git en TempDir con tags v<semver> y de otra forma.
func repoConTags(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	git(t, dir, "commit", "-q", "--allow-empty", "-m", "primero")
	for _, tag := range []string{"v0.1.0", "latest", "release-1", "v0.10.0", "v0.9.1"} {
		git(t, dir, "tag", tag)
	}
	return dir
}

func TestListTags_RepoReal(t *testing.T) {
	dir := repoConTags(t)

	tags, err := listTags(context.Background(), dir)
	if err != nil {
		t.Fatalf("listTags: %v", err)
	}
	slices.Sort(tags)
	if want := []string{"v0.1.0", "v0.10.0", "v0.9.1"}; !slices.Equal(tags, want) {
		t.Fatalf("listTags = %q, want %q", tags, want)
	}
}

func TestRun_RepoReal(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		wantCode int
		stdout   string
		stderr   string
	}{
		{name: "mayor que el último tag", version: "0.11.0\n", wantCode: 0, stdout: "VERSION 0.11.0 > último tag v0.10.0\n"},
		{name: "igual al último tag", version: "0.10.0\n", wantCode: 1, stderr: "0.10.0 no es mayor que v0.10.0"},
		{name: "VERSION inválida", version: "x\n", wantCode: 1, stderr: `"x"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := repoConTags(t)
			if err := os.WriteFile(filepath.Join(dir, "VERSION"), []byte(tc.version), 0o644); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), dir, &stdout, &stderr)
			if code != tc.wantCode {
				t.Fatalf("run = %d, want %d (stdout %q, stderr %q)", code, tc.wantCode, stdout.String(), stderr.String())
			}
			if tc.stdout != "" && stdout.String() != tc.stdout {
				t.Fatalf("stdout = %q, want %q", stdout.String(), tc.stdout)
			}
			if tc.stderr != "" && !strings.Contains(stderr.String(), tc.stderr) {
				t.Fatalf("stderr = %q, want que diga %q", stderr.String(), tc.stderr)
			}
		})
	}
}

func TestRun_SinArchivoVersion(t *testing.T) {
	dir := repoConTags(t)
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), dir, &stdout, &stderr); code != 1 {
		t.Fatalf("run sin VERSION = %d, want 1 (stdout %q)", code, stdout.String())
	}
	if !strings.Contains(stderr.String(), "VERSION") {
		t.Fatalf("stderr = %q, want que nombre el archivo VERSION", stderr.String())
	}
}
