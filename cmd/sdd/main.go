// Command sdd levanta el visor local del flujo SDD: un server HTTP que sirve
// el front embebido y una API JSON sobre los repos de una carpeta.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/chichex/sdd/web"
)

const version = "0.0.0-dev"

func main() {
	host := flag.String("host", "127.0.0.1", "interfaz donde escucha; 0.0.0.0 para abrirlo desde la LAN")
	port := flag.Int("port", 4321, "puerto")
	dir := flag.String("dir", ".", "carpeta cuyos subdirectorios se escanean buscando .sdd/project.md")
	flag.Parse()

	root, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"ok": true, "version": version, "root": root})
	})
	mux.Handle("/", spaHandler(web.Dist()))

	ln, err := net.Listen("tcp", net.JoinHostPort(*host, fmt.Sprint(*port)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("sdd %s · escaneando %s\n", version, root)
	fmt.Printf("abrí http://%s\n", ln.Addr())
	if *host != "127.0.0.1" && *host != "localhost" {
		fmt.Println("aviso: escuchando en la red local, sin autenticación")
	}
	log.Fatal(http.Serve(ln, mux))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("json:", err)
	}
}

// spaHandler sirve los archivos del build y cae a index.html para cualquier
// ruta del front. Si el front no está buildeado, lo dice en vez de dar 404.
func spaHandler(dist fs.FS) http.Handler {
	files := http.FileServerFS(dist)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name != "" && name != "index.html" {
			if f, err := dist.Open(name); err == nil {
				_ = f.Close()
				files.ServeHTTP(w, r)
				return
			}
		}
		index, err := fs.ReadFile(dist, "index.html")
		if err != nil {
			http.Error(w, "front sin buildear: corré `make web` y volvé a compilar", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(index)
	})
}
