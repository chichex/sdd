# sdd

Visor local del flujo SDD. Muestra cómo cada *trabajo* avanza por
Idea → Grill → Spec → Run → PR → Mergeado → Publicado en todos los repos de una
carpeta, como grafo con ramas en carriles por etapa y con un detalle tipo Gantt.
Solo lectura. Consume el log `events-v1` y los markers `SDD-Tracking v1` que
escriben los skills de [`chichex/skills`](https://github.com/chichex/skills),
más GitHub vía `gh`.

Estado: esqueleto. Las decisiones de producto están en
[`.sdd/grills/2026-10-23-visor-flujo-sdd.md`](.sdd/grills/2026-10-23-visor-flujo-sdd.md).

## Cómo se construye

Binario Go con el front (Vite + React + TypeScript) embebido con `go:embed`.
Node hace falta solo para buildear el front; el binario no lo necesita.

```sh
make build          # npm ci + vite build + go build → bin/sdd
./bin/sdd --dir ~/workspace
./bin/sdd --dir ~/workspace --host 0.0.0.0   # para abrirlo desde la LAN
make test
```

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
- **Binarios**: en [Releases](https://github.com/chichex/sdd/releases), uno por
  plataforma (`sdd_<v>_<os>_<arch>`, con `.exe` en Windows) más `checksums.txt`
  (verificable con `sha256sum -c checksums.txt`). `sdd --version` imprime la
  versión; `make build` compila `<VERSION>-dev` para uso local.
- **Check requerido**: en GitHub, Settings → Branches → regla de `main` →
  «Require status checks to pass» → marcar `Version`. Hasta entonces, un merge
  sin bump deja `release.yml` rojo sin publicar nada.

## Próximos pasos

1. `/sdd-init` para el contrato de autonomía.
2. `/design-system` para la dirección visual.
3. `/sdd-spec --from-grill .sdd/grills/2026-10-23-visor-flujo-sdd.md`.
