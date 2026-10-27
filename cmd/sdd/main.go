// Command sdd levanta el visor local del flujo SDD: un server HTTP que sirve
// el front embebido y una API JSON sobre los repos de una carpeta.
package main

import (
	"context"
	"errors"
	"flag"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/chichex/sdd/internal/api"
	"github.com/chichex/sdd/internal/workspace"
	"github.com/chichex/sdd/web"
)

const version = "0.0.0-dev"

func main() {
	host := flag.String("host", "127.0.0.1", "interfaz donde escucha; 0.0.0.0 para abrirlo desde la LAN")
	port := flag.Int("port", 4321, "puerto")
	dir := flag.String("dir", ".", "carpeta cuyos subdirectorios se escanean buscando .sdd/project.md")
	watch := flag.Duration("watch-interval", workspace.DefaultWatchInterval, "cada cuánto se revisan los logs y artefactos locales")
	flag.Parse()

	root, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	svc := workspace.New(workspace.Config{Root: root, WatchInterval: *watch})
	mux := http.NewServeMux()
	api.New(svc, version).Register(mux)
	mux.Handle("/", spaHandler(web.Dist()))

	ln, err := net.Listen("tcp", net.JoinHostPort(*host, strconv.Itoa(*port)))
	if err != nil {
		log.Fatal(err)
	}
	listenPort := *port
	if tcp, ok := ln.Addr().(*net.TCPAddr); ok {
		listenPort = tcp.Port
	}
	banner(os.Stdout, bannerInfo{Host: *host, Port: listenPort, Root: root, Version: version}, net.InterfaceAddrs)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = svc.Run(ctx) // solo termina al cancelar ctx
	}()
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Println("shutdown:", err)
		}
	}()
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
	wg.Wait()
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
