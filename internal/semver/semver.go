// Package semver parsea y compara versiones MAJOR.MINOR.PATCH numéricas, sin
// prefijo, pre-release ni metadata de build.
package semver

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrInvalid indica que el valor no es MAJOR.MINOR.PATCH numérico.
var ErrInvalid = errors.New("no es MAJOR.MINOR.PATCH numérico")

// Version es una versión MAJOR.MINOR.PATCH.
type Version struct {
	Major, Minor, Patch int
}

// String devuelve la versión como MAJOR.MINOR.PATCH, sin prefijo.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Parse acepta solo MAJOR.MINOR.PATCH numéricos, sin ceros a la izquierda,
// prefijo, sufijos ni espacios. El error nombra el valor rechazado.
func Parse(s string) (Version, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("semver: %q %w", s, ErrInvalid)
	}
	var nums [3]int
	for i, p := range parts {
		n, ok := number(p)
		if !ok {
			return Version{}, fmt.Errorf("semver: %q %w", s, ErrInvalid)
		}
		nums[i] = n
	}
	return Version{Major: nums[0], Minor: nums[1], Patch: nums[2]}, nil
}

// number acepta un entero decimal no negativo sin signo ni ceros a la izquierda.
func number(p string) (int, bool) {
	if p == "" || (len(p) > 1 && p[0] == '0') {
		return 0, false
	}
	for _, c := range p {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(p)
	return n, err == nil
}

// Compare devuelve -1, 0 o 1 según a sea menor, igual o mayor que b,
// comparando numéricamente major, minor y patch en ese orden.
func Compare(a, b Version) int {
	if c := cmp.Compare(a.Major, b.Major); c != 0 {
		return c
	}
	if c := cmp.Compare(a.Minor, b.Minor); c != 0 {
		return c
	}
	return cmp.Compare(a.Patch, b.Patch)
}

// Latest devuelve la mayor versión entre los tags con forma v<semver> e ignora
// el resto. Sin ningún tag válido devuelve ok en false: «sin versión».
func Latest(tags []string) (Version, bool) {
	var latest Version
	found := false
	for _, tag := range tags {
		rest, isV := strings.CutPrefix(tag, "v")
		if !isV {
			continue
		}
		v, err := Parse(rest)
		if err != nil {
			continue // no es v<semver>: no cuenta como versión publicada
		}
		if !found || Compare(v, latest) > 0 {
			latest, found = v, true
		}
	}
	return latest, found
}
