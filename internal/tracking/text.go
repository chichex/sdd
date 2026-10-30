package tracking

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// accentFold mapea las letras latinas precompuestas a su base; junto con
// descartar las marcas combinantes reemplaza al NFD + strip de la referencia
// TS sin depender de golang.org/x/text.
var accentFold = func() map[rune]rune {
	groups := map[rune]string{
		'a': "àáâãäåāăą", 'c': "çćĉċč", 'd': "ď", 'e': "èéêëēĕėęě", 'g': "ĝğġģ",
		'h': "ĥ", 'i': "ìíîïĩīĭįı", 'j': "ĵ", 'k': "ķ", 'l': "ĺļľ", 'n': "ñńņňŉ",
		'o': "òóôõöōŏő", 'r': "ŕŗř", 's': "śŝşš", 't': "ţť", 'u': "ùúûüũūŭůűų",
		'w': "ŵ", 'y': "ýÿŷ", 'z': "źżž",
	}
	m := map[rune]rune{}
	for base, variants := range groups {
		for _, r := range variants {
			m[r] = base
			m[unicode.ToUpper(r)] = unicode.ToUpper(base)
		}
	}
	return m
}()

func foldRune(r rune) (rune, bool) {
	if unicode.Is(unicode.Mn, r) {
		return 0, false
	}
	if base, ok := accentFold[r]; ok {
		r = base
	}
	return unicode.ToLower(r), true
}

// normalizeLegacyText recorta, colapsa espacios, ignora mayúsculas y quita
// los acentos antes de comparar alias completos.
func normalizeLegacyText(value string) string {
	var b strings.Builder
	for _, r := range value {
		if folded, ok := foldRune(r); ok {
			b.WriteRune(folded)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

var supersededPrefix = regexp.MustCompile(`^\s*(?:reemplazada\s+por|superseded\s+by)\s+`)

// supersededRefFromRaw recupera la referencia que sigue a «reemplazada por» o
// «superseded by» leyéndola del texto crudo: el prefijo se compara plegado
// (tolera acentos y marcas sueltas) pero la referencia conserva sus bytes.
func supersededRefFromRaw(raw string) string {
	var folded []byte
	foldedToRaw := []int{0}
	for offset, r := range raw {
		end := offset + utf8.RuneLen(r)
		f, ok := foldRune(r)
		if !ok {
			foldedToRaw[len(foldedToRaw)-1] = end
			continue
		}
		folded = utf8.AppendRune(folded, f)
		for len(foldedToRaw) <= len(folded) {
			foldedToRaw = append(foldedToRaw, end)
		}
	}
	loc := supersededPrefix.FindIndex(folded)
	if loc == nil {
		return ""
	}
	return strings.TrimSpace(raw[foldedToRaw[loc[1]]:])
}

// fieldEnd devuelve dónde termina el valor de un campo que arranca en start:
// la primera posición donde empieza otro campo rotulado (`. Label:`), el
// cierre del comentario (`-->`, con un punto previo opcional si dotBeforeClose)
// o el fin de la línea. Reemplaza el lookahead lazy de la regex TS.
func fieldEnd(line string, start int, nextField, closing *regexp.Regexp) int {
	for p := start; p < len(line); p++ {
		rest := line[p:]
		if nextField.MatchString(rest) || closing.MatchString(rest) {
			return p
		}
	}
	return len(line)
}

var (
	stateLabel      = regexp.MustCompile(`(?i)(?:Estado|Status)\s*:\s*`)
	stateNextField  = regexp.MustCompile(`(?i)^\.\s+(?:Proyecto|Project|Fuente|Source|Grill|Estado|Status)\s*:`)
	stateClosing    = regexp.MustCompile(`^\.?\s*-->`)
	fieldNextField  = regexp.MustCompile(`(?i)^\.\s+(?:Estado|Status|Proyecto|Project|Fuente|Source|Grill)\s*:`)
	fieldClosing    = regexp.MustCompile(`^\s*-->`)
	trackingSplitAt = regexp.MustCompile(`(?i)^\s*[a-z][a-z-]*\s*=`)
)

// stateValues devuelve los valores crudos (recortados) de cada campo
// Estado:/Status: de la línea, en orden.
func stateValues(line string) []string {
	var out []string
	pos := 0
	for pos < len(line) {
		loc := stateLabel.FindStringIndex(line[pos:])
		if loc == nil {
			break
		}
		start := pos + loc[1]
		end := fieldEnd(line, start, stateNextField, stateClosing)
		if value := strings.TrimSpace(line[start:end]); value != "" {
			out = append(out, value)
		}
		pos = end
		if end == start {
			pos = start + 1
		}
	}
	return out
}

// labeledField devuelve el primer valor no vacío de alguno de los rótulos en
// los comentarios de una línea del preámbulo.
func labeledField(lines []string, labels ...string) string {
	label := regexp.MustCompile(`(?i)(?:` + strings.Join(labels, "|") + `)\s*:\s*`)
	for _, line := range lines {
		if !singleLineComment.MatchString(line) {
			continue
		}
		loc := label.FindStringIndex(line)
		if loc == nil {
			continue
		}
		end := fieldEnd(line, loc[1], fieldNextField, fieldClosing)
		if value := strings.TrimSpace(line[loc[1]:end]); value != "" {
			return value
		}
	}
	return ""
}

// splitTrackingFields parte el payload de un marker histórico en `;` solo
// cuando lo sigue el próximo `clave=`, así un valor con `;` no se trunca.
func splitTrackingFields(payload string) []string {
	if payload == "" {
		return nil
	}
	var parts []string
	last := 0
	for i := 0; i < len(payload); i++ {
		if payload[i] == ';' && trackingSplitAt.MatchString(payload[i+1:]) {
			parts = append(parts, payload[last:i])
			last = i + 1
		}
	}
	return append(parts, payload[last:])
}
