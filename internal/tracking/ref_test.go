package tracking

import "testing"

func TestRef_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		decoded string
		encoded string
	}{
		{name: "un issue relativo codifica el numeral", decoded: "#93", encoded: "%2393"},
		{name: "una ruta lógica codifica las barras", decoded: ".sdd/specs/x.md", encoded: ".sdd%2Fspecs%2Fx.md"},
		{name: "espacio, barra y no-ASCII van en bytes UTF-8", decoded: "session 9/á", encoded: "session%209%2F%C3%A1"},
		{name: "los bytes no reservados quedan literales", decoded: "2026-10-23-visor_flujo.sdd~v1", encoded: "2026-10-23-visor_flujo.sdd~v1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := EncodeRef(tc.decoded); got != tc.encoded {
				t.Fatalf("EncodeRef(%q) = %q, want %q", tc.decoded, got, tc.encoded)
			}
			got, ok := DecodeRef(tc.encoded)
			if !ok || got != tc.decoded {
				t.Fatalf("DecodeRef(%q) = %q, %v; want %q, true", tc.encoded, got, ok, tc.decoded)
			}
		})
	}
}

func TestDecodeRef_Invalidas(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
	}{
		{name: "hexadecimal en minúscula", encoded: "path%2fchild"},
		{name: "escape truncado al final", encoded: "abc%C"},
		{name: "secuencia UTF-8 truncada", encoded: "a%C3"},
		{name: "UTF-8 inválido", encoded: "%FF%FE"},
		{name: "no-ASCII crudo", encoded: "sesión"},
		{name: "barra cruda reservada", encoded: ".sdd/specs/x.md"},
		{name: "none codificado", encoded: "%6E%6F%6E%65"},
		{name: "byte no reservado codificado", encoded: "%41bc"},
		{name: "vacía", encoded: ""},
		{name: "el centinela none no es una referencia", encoded: "none"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got, ok := DecodeRef(tc.encoded); ok {
				t.Fatalf("DecodeRef(%q) = %q, true; want inválida", tc.encoded, got)
			}
		})
	}
}
