package model

import (
	"regexp"

	"github.com/chichex/visor/internal/scan"
)

// brandedSpec separa una ruta de spec en su marca (chiche o sdd) y el resto.
var brandedSpec = regexp.MustCompile(`^\.(chiche|sdd)/(specs/.+)$`)

// specBrands elige, para cada spec nombrada con las dos marcas, una sola ruta:
// la del artefacto en disco si existe; si no, `.chiche/` cuando algún evento
// la nombra así; si no, `.sdd/`. Así una spec migrada de `.sdd/` a `.chiche/`
// sigue siendo un solo trabajo y un repo sin chiche conserva sus identidades.
func specBrands(d scan.Data) map[string]string {
	out := map[string]string{}
	for _, ev := range d.Events {
		for _, ref := range []string{ev.Spec, ev.Previous} {
			if m := brandedSpec.FindStringSubmatch(ref); m != nil && (out[m[2]] == "" || m[1] == "chiche") {
				out[m[2]] = ref
			}
		}
	}
	for _, a := range d.Specs {
		if m := brandedSpec.FindStringSubmatch(a.Logical); m != nil {
			out[m[2]] = a.Logical
		}
	}
	return out
}

// canonSpec lleva una ruta de spec a la elegida por specBrands; cualquier otra
// referencia (un issue, una ruta sin marca) queda tal cual.
func (g *grouper) canonSpec(spec string) string {
	if m := brandedSpec.FindStringSubmatch(spec); m != nil {
		if ref, ok := g.brands[m[2]]; ok {
			return ref
		}
	}
	return spec
}
