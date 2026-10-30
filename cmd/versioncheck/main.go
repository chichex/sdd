// Command versioncheck falla si VERSION no es mayor que el último tag v* del
// repo. Lo corren el job Version de CI en cada PR y release.yml antes de
// publicar. Sale 0 si la versión es publicable y 1 si no, diciendo qué falló.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/chichex/sdd/internal/semver"
)

const gitTimeout = 10 * time.Second

// errNotGreater: VERSION es igual o menor que el último tag publicado.
var errNotGreater = errors.New("no es mayor que")

func main() {
	os.Exit(run(context.Background(), ".", os.Stdout, os.Stderr))
}

// run lee VERSION y los tags v* del repo en dir, imprime el resultado y
// devuelve el exit code.
func run(ctx context.Context, dir string, stdout, stderr io.Writer) int {
	msg, err := verify(ctx, dir)
	if err != nil {
		fmt.Fprintln(stderr, "versioncheck:", err)
		return 1
	}
	fmt.Fprintln(stdout, msg)
	return 0
}

func verify(ctx context.Context, dir string) (string, error) {
	version, err := readVersion(dir)
	if err != nil {
		return "", err
	}
	tags, err := listTags(ctx, dir)
	if err != nil {
		return "", err
	}
	return check(version, tags)
}

// readVersion devuelve el contenido de VERSION sin el salto de línea final.
func readVersion(dir string) (string, error) {
	b, err := os.ReadFile(filepath.Join(dir, "VERSION"))
	if err != nil {
		return "", fmt.Errorf("leer VERSION: %w", err)
	}
	return strings.TrimSuffix(string(b), "\n"), nil
}

// listTags devuelve los tags v* del repo en dir (git tag --list 'v*').
func listTags(ctx context.Context, dir string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "-C", dir, "tag", "--list", "v*").Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("listar tags v*: %w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("listar tags v*: %w", err)
	}
	return strings.Fields(string(out)), nil
}

// check es el núcleo puro: VERSION tiene que ser semver válida y mayor que el
// último tag v<semver> de tags.
func check(version string, tags []string) (string, error) {
	v, err := semver.Parse(version)
	if err != nil {
		return "", fmt.Errorf("VERSION inválida: %w", err)
	}
	latest, ok := semver.Latest(tags)
	if !ok {
		return fmt.Sprintf("VERSION %s, sin tags previos", v), nil
	}
	if semver.Compare(v, latest) <= 0 {
		return "", fmt.Errorf("VERSION %s %w v%s, el último tag: subí VERSION en este PR", v, errNotGreater, latest)
	}
	return fmt.Sprintf("VERSION %s > último tag v%s", v, latest), nil
}
