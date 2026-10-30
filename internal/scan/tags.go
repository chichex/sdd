package scan

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var shaPattern = regexp.MustCompile(`^[0-9a-f]{7,64}$`)

// TagsContaining lista los tags locales que contienen el commit sha, sin
// fetch. found es false si el commit no existe en el clon (sin dato local).
func TagsContaining(ctx context.Context, dir, sha string) (tags []string, found bool, err error) {
	if !shaPattern.MatchString(sha) {
		return nil, false, fmt.Errorf("scan: sha inválido %q", sha)
	}
	if _, err := runGit(ctx, dir, "cat-file", "-e", sha+"^{commit}"); err != nil {
		return nil, false, nil // el merge commit no está en el clon
	}
	out, err := runGit(ctx, dir, "tag", "--contains", sha)
	if err != nil {
		return nil, true, err
	}
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			tags = append(tags, line)
		}
	}
	return tags, true, nil
}
