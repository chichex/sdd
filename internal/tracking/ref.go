// Package tracking porta a Go el contrato SDD-Tracking v1 (skills/docs/sdd-tracking-v1.md):
// la codificación de referencias, el parseo total de markers canónicos y
// legacy, y la serialización canónica. Es solo lectura: no hay upsert.
package tracking

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

const hexDigits = "0123456789ABCDEF"

func isUnreserved(b byte) bool {
	switch {
	case b >= 'A' && b <= 'Z', b >= 'a' && b <= 'z', b >= '0' && b <= '9':
		return true
	case b == '.', b == '_', b == '~', b == '-':
		return true
	}
	return false
}

// EncodeRef codifica una referencia con el `<ref>` de SDD-Tracking v1: los
// bytes UTF-8 de `[A-Za-z0-9._~-]` quedan literales y el resto va como %HH
// con hexadecimal en mayúscula.
func EncodeRef(value string) string {
	var b strings.Builder
	for i := 0; i < len(value); i++ {
		c := value[i]
		if isUnreserved(c) {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hexDigits[c>>4])
		b.WriteByte(hexDigits[c&0x0F])
	}
	return b.String()
}

func hexValue(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

// DecodeRef decodifica una referencia canónica. Devuelve false ante una
// referencia vacía, `none` o que decodifica a `none`, escapes en minúscula o
// truncados, bytes no reservados codificados, bytes reservados o no-ASCII
// crudos y UTF-8 inválido: toda forma no canónica es inválida.
func DecodeRef(value string) (string, bool) {
	if value == "" || value == "none" {
		return "", false
	}
	decoded := make([]byte, 0, len(value))
	for i := 0; i < len(value); i++ {
		c := value[i]
		if isUnreserved(c) {
			decoded = append(decoded, c)
			continue
		}
		if c != '%' || i+2 >= len(value) {
			return "", false
		}
		hi, okHi := hexValue(value[i+1])
		lo, okLo := hexValue(value[i+2])
		if !okHi || !okLo {
			return "", false
		}
		b := hi<<4 | lo
		if isUnreserved(b) {
			return "", false
		}
		decoded = append(decoded, b)
		i += 2
	}
	if !utf8.Valid(decoded) {
		return "", false
	}
	out := string(decoded)
	if !isWellFormedRef(out) || EncodeRef(out) != value {
		return "", false
	}
	return out, true
}

// isWellFormedRef: una referencia real no es vacía, no es `none` y es UTF-8
// válido (en Go los strings pueden traer bytes sueltos).
func isWellFormedRef(value string) bool {
	return value != "" && value != "none" && utf8.ValidString(value)
}

var issueRef = regexp.MustCompile(`^(?:#[1-9][0-9]*|[A-Za-z0-9._-]+/[A-Za-z0-9._-]+#[1-9][0-9]*)$`)

// IsIssueRef dice si value es `#NN` u `owner/repo#NN`.
func IsIssueRef(value string) bool { return issueRef.MatchString(value) }

// IsCalendarDate dice si value es una fecha gregoriana real `YYYY-MM-DD` con
// año entre 0001 y 9999.
func IsCalendarDate(value string) bool {
	if len(value) != 10 || value[4] != '-' || value[7] != '-' {
		return false
	}
	year, okY := digits(value[0:4])
	month, okM := digits(value[5:7])
	day, okD := digits(value[8:10])
	if !okY || !okM || !okD || year == 0 || month < 1 || month > 12 || day < 1 {
		return false
	}
	leap := year%4 == 0 && (year%100 != 0 || year%400 == 0)
	days := [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if leap {
		days[1] = 29
	}
	return day <= days[month-1]
}

func digits(s string) (int, bool) {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		n = n*10 + int(s[i]-'0')
	}
	return n, true
}
