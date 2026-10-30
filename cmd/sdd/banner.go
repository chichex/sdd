package main

import (
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/skip2/go-qrcode"
)

type bannerInfo struct {
	Host    string
	Port    int
	Root    string
	Version string
}

// isLoopback: el default 127.0.0.1 y localhost solo se abren desde esta máquina.
func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// lanIPv4 devuelve la primera IPv4 no loopback de las interfaces.
func lanIPv4(addrs func() ([]net.Addr, error)) string {
	list, err := addrs()
	if err != nil {
		return ""
	}
	for _, a := range list {
		if n, ok := a.(*net.IPNet); ok && !n.IP.IsLoopback() && n.IP.To4() != nil {
			return n.IP.String()
		}
	}
	return ""
}

// banner imprime cómo abrir el visor. Escuchando en la LAN suma un QR de la
// URL (para el celular) y el aviso de que no hay autenticación.
func banner(w io.Writer, info bannerInfo, addrs func() ([]net.Addr, error)) {
	fmt.Fprintf(w, "sdd %s · escaneando %s\n", info.Version, info.Root)
	host := info.Host
	if isLoopback(host) {
		fmt.Fprintf(w, "abrí http://%s\n", net.JoinHostPort(host, strconv.Itoa(info.Port)))
		return
	}
	if ip := net.ParseIP(host); host == "" || (ip != nil && ip.IsUnspecified()) {
		host = lanIPv4(addrs)
	}
	if host == "" {
		host = "127.0.0.1"
	}
	url := "http://" + net.JoinHostPort(host, strconv.Itoa(info.Port))
	fmt.Fprintf(w, "abrí %s\n", url)
	// Nivel L: una URL LAN entra en la versión 2 (25 módulos + 4 de borde
	// por lado = 33 de ancho), que la terminal del celular escanea sin zoom.
	if qr, err := qrcode.New(url, qrcode.Low); err == nil {
		fmt.Fprint(w, qr.ToSmallString(false))
	}
	fmt.Fprintln(w, "aviso: escuchando en la red local, sin autenticación")
}
