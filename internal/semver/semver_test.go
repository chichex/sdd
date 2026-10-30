package semver

import (
	"strings"
	"testing"
)

func TestParse_Validas(t *testing.T) {
	tests := []struct {
		in   string
		want Version
	}{
		{"0.1.0", Version{0, 1, 0}},
		{"1.2.3", Version{1, 2, 3}},
		{"0.10.0", Version{0, 10, 0}},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, err := Parse(tc.in)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v, want nil", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("Parse(%q) = %+v, want %+v", tc.in, got, tc.want)
			}
			if got.String() != tc.in {
				t.Fatalf("Parse(%q).String() = %q, want %q", tc.in, got.String(), tc.in)
			}
		})
	}
}

func TestParse_Invalidas(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"con prefijo v", "v1.2.3"},
		{"sin patch", "1.2"},
		{"con pre-release", "1.2.3-rc1"},
		{"con metadata de build", "1.2.3+build"},
		{"con cero a la izquierda", "01.2.3"},
		{"con espacio final", "1.2.3 "},
		{"vacía", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.in)
			if err == nil {
				t.Fatalf("Parse(%q) error = nil, want error", tc.in)
			}
			if !strings.Contains(err.Error(), `"`+tc.in+`"`) {
				t.Fatalf("Parse(%q) error = %q, want que nombre el valor entre comillas", tc.in, err)
			}
		})
	}
}

func TestCompare_OrdenNumerico(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"minor de dos dígitos le gana a uno", "0.10.0", "0.9.1", 1},
		{"major le gana a minor y patch altos", "1.0.0", "0.99.99", 1},
		{"patch menor", "1.2.3", "1.2.4", -1},
		{"iguales", "0.1.0", "0.1.0", 0},
		{"minor menor aunque el patch sea mayor", "2.0.9", "2.1.0", -1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a, errA := Parse(tc.a)
			b, errB := Parse(tc.b)
			if errA != nil || errB != nil {
				t.Fatalf("Parse(%q, %q): %v, %v", tc.a, tc.b, errA, errB)
			}
			if got := Compare(a, b); got != tc.want {
				t.Fatalf("Compare(%s, %s) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
			if got := Compare(b, a); got != -tc.want {
				t.Fatalf("Compare(%s, %s) = %d, want %d", tc.b, tc.a, got, -tc.want)
			}
		})
	}
}

func TestLatest_TagsV(t *testing.T) {
	tests := []struct {
		name   string
		tags   []string
		want   string
		wantOK bool
	}{
		{"mayor por orden numérico, no lexicográfico", []string{"v0.9.1", "v0.10.0", "v0.2.0"}, "0.10.0", true},
		{"ignora los tags que no son v<semver>", []string{"latest", "v1.2", "release-1", "v0.3.0", "v9.9.9-rc1", "2.0.0"}, "0.3.0", true},
		{"sin tags válidos es sin versión", []string{"latest", "v1.2", "release-1"}, "", false},
		{"sin tags es sin versión", nil, "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := Latest(tc.tags)
			if ok != tc.wantOK {
				t.Fatalf("Latest(%q) ok = %v, want %v", tc.tags, ok, tc.wantOK)
			}
			if ok && got.String() != tc.want {
				t.Fatalf("Latest(%q) = %s, want %s", tc.tags, got, tc.want)
			}
		})
	}
}
