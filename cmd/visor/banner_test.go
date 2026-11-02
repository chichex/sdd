package main

import (
	"bytes"
	"errors"
	"net"
	"strings"
	"testing"
	"unicode/utf8"
)

func addrs(cidrs ...string) func() ([]net.Addr, error) {
	return func() ([]net.Addr, error) {
		var out []net.Addr
		for _, c := range cidrs {
			ip, n, err := net.ParseCIDR(c)
			if err != nil {
				return nil, err
			}
			n.IP = ip
			out = append(out, n)
		}
		return out, nil
	}
}

func TestBanner_LAN(t *testing.T) {
	var out bytes.Buffer
	banner(&out, bannerInfo{Host: "0.0.0.0", Port: 4321, Root: "/ws", Version: "1.2.3"}, addrs("127.0.0.1/8", "::1/128", "fe80::1/64", "192.168.1.23/24", "10.0.0.5/8"))
	text := out.String()
	if !strings.Contains(text, "abrí http://192.168.1.23:4321\n") {
		t.Fatalf("banner sin la URL LAN de la primera IPv4 no loopback:\n%s", text)
	}
	if !strings.Contains(text, "sin autenticación") {
		t.Fatalf("banner sin el aviso de autenticación:\n%s", text)
	}
	qr := qrLines(text)
	if len(qr) < 10 {
		t.Fatalf("banner sin QR (%d líneas de bloques):\n%s", len(qr), text)
	}
	for _, line := range qr {
		if w := utf8.RuneCountInString(line); w > 33 {
			t.Fatalf("QR de %d módulos de ancho, want ≤ 33:\n%s", w, text)
		}
	}
}

func TestBanner_Local(t *testing.T) {
	var out bytes.Buffer
	banner(&out, bannerInfo{Host: "127.0.0.1", Port: 4321, Root: "/ws", Version: "1.2.3"}, addrs("192.168.1.23/24"))
	text := out.String()
	if !strings.Contains(text, "abrí http://127.0.0.1:4321\n") || strings.Contains(text, "192.168") {
		t.Fatalf("banner local inesperado:\n%s", text)
	}
	if len(qrLines(text)) != 0 || strings.Contains(text, "sin autenticación") {
		t.Fatalf("el banner local no lleva QR ni aviso:\n%s", text)
	}
}

func TestBanner_Tailscale(t *testing.T) {
	var out bytes.Buffer
	banner(&out, bannerInfo{Host: "0.0.0.0", Port: 4321, Root: "/ws", Version: "1.2.3", TailscaleDNS: "chichex-linux.tail8ca4e8.ts.net"}, addrs("127.0.0.1/8", "100.110.35.128/32", "192.168.1.23/24"))
	text := out.String()
	if !strings.Contains(text, "abrí http://192.168.1.23:4321\n") {
		t.Fatalf("la URL principal tiene que ser la LAN, no la de Tailscale:\n%s", text)
	}
	if !strings.Contains(text, "tailscale: http://chichex-linux.tail8ca4e8.ts.net:4321\n") {
		t.Fatalf("banner sin la URL MagicDNS de Tailscale:\n%s", text)
	}
}

func TestBanner_TailscaleSinDNS(t *testing.T) {
	var out bytes.Buffer
	banner(&out, bannerInfo{Host: "0.0.0.0", Port: 4321, Root: "/ws", Version: "1.2.3"}, addrs("192.168.1.23/24", "100.110.35.128/32"))
	text := out.String()
	if !strings.Contains(text, "abrí http://192.168.1.23:4321\n") || !strings.Contains(text, "tailscale: http://100.110.35.128:4321\n") {
		t.Fatalf("sin MagicDNS la línea de Tailscale lleva la IP 100.x:\n%s", text)
	}
}

func TestBanner_SoloTailscale(t *testing.T) {
	var out bytes.Buffer
	banner(&out, bannerInfo{Host: "0.0.0.0", Port: 4321, Root: "/ws", Version: "1.2.3"}, addrs("127.0.0.1/8", "100.110.35.128/32"))
	text := out.String()
	if !strings.Contains(text, "abrí http://100.110.35.128:4321\n") || strings.Count(text, "100.110.35.128") != 1 {
		t.Fatalf("sin LAN la URL principal es la de Tailscale, sin repetirla:\n%s", text)
	}
}

func TestBanner_SinTailscale(t *testing.T) {
	var out bytes.Buffer
	banner(&out, bannerInfo{Host: "0.0.0.0", Port: 4321, Root: "/ws", Version: "1.2.3"}, addrs("192.168.1.23/24"))
	if text := out.String(); strings.Contains(text, "tailscale") {
		t.Fatalf("sin interfaz 100.64/10 no hay línea de Tailscale:\n%s", text)
	}
}

func TestBanner_LocalConTailscale(t *testing.T) {
	var out bytes.Buffer
	banner(&out, bannerInfo{Host: "127.0.0.1", Port: 4321, Root: "/ws", Version: "1.2.3"}, addrs("100.110.35.128/32"))
	text := out.String()
	if !strings.Contains(text, "abrí http://127.0.0.1:4321\n") || strings.Contains(text, "100.110") {
		t.Fatalf("en loopback no se promete una URL de Tailscale:\n%s", text)
	}
	if !strings.Contains(text, "tailscale: corré con --host 0.0.0.0 para abrirlo desde tu tailnet\n") {
		t.Fatalf("en loopback con Tailscale va la sugerencia del flag:\n%s", text)
	}
}

func TestBanner_SSHLocal(t *testing.T) {
	var out bytes.Buffer
	banner(&out, bannerInfo{Host: "127.0.0.1", Port: 4321, Root: "/ws", Version: "1.2.3", SSHHost: "chiche@chichex-linux"}, addrs("192.168.1.23/24"))
	text := out.String()
	if !strings.Contains(text, "abrí http://127.0.0.1:4321\n") {
		t.Fatalf("por ssh en loopback sigue la URL local:\n%s", text)
	}
	if !strings.Contains(text, "  ssh -N -L 4321:127.0.0.1:4321 chiche@chichex-linux\n") {
		t.Fatalf("por ssh en loopback va el comando del túnel:\n%s", text)
	}
	if !strings.Contains(text, "--host 0.0.0.0") {
		t.Fatalf("por ssh en loopback se sugiere la alternativa del flag:\n%s", text)
	}
}

func TestBanner_SSHLAN(t *testing.T) {
	var out bytes.Buffer
	banner(&out, bannerInfo{Host: "0.0.0.0", Port: 4321, Root: "/ws", Version: "1.2.3", SSHHost: "chiche@chichex-linux"}, addrs("192.168.1.23/24"))
	text := out.String()
	if !strings.Contains(text, "abrí http://192.168.1.23:4321\n") {
		t.Fatalf("por ssh en LAN sigue la URL LAN:\n%s", text)
	}
	if !strings.Contains(text, "ssh -N -L 4321:127.0.0.1:4321 chiche@chichex-linux") {
		t.Fatalf("por ssh en LAN el túnel va como alternativa:\n%s", text)
	}
}

func TestBanner_SinSSH(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "0.0.0.0"} {
		var out bytes.Buffer
		banner(&out, bannerInfo{Host: host, Port: 4321, Root: "/ws", Version: "1.2.3"}, addrs("192.168.1.23/24"))
		if text := out.String(); strings.Contains(text, "ssh") {
			t.Fatalf("sin sesión ssh no hay línea de ssh (host %s):\n%s", host, text)
		}
	}
}

func TestSSHHost(t *testing.T) {
	env := func(vars map[string]string) func(string) string {
		return func(k string) string { return vars[k] }
	}
	hostname := func(name string, err error) func() (string, error) {
		return func() (string, error) { return name, err }
	}
	cases := []struct {
		name     string
		env      map[string]string
		hostname func() (string, error)
		want     string
	}{
		{"sin ssh", map[string]string{"USER": "chiche"}, hostname("chichex-linux", nil), ""},
		{"ssh_connection", map[string]string{"SSH_CONNECTION": "1.2.3.4 5 6.7.8.9 22", "USER": "chiche"}, hostname("chichex-linux", nil), "chiche@chichex-linux"},
		{"ssh_client", map[string]string{"SSH_CLIENT": "1.2.3.4 5 22", "USER": "chiche"}, hostname("chichex-linux", nil), "chiche@chichex-linux"},
		{"ssh_tty", map[string]string{"SSH_TTY": "/dev/pts/1", "LOGNAME": "chiche"}, hostname("chichex-linux", nil), "chiche@chichex-linux"},
		{"sin usuario", map[string]string{"SSH_CONNECTION": "x"}, hostname("chichex-linux", nil), "chichex-linux"},
		{"sin hostname", map[string]string{"SSH_CONNECTION": "x", "USER": "chiche"}, hostname("", errors.New("nope")), "chiche@<esta-máquina>"},
	}
	for _, c := range cases {
		if got := sshHost(env(c.env), c.hostname); got != c.want {
			t.Errorf("%s: sshHost = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestMagicDNS(t *testing.T) {
	cases := map[string]string{
		`{"Self":{"DNSName":"chichex-linux.tail8ca4e8.ts.net."}}`: "chichex-linux.tail8ca4e8.ts.net",
		`{"Self":{"DNSName":""}}`:                                 "",
		`no es json`:                                              "",
		``:                                                        "",
	}
	for in, want := range cases {
		if got := magicDNS([]byte(in)); got != want {
			t.Errorf("magicDNS(%q) = %q, want %q", in, got, want)
		}
	}
}

// qrLines devuelve las líneas dibujadas con bloques Unicode del QR.
func qrLines(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		if strings.ContainsAny(line, "█▀▄") {
			out = append(out, line)
		}
	}
	return out
}
