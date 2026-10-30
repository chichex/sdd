package tracking

import (
	"regexp"
	"slices"
	"strings"
)

var (
	trackingMarker   = regexp.MustCompile(`(?i)^[ \t]*<!--\s*SDD-Tracking\s*:\s*(.*?)\s*-->[ \t]*$`)
	trackingField    = regexp.MustCompile(`(?i)^\s*([a-z][a-z-]*)\s*=\s*(.*?)\s*$`)
	specHeading      = regexp.MustCompile(`(?i)^#\s+spec(?:[\s:—–-]|$)`)
	grillHeading     = regexp.MustCompile(`(?i)^#\s+grill(?:[\s:—–-]|$)`)
	specGenerator    = regexp.MustCompile(`(?i)(?:generada|generated)[^\n>]*sdd-spec`)
	projectGenerator = regexp.MustCompile(`(?i)(?:generado|generated)[^\n>]*sdd-init`)
	projectLabel     = regexp.MustCompile(`(?i)(?:Proyecto|Project)\s*:`)
	sourceLabel      = regexp.MustCompile(`(?i)(?:Fuente|Source)\s*:`)
	generatedLine    = regexp.MustCompile(`(?i)(?:Generado|Generated)`)
	isoDate          = regexp.MustCompile(`\b[0-9]{4}-[0-9]{2}-[0-9]{2}\b`)
	sourceIssue      = regexp.MustCompile(`(?:[A-Za-z0-9._-]+/[A-Za-z0-9._-]+)?#[1-9][0-9]*`)
	h2Title          = regexp.MustCompile(`^##[ \t]+(.+?)[ \t]*$`)
	executionTitle   = regexp.MustCompile(`^resultado de ejecucion(?:\s*\(.*\))?$`)
)

type legacyState struct {
	raw, line, value, supersededBy string
}

func normalizeLegacyState(raw string) (value, supersededBy string) {
	n := normalizeLegacyText(raw)
	switch n {
	case "draft", "borrador":
		return "draft", ""
	case "aprobada", "approved", "pendiente de ejecucion":
		return "approved", ""
	case "implementada", "implemented":
		return "implemented", ""
	case "reemplazada", "superseded":
		return "superseded", ""
	case "paused", "finalized":
		return n, ""
	}
	for _, prefix := range []string{"reemplazada por ", "superseded by "} {
		if rest, ok := strings.CutPrefix(n, prefix); ok && rest != "" {
			if ref := supersededRefFromRaw(raw); ref != "" {
				return "superseded", ref
			}
			return "superseded", rest
		}
	}
	return "unknown", ""
}

func legacyStates(lines []string) []legacyState {
	var out []legacyState
	for _, line := range lines {
		if !singleLineComment.MatchString(line) {
			continue
		}
		for _, raw := range stateValues(line) {
			value, superseded := normalizeLegacyState(raw)
			out = append(out, legacyState{raw: raw, line: line, value: value, supersededBy: superseded})
		}
	}
	return out
}

// historicalTracking devuelve el primer marker SDD-Tracking histórico (sin
// version ni type) bien formado, con sus campos en minúscula.
func historicalTracking(lines []string) (string, map[string]string, bool) {
	for _, line := range lines {
		m := trackingMarker.FindStringSubmatch(line)
		if m == nil || declaresCanonical.MatchString(line) {
			continue
		}
		fields := map[string]string{}
		for _, part := range splitTrackingFields(m[1]) {
			if f := trackingField.FindStringSubmatch(part); f != nil {
				fields[strings.ToLower(f[1])] = f[2]
			}
		}
		return line, fields, true
	}
	return "", nil, false
}

func legacyRef(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "none" {
		return ""
	}
	return strings.TrimSuffix(value, ".")
}

// legacyIdentity deriva issue y grill con los fallbacks del contrato: el issue
// del marker histórico o del campo Fuente/Source; el grill del marker o del
// rótulo Grill:.
func legacyIdentity(lines []string, fields map[string]string) (issue, grill string) {
	tracked, hasIssue := fields["issue"]
	switch {
	case hasIssue && tracked == "none":
	case IsIssueRef(tracked):
		issue = tracked
	default:
		if m := sourceIssue.FindString(labeledField(lines, "Fuente", "Source")); IsIssueRef(m) {
			issue = m
		}
	}
	if g, ok := fields["grill"]; ok {
		return issue, legacyRef(g)
	}
	return issue, legacyRef(labeledField(lines, "Grill"))
}

type typeSignals struct{ spec, grill, project bool }

func (s typeSignals) single() (string, bool) {
	var found []string
	for _, c := range []struct {
		on   bool
		name string
	}{{s.spec, "spec"}, {s.grill, "grill"}, {s.project, "project"}} {
		if c.on {
			found = append(found, c.name)
		}
	}
	switch len(found) {
	case 0:
		return "unknown", true
	case 1:
		return found[0], true
	}
	return "", false
}

func detectType(lines, headings []string, preamble string, explicit *legacyState, hasTracking bool) typeSignals {
	var s typeSignals
	anyHeading := func(re *regexp.Regexp) bool { return slices.ContainsFunc(headings, re.MatchString) }
	projectHeading := slices.ContainsFunc(headings, func(h string) bool {
		n := normalizeLegacyText(h)
		return strings.HasPrefix(n, "# contrato de autonomia") || strings.HasPrefix(n, "# autonomy contract")
	})
	handoff := explicit != nil && (explicit.value == "paused" || explicit.value == "finalized") &&
		projectLabel.MatchString(preamble) && sourceLabel.MatchString(preamble)
	s.spec = anyHeading(specHeading) || specGenerator.MatchString(preamble) || hasTracking
	s.grill = anyHeading(grillHeading) || handoff
	s.project = projectHeading && projectGenerator.MatchString(preamble)
	if explicit != nil && slices.Contains(specStates, explicit.value) {
		s.spec = true
	}
	if explicit != nil && slices.Contains(grillStates, explicit.value) {
		s.grill = true
	}
	return s
}

func hasExecutionHeading(markdown string) bool {
	for _, line := range sourceLines(markdown) {
		if m := h2Title.FindStringSubmatch(line); m != nil && executionTitle.MatchString(normalizeLegacyText(m[1])) {
			return true
		}
	}
	return false
}

func parseLegacy(markdown string) Result {
	lines := preambleLines(markdown)
	preamble := strings.Join(lines, "\n")
	trackingRaw, fields, hasTracking := historicalTracking(lines)
	states := legacyStates(lines)
	distinct := map[string]bool{}
	for _, st := range states {
		key := st.value + ":" + st.supersededBy
		if st.value == "unknown" {
			key = "unknown:" + normalizeLegacyText(st.raw)
		}
		distinct[key] = true
	}
	if len(distinct) > 1 {
		raws := make([]string, len(states))
		for i, st := range states {
			raws[i] = st.line
		}
		return Result{Kind: KindConflict, Raw: raws, State: absentState(),
			Diagnostics: []Diagnostic{{Code: "legacy-state-conflict", Message: "legacy state fields disagree"}}}
	}
	var explicit *legacyState
	if len(states) > 0 {
		explicit = &states[0]
	}
	var headings []string
	for _, line := range lines {
		if isHeading1(line) {
			headings = append(headings, line)
		}
	}
	artifactType, ok := detectType(lines, headings, preamble, explicit, hasTracking).single()
	if !ok {
		return Result{Kind: KindConflict, Raw: headingsAndComments(lines), State: absentState(),
			Diagnostics: []Diagnostic{{Code: "legacy-type-conflict", Message: "legacy signals identify incompatible artifact types"}}}
	}
	state := absentState()
	if explicit != nil {
		state = State{Value: explicit.value, Provenance: "legacy-explicit", Raw: explicit.raw}
	} else if artifactType == "spec" && hasExecutionHeading(markdown) {
		state = State{Value: "implemented", Provenance: "legacy-inferred"}
	}
	if artifactType == "unknown" && explicit == nil && !hasTracking {
		return Result{Kind: KindAbsent, State: absentState()}
	}
	meta := Metadata{Type: artifactType, Project: legacyRef(labeledField(lines, "Proyecto", "Project"))}
	meta.Issue, meta.Grill = legacyIdentity(lines, fields)
	generated := ""
	if i := slices.IndexFunc(lines, generatedLine.MatchString); i >= 0 {
		generated = lines[i]
	}
	if artifactType == "project" {
		meta.GeneratedAt = isoDate.FindString(generated)
	}
	if explicit != nil {
		meta.SupersededBy = explicit.supersededBy
	}
	return Result{Kind: KindLegacy, Metadata: meta, Raw: []string{firstNonEmpty(trackingRaw, explicitLine(explicit), first(headings), generated, preamble)}, State: state}
}

func headingsAndComments(lines []string) []string {
	var out []string
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "<!--") {
			out = append(out, line)
		}
	}
	return out
}

func explicitLine(s *legacyState) string {
	if s == nil {
		return ""
	}
	return s.line
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
