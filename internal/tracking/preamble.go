package tracking

import (
	"regexp"
	"strings"
)

const bom = "\xEF\xBB\xBF"

var (
	fenceLine         = regexp.MustCompile("^[ \t]{0,3}(`{3,}|~{3,})")
	h2Line            = regexp.MustCompile(`^##(?:[ \t]+.*|[ \t]*)$`)
	singleLineComment = regexp.MustCompile(`^[ \t]*<!--.*-->[ \t]*$`)
	mentionsTracking  = regexp.MustCompile(`(?i)SDD-Tracking`)
	declaresCanonical = regexp.MustCompile(`(?i)(?:^|[;\s])(?:version|type)\s*=`)
)

// sourceLines parte el documento en líneas sin su fin de línea (LF o CRLF).
// Un BOM inicial no es parte de la primera línea: el preámbulo arranca después.
func sourceLines(markdown string) []string {
	markdown = strings.TrimPrefix(markdown, bom)
	if markdown == "" {
		return nil
	}
	lines := strings.Split(markdown, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for i, line := range lines {
		lines[i] = strings.TrimSuffix(line, "\r")
	}
	return lines
}

// preambleLines devuelve las líneas hasta el primer `##` que no esté dentro
// de un bloque de código delimitado por backticks o tildes.
func preambleLines(markdown string) []string {
	lines := sourceLines(markdown)
	var fence byte
	for i, line := range lines {
		if m := fenceLine.FindStringSubmatch(line); m != nil {
			char := m[1][0]
			switch fence {
			case 0:
				fence = char
			case char:
				fence = 0
			}
			continue
		}
		if fence == 0 && h2Line.MatchString(line) {
			return lines[:i]
		}
	}
	return lines
}

// markerLines son las líneas del preámbulo con forma de comentario de una
// línea que nombran SDD-Tracking, bien formadas o no: un intento canónico
// roto se reporta inválido en vez de caer a legacy.
func markerLines(markdown string) []string {
	var out []string
	for _, line := range preambleLines(markdown) {
		if singleLineComment.MatchString(line) && mentionsTracking.MatchString(line) {
			out = append(out, line)
		}
	}
	return out
}

func isHeading1(line string) bool {
	return line == "#" || strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "#\t")
}
