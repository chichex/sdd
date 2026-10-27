package main

import (
	"bytes"
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
