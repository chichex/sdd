package tracking

import (
	"regexp"
	"slices"
	"strings"
)

var (
	canonicalWrapper = regexp.MustCompile(`^<!-- SDD-Tracking: (.+) -->$`)
	canonicalPair    = regexp.MustCompile(`^([a-z][a-z-]*)=([^;=\s]+)$`)
)

var schemas = map[string][]string{
	"spec":    {"version", "type", "state", "issue", "grill", "superseded-by"},
	"grill":   {"version", "type", "state", "issue", "grill", "project"},
	"project": {"version", "type", "generated-at"},
}

// parseCanonical valida un marker canónico byte a byte: sintaxis, claves
// (únicas, completas y en orden), valores y que su serialización reproduzca
// exactamente los bytes leídos.
func parseCanonical(raw string) (Metadata, *Error) {
	wrapper := canonicalWrapper.FindStringSubmatch(raw)
	if wrapper == nil {
		return Metadata{}, diag("invalid-marker-syntax", "canonical marker wrapper is invalid")
	}
	var keys []string
	values := map[string]string{}
	for _, pair := range strings.Split(wrapper[1], "; ") {
		m := canonicalPair.FindStringSubmatch(pair)
		if m == nil {
			return Metadata{}, diag("invalid-marker-syntax", "canonical field syntax is invalid")
		}
		if _, dup := values[m[1]]; dup {
			return Metadata{}, diag("duplicate-key", "repeated canonical key: %s", m[1])
		}
		keys = append(keys, m[1])
		values[m[1]] = m[2]
	}
	version, ok := values["version"]
	if !ok {
		return Metadata{}, diag("missing-key", "missing canonical key: version")
	}
	if version != "1" {
		return Metadata{}, diag("unsupported-version", "unsupported version: %s", version)
	}
	expected, ok := schemas[values["type"]]
	if !ok {
		return Metadata{}, diag("unknown-type", "unknown artifact type: %s", values["type"])
	}
	if err := checkSchema(keys, values, expected); err != nil {
		return Metadata{}, err
	}
	meta, err := canonicalValues(values)
	if err != nil {
		return Metadata{}, err
	}
	serialized, serr := serialize(meta)
	if serr != nil {
		return Metadata{}, serr
	}
	if serialized != raw {
		return Metadata{}, diag("invalid-marker-syntax", "canonical marker bytes are not deterministic")
	}
	return meta, nil
}

func checkSchema(keys []string, values map[string]string, expected []string) *Error {
	for _, k := range expected {
		if _, ok := values[k]; !ok {
			return diag("missing-key", "missing canonical key: %s", k)
		}
	}
	for _, k := range keys {
		if !slices.Contains(expected, k) {
			return diag("extra-key", "unexpected canonical key: %s", k)
		}
	}
	if !slices.Equal(keys, expected) {
		return diag("key-order", "canonical keys are out of order")
	}
	return nil
}

// optionalDecoded decodifica una referencia opcional: `none` es vacío.
func optionalDecoded(value string) (string, bool) {
	if value == "none" {
		return "", true
	}
	return DecodeRef(value)
}

func canonicalValues(values map[string]string) (Metadata, *Error) {
	meta := Metadata{Version: 1, Type: values["type"]}
	if meta.Type == "project" {
		meta.GeneratedAt = values["generated-at"]
		return meta, nil
	}
	issue := values["issue"]
	if issue != "none" && !IsIssueRef(issue) {
		return meta, diag("invalid-issue", "invalid issue reference: %s", issue)
	}
	if issue != "none" {
		meta.Issue = issue
	}
	meta.State = values["state"]
	grill, ok := optionalDecoded(values["grill"])
	if !ok {
		return meta, diag("invalid-encoding", "invalid grill encoding: %s", values["grill"])
	}
	meta.Grill = grill
	if meta.Type == "spec" {
		if meta.SupersededBy, ok = optionalDecoded(values["superseded-by"]); !ok {
			return meta, diag("invalid-encoding", "invalid superseded-by encoding: %s", values["superseded-by"])
		}
		return meta, nil
	}
	project, ok := DecodeRef(values["project"])
	if grill == "" || !ok {
		return meta, diag("invalid-encoding", "required grill/project reference is invalid")
	}
	meta.Project = project
	return meta, nil
}
