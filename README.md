# visor

Visor local del flujo chiche (antes SDD). Antes `chichex/sdd`; ahora `chichex/visor`. Muestra cómo cada *trabajo* avanza por
Idea → Grill → Spec → Run → PR → Mergeado → Publicado en todos los repos de una
carpeta, como grafo con ramas en carriles por etapa y con un detalle tipo Gantt.
Solo lectura. Consume el log `events-v1` y los markers `Chiche-Tracking v2` y
`SDD-Tracking v1` que escriben los skills de
[`chichex/skills`](https://github.com/chichex/skills), más GitHub vía `gh`.

Lee las dos marcas de la suite, la actual (`chiche`) y la anterior (`sdd`):

- **Repos**: los que tienen `.chiche/project.md` o `.sdd/project.md` bajo `--dir`.
- **Grills y specs**: de `.chiche/` y `.sdd/` del árbol y de
  `<git-common-dir>/chiche/` y `<git-common-dir>/sdd/`. Ante el mismo nombre
  gana `.chiche/` > `.sdd/` > `<common>/chiche/` > `<common>/sdd/`.
- **Eventos**: `<git-common-dir>/chiche/events.jsonl` y
  `<git-common-dir>/sdd/events.jsonl`, unidos y ordenados por `ts`.

Estado: esqueleto. Las decisiones de producto están en
[`.sdd/grills/2026-10-23-visor-flujo-sdd.md`](.sdd/grills/2026-10-23-visor-flujo-sdd.md).

## Instalación

Un solo comando instala o actualiza el binario en `~/.local/bin` (Linux y
macOS, amd64 y arm64):

```sh
curl -fsSL https://raw.githubusercontent.com/chichex/visor/main/install.sh | sh
```

`install.sh` baja el último release, verifica el binario contra `checksums.txt`
y no hace nada si esa versión ya está instalada. Sirve desde `v0.3.0`: los
releases anteriores publicaron binarios con el nombre viejo `sdd_*`. Para fijar
una versión o la carpeta destino:

```sh
VISOR_VERSION=0.3.0 VISOR_INSTALL_DIR=/usr/local/bin sh install.sh
```

En Windows, bajá el `.exe` de [Releases](https://github.com/chichex/visor/releases).

## Cómo se construye

Binario Go con el front (Vite + React + TypeScript) embebido con `go:embed`.
Node hace falta solo para buildear el front; el binario no lo necesita.

```sh
make build          # npm ci + vite build + go build → bin/visor
./bin/visor --dir ~/workspace
./bin/visor --dir ~/workspace --host 0.0.0.0   # para abrirlo desde la LAN o Tailscale
make test
```

Con `--host 0.0.0.0` el banner imprime la URL de la LAN con un QR para el
celular. Si la máquina está en una tailnet, suma la URL de Tailscale: el nombre
MagicDNS cuando `tailscale status` responde, si no la IP `100.x`.

### Por ssh

Si corrés el visor en una máquina a la que entraste por ssh, un `127.0.0.1` es
el de esa máquina, no el tuyo. El visor detecta la sesión (por `SSH_CONNECTION`)
y resuelve según tenga Tailscale o no:

- **Con Tailscale**: escucha solo en la IP `100.x` de la tailnet (no en la LAN)
  y el banner imprime la URL MagicDNS. La abrís desde tu máquina y listo, sin
  túnel ni segunda terminal:

  ```
  abrí http://chichex-linux.tail8ca4e8.ts.net:4321
  ssh: escucho en Tailscale para que lo abras desde tu máquina sin túnel; --host 127.0.0.1 lo deja solo local
  aviso: escuchando solo en tu tailnet, sin autenticación
  ```

- **Sin Tailscale**: sigue en `127.0.0.1` y el banner imprime el túnel que
  tenés que abrir **desde tu máquina**, en otra terminal:

  ```sh
  ssh -N -L 4321:127.0.0.1:4321 chiche@chichex-linux
  ```

  Con eso abrís `http://127.0.0.1:4321` en tu navegador local. `-N` no abre
  shell: solo sostiene el túnel; cortalo con Ctrl-C. Si el puerto está ocupado
  en tu máquina (`bind: Address already in use`), cambiá el primer número:
  `-L 4322:127.0.0.1:4321` y abrís `http://127.0.0.1:4322`.

Un `--host` explícito siempre gana: `--host 127.0.0.1` lo deja solo local
aunque haya Tailscale, `--host 0.0.0.0` lo abre a la LAN.

## Versión y releases

`VERSION`, en la raíz, es la única fuente de la versión: una sola línea
`MAJOR.MINOR.PATCH`, sin `v` ni sufijos (por ejemplo `0.1.0`).

- **En cada PR, subí `VERSION`**: `0.1.0` → `0.1.1` para un fix, `0.2.0` para
  algo nuevo. El job `Version` del CI corre `go run ./cmd/versioncheck` y queda
  rojo si `VERSION` no es mayor que el último tag `v*`. En local:
  `go run ./cmd/versioncheck`.
- **Al mergear a `main`**, `.github/workflows/release.yml` corre los tests,
  repite el check, compila con `make release-assets` y publica el release
  `v<VERSION>` con notas generadas desde los PRs. El tag lo crea ese workflow en
  el commit de merge; nunca se crea a mano.
- **Binarios**: en [Releases](https://github.com/chichex/visor/releases), uno por
  plataforma (`visor_<v>_<os>_<arch>`, con `.exe` en Windows) más `checksums.txt`
  (verificable con `sha256sum -c checksums.txt`). `visor --version` imprime la
  versión; `make build` compila `<VERSION>-dev` para uso local.
- **Check requerido**: en GitHub, Settings → Branches → regla de `main` →
  «Require status checks to pass» → marcar `Version`. Hasta entonces, un merge
  sin bump deja `release.yml` rojo sin publicar nada.

## Próximos pasos

1. `/sdd-init` para el contrato de autonomía.
2. `/design-system` para la dirección visual.
3. `/sdd-spec --from-grill .sdd/grills/2026-10-23-visor-flujo-sdd.md`.
