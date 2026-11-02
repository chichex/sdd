package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
)

type bannerInfo struct {
	Host    string
	Port    int
	Root    string
	Version string
	// TailscaleDNS es el nombre MagicDNS de esta máquina, vacío si no se pudo
	// averiguar. La IP 100.x se detecta aparte, mirando las interfaces.
	TailscaleDNS string
	// SSHHost es "usuario@máquina" cuando el visor corre dentro de una sesión
	// ssh, vacío si no. Con él el banner imprime el túnel que hay que abrir
	// desde la máquina del cliente para ver el visor.
	SSHHost string
}

// sshHost detecta una sesión ssh por las variables que sshd deja en el entorno
// y arma "usuario@máquina" para el comando del túnel. Devuelve vacío fuera
// de ssh.
func sshHost(getenv func(string) string, hostname func() (string, error)) string {
	if getenv("SSH_CONNECTION") == "" && getenv("SSH_CLIENT") == "" && getenv("SSH_TTY") == "" {
		return ""
	}
	name, err := hostname()
	if err != nil || name == "" {
		name = "<esta-máquina>"
	}
	user := getenv("USER")
	if user == "" {
		user = getenv("LOGNAME")
	}
	if user == "" {
		return name
	}
	return user + "@" + name
}

// sshTunnel es el comando que, corrido en la máquina del cliente, trae el
// puerto del visor a su 127.0.0.1. -N no abre shell: solo mantiene el túnel.
func sshTunnel(info bannerInfo) string {
	port := strconv.Itoa(info.Port)
	return "ssh -N -L " + port + ":127.0.0.1:" + port + " " + info.SSHHost
}

// tailscaleNet es el rango CGNAT que Tailscale reserva para la tailnet.
var tailscaleNet = net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

// isLoopback: el default 127.0.0.1 y localhost solo se abren desde esta máquina.
func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// ipv4s devuelve las IPv4 no loopback de las interfaces, en orden.
func ipv4s(addrs func() ([]net.Addr, error)) []net.IP {
	list, err := addrs()
	if err != nil {
		return nil
	}
	var out []net.IP
	for _, a := range list {
		if n, ok := a.(*net.IPNet); ok && !n.IP.IsLoopback() && n.IP.To4() != nil {
			out = append(out, n.IP.To4())
		}
	}
	return out
}

// lanIPv4 devuelve la primera IPv4 de la LAN, dejando afuera la de Tailscale.
func lanIPv4(addrs func() ([]net.Addr, error)) string {
	for _, ip := range ipv4s(addrs) {
		if !tailscaleNet.Contains(ip) {
			return ip.String()
		}
	}
	return ""
}

// tailscaleIPv4 devuelve la IPv4 de la tailnet, vacío si no hay Tailscale.
func tailscaleIPv4(addrs func() ([]net.Addr, error)) string {
	for _, ip := range ipv4s(addrs) {
		if tailscaleNet.Contains(ip) {
			return ip.String()
		}
	}
	return ""
}

// magicDNS saca el nombre MagicDNS de la salida de `tailscale status --json`.
func magicDNS(out []byte) string {
	var status struct {
		Self struct{ DNSName string }
	}
	if err := json.Unmarshal(out, &status); err != nil {
		return ""
	}
	return strings.TrimSuffix(status.Self.DNSName, ".")
}

// tailscaleDNS pregunta al CLI el nombre MagicDNS. Sin `tailscale` en el PATH
// o con el daemon caído devuelve vacío: el banner se arregla con la IP 100.x.
func tailscaleDNS() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "tailscale", "status", "--json").Output()
	if err != nil {
		return ""
	}
	return magicDNS(out)
}

// banner imprime cómo abrir el visor. Escuchando en la LAN suma un QR de la
// URL (para el celular), la URL por Tailscale si la máquina está en una
// tailnet, y el aviso de que no hay autenticación.
func banner(w io.Writer, info bannerInfo, addrs func() ([]net.Addr, error)) {
	fmt.Fprintf(w, "visor %s · escaneando %s\n", info.Version, info.Root)
	port := strconv.Itoa(info.Port)
	host := info.Host
	tsIP := tailscaleIPv4(addrs)
	if isLoopback(host) {
		localURL := "http://" + net.JoinHostPort(host, port)
		fmt.Fprintf(w, "abrí %s\n", localURL)
		if info.SSHHost != "" {
			// Por ssh esa URL es la de la máquina remota: desde el cliente
			// solo se llega con un túnel (o abriendo el server a la red).
			fmt.Fprintln(w, "ssh: estás por ssh y el visor solo escucha en esta máquina. Desde la tuya corré")
			fmt.Fprintf(w, "  %s\n", sshTunnel(info))
			fmt.Fprintf(w, "y abrí %s ahí; o levantá el visor con --host 0.0.0.0 y entrá por la LAN o Tailscale\n", localURL)
		}
		if tsIP != "" {
			fmt.Fprintln(w, "tailscale: corré con --host 0.0.0.0 para abrirlo desde tu tailnet")
		}
		return
	}
	unspecified := false
	if ip := net.ParseIP(host); host == "" || (ip != nil && ip.IsUnspecified()) {
		unspecified = true
		host = lanIPv4(addrs)
		if host == "" {
			host = tsIP
		}
	}
	if host == "" {
		host = "127.0.0.1"
	}
	url := "http://" + net.JoinHostPort(host, port)
	fmt.Fprintf(w, "abrí %s\n", url)
	// Nivel L: una URL LAN entra en la versión 2 (25 módulos + 4 de borde
	// por lado = 33 de ancho), que la terminal del celular escanea sin zoom.
	if qr, err := qrcode.New(url, qrcode.Low); err == nil {
		fmt.Fprint(w, qr.ToSmallString(false))
	}
	// La URL de Tailscale solo sirve si el server escucha en esa interfaz:
	// con 0.0.0.0 o con la propia IP 100.x. Con MagicDNS va el nombre.
	if tsIP != "" && (unspecified || host == tsIP) {
		tsHost := tsIP
		if info.TailscaleDNS != "" {
			tsHost = info.TailscaleDNS
		}
		if tsHost != host {
			fmt.Fprintf(w, "tailscale: http://%s\n", net.JoinHostPort(tsHost, port))
		}
	}
	if info.SSHHost != "" {
		fmt.Fprintf(w, "ssh: si esa red no te llega, desde tu máquina corré `%s` y abrí http://127.0.0.1:%s\n", sshTunnel(info), port)
	}
	fmt.Fprintln(w, "aviso: escuchando en la red local, sin autenticación")
}
