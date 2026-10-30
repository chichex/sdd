# Contrato de autonomia — sdd
<!-- Generado por /sdd-init el 2026-10-25, refrescado el 2026-10-29. Refrescar con /sdd-init --update. -->
<!-- SDD-Tracking: version=1; type=project; generated-at=2026-10-29 -->

## Stack
- Go 1.26 (`go.mod`; local go1.26.3), módulo `github.com/chichex/sdd`. Una sola dependencia externa: `github.com/skip2/go-qrcode`, para el QR de la URL en el banner de arranque (`cmd/sdd/banner.go`). Binario único en `cmd/sdd` sobre `net/http`; `web/embed.go` embebe `web/dist` con `//go:embed all:dist` (`web/dist/.gitkeep` permite compilar sin front buildeado).
- Paquetes: `internal/events` (parser `events-v1`), `internal/tracking` (markers `SDD-Tracking v1`), `internal/scan` (lectura de repos, worktrees y artefactos), `internal/github` (snapshots vía `gh`, con el doble `internal/github/fake`), `internal/model` (trabajos, etapas, ramas, categorías, Gantt), `internal/workspace` (servicio: escaneo periódico, caché, SSE) e `internal/api` (handlers HTTP).
- Front en `web/`: Vite 8 + React 19 + TypeScript 6 + Tailwind 4 (`@tailwindcss/vite`), shadcn/ui (`web/components.json`, `web/src/components/ui`), lucide-react y sonner; tests con Vitest 5 + jsdom + Testing Library; lint con oxlint. Package manager npm (`web/package-lock.json`), Node 24 (CI y local 24.15.0). Node hace falta solo para buildear y testear el front; el binario no lo necesita.
- Repo único, no monorepo: `cmd/`, `internal/`, `web/`, `testdata/`, `.sdd/`, `.github/workflows/`. El Makefile orquesta: `make web` (npm ci + build) → `make build` (go build) → `make run`; `make test` corre solo los tests Go.
- Estado: visor funcional. Cuatro specs implementadas en `.sdd/specs/` (`visor-flujo-sdd`, `visor-anclas-traza-filtros`, `visor-lista-etapas`, `visor-categorias-grafo`) y el grill de producto en `.sdd/grills/2026-10-23-visor-flujo-sdd.md`. API: `GET /api/health`, `GET /api/workspace?repo=&mostrar=`, `GET /api/trabajos/{id}`, `POST /api/refresh` y `GET /api/stream` (SSE). Flags: `--host`, `--port`, `--dir`, `--watch-interval`, `--github-interval`, `--cache-dir`, `--no-github`.

## Diseño
| Superficie | Raices | Tokens | Componentes | Docs y mocks | Claude Design | Estado |
|---|---|---|---|---|---|---|
| web | web/src | web/src/index.css (`@theme`, 119 variables) | web/src/components/ui (9 de shadcn/ui) | web/DESIGN_SYSTEM.md | — | con sistema |

Dirección `sobrio-indigo`. Los tokens del doc están materializados en `web/src/index.css`; «Ajustes de este proyecto» del doc lleva las piezas que sumaron las tres specs de UI. Baselines de gates de diseño medidos el 2026-10-29, no activados: hex fuera de `web/src/index.css` 0, `style={{` fuera de `components/ui` 3, nombres de componente repetidos entre archivos 4.

## Comandos
| Accion | Comando | cwd | Estado | Duracion | Notas |
|---|---|---|---|---|---|
| deps front | `npm ci` | web | verificado 2026-10-29 | ~1 s con cache | 102 paquetes; prerequisito del build y los tests del front; incluido en `make web`. |
| test (Go) | `make test` | . | verificado 2026-10-29 | 0,1 s con cache · 3,0 s con `-count=1` | `go vet ./... && go test ./...`; 7 paquetes ok, 63 tests top-level. `internal/workspace` y `web` sin tests propios. `go test -count=1 ./...` evita el cache. |
| test (front) | `npm test` | web | verificado 2026-10-29 | 2,1 s | `vitest run`: 12 archivos, 176 tests pasan (jsdom, setup en `src/test/setup.ts`, API mockeada en `src/test/api-mock.ts`). En CI desde el chore de este contrato. |
| coverage (Go) | `go test -coverprofile=/tmp/cover.out ./... && go tool cover -func=/tmp/cover.out \| tail -1` | . | verificado 2026-10-29 | 3,0 s | baseline 74,8 % total (api 93,8, tracking 91,4, model 89,7, events 86,0, github 81,0, scan 77,6, cmd/sdd 41,7). El front no tiene tooling de coverage. |
| format (Go) | `gofmt -l .` | . | verificado 2026-10-29 | <1 s | salida vacía = ok |
| typecheck (TS) | `npx tsc -b` | web | verificado 2026-10-29 | 1,2 s | también corre dentro de `npm run build` |
| lint (TS) | `npm run lint` | web | verificado 2026-10-29 | 0,1 s | oxlint con `web/.oxlintrc.json`; 0 errores |
| build front | `make web` | . | verificado 2026-10-29 | 3,1 s | `npm ci && npm run build` (`rm -rf dist/assets && tsc -b && vite build`; solo `npm run build` en `web`: 1,3 s); deja `web/dist/index.html` y `web/dist/assets/` |
| build | `make build` | . | verificado 2026-10-29 | 3,0 s | `make web` + `go build -o bin/sdd ./cmd/sdd` (0,2 s solo); binario de 11 MB |
| run | `./bin/sdd --dir <carpeta>` | . | verificado 2026-10-29 | — | listo cuando `curl -sf http://127.0.0.1:4321/api/health` responde (unos ms tras arrancar); el primer escaneo terminó cuando `/api/workspace` trae `scanning:false`. `make run` hace build + run con `DIR ?= ..`. Con `--no-github` no invoca `gh`; `--host 0.0.0.0` lo abre a la LAN e imprime un QR de la URL. |
| dev front | `npm run dev` | web | verificado 2026-10-29 | — | Vite en `http://localhost:5173`, listo en ~1 s; escucha solo en `[::1]`, así que `127.0.0.1:5173` no responde. Proxy `/api` → `127.0.0.1:4321`: necesita el binario corriendo. |

## Ambientes
- Un solo ambiente: local. No hay staging ni prod; el producto es una herramienta local que se distribuye como binario (releases de GitHub que publica `release.yml` en cada merge a `main`).
- Env vars: ninguna. Toda la configuración entra por flags. No hay `.env` ni docker-compose. Los snapshots de GitHub se cachean en la caché del usuario (`--cache-dir` la cambia).
- Servicios externos: `gh` CLI como subproceso (GitHub, solo lectura; `--no-github` lo apaga) y lectura de clones git locales. Para probar a mano: `./bin/sdd --host 0.0.0.0 --dir /home/chiche/workspace` escanea los repos con `.sdd/project.md` bajo esa carpeta.
- Datos para verificación autónoma: solo fixtures dentro del repo (Decisiones humanas 2026-10-25): `testdata/vectors` (conformance de `events-v1` y `SDD-Tracking v1`), `testdata/events`, `testdata/tracking`, `testdata/gh` (respuestas de `gh` grabadas, incluidos rollups de checks), `testdata/mkworkspace` (arma un workspace con repos git reales en `t.TempDir()`; `fixtureWorkspace` en `internal/api/helpers_test.go`) y `testdata/workspace` (markdown sueltos, sin repos git). Los tests no usan red ni el workspace real.
- Git: branch default `main` (`gh repo view`); `origin/HEAD` apunta a `main`. Remote `origin` = `https://github.com/chichex/sdd.git`, protocolo https. /sdd-run ramifica desde `origin/main`.
- Capacidad de PR: remote configurado y `gh` autenticado como `chichex` con scopes `repo` y `workflow`: /sdd-run puede pushear branches y abrir PRs. Extensión `gh-stack` v0.2.1 instalada: los stacks de capas se vinculan y mergean como stacks nativos de GitHub (`/sdd-land`).
- CI (`.github/workflows/ci.yml`, en push a `main` y en PR): job Front (`npm ci`, `npm test`, `npm run build`, sube `web-dist`), job Binario (`go vet`, `go test`, `go build` con ese dist) y, solo en PR, job Version (`go run ./cmd/versioncheck`: `VERSION` mayor que el último tag `v*`). Release (`.github/workflows/release.yml`, en push a `main`): tests, `versioncheck`, `make release-assets` y `gh release create v<VERSION>`. Eso es lo que tiene que pasar. Review a pedido: `gh workflow run claude-review.yml --repo chichex/sdd -f pr=<N>`; el secret `CLAUDE_CODE_OAUTH_TOKEN` todavía no está cargado (ver Gaps).
- Log SDD propio: `.git/sdd/events.jsonl` (fuera del árbol) es lo que el visor lee de este mismo repo.

## Verificacion autonoma
Escalera, de menor a mayor confianza:
1. Estática: `gofmt -l .` vacío, `go vet ./...`, y en `web` `npx tsc -b` y `npm run lint`.
2. Unit e integración: `go test -count=1 ./...` (63 tests: parsers contra los vectores de conformance, modelo, scan sobre repos git en `TempDir`, API con `httptest` y `gh` grabado) y en `web` `npm test` (176 tests de componentes y hooks con jsdom y la API mockeada).
3. Build: `make build` produce `bin/sdd` con el front embebido; `web/dist/index.html` existe.
4. App levantada: `./bin/sdd --port 4322 --dir testdata/workspace --no-github` en background; `curl` a `/api/health`, a `/api/workspace` (`scanning:false`; `?mostrar=otro` → 400) y a una ruta del SPA (`/t/<id>` → 200 `text/html`); matar el proceso. `testdata/workspace` no tiene repos git: sirve para probar la API y el fallback del SPA, no el modelo; para datos reales está el workspace que arma `fixtureWorkspace` en los tests de `internal/api`. Siempre contra fixtures, nunca contra `/home/chiche/workspace` ni GitHub.
5. E2E de UI: no disponible en el repo (sin Playwright ni MCP de navegador). Ver Gaps para el probe manual que existe en esta máquina.

No se puede verificar sin humano:
- La UI renderizada en celular real (LAN o Tailscale), modo oscuro y gestos: hace falta un navegador y un teléfono. El humano mira la app con `make run` o `./bin/sdd --host 0.0.0.0`.
- El comportamiento contra GitHub real y contra el workspace real: queda fuera por decisión humana; se prueba a mano.
- La degradación sin red y el rate limit: solo con respuestas grabadas.

## Limites
Sin confirmación humana, NO:
- Releases y tags `v*`: los crea solo `release.yml` al mergear a `main`; nunca a mano ni desde un run.
- `git push` a `main`: solo branches `sdd/*` o `chore/*` con PR.
- Migraciones sobre datos compartidos: no aplica (sin base de datos).
- Tocar servicios pagos: no aplica. Tampoco gastar el rate limit de GitHub en verificaciones: los tests no usan red.
- Escribir en los repos escaneados ni en GitHub: el visor es solo lectura (grill D25 y C22). Ningún test ni verificación modifica clones ajenos ni `/home/chiche/workspace`.
- Disparar `claude-review.yml` ni otros workflows.

## Politicas de generacion
| Politica | Valor | Gate |
|---|---|---|
| commits | `tipo(scope)?: resumen`; tipos feat, fix, chore, docs, test, refactor, ci, build, perf, style; `!` opcional antes de `:` | `git log --format=%s origin/main..HEAD \| grep -vE '^(feat\|fix\|chore\|docs\|test\|refactor\|ci\|build\|perf\|style)(\([a-z0-9/_.-]+\))?!?: .+$'` vacío (verificado 2026-10-25 sobre los 2 commits de main) |
| max-lineas-archivo | 300 por archivo `.go`, `.ts` o `.tsx` tocado (hoy el más largo tiene 79) | `git diff --name-only --diff-filter=AM origin/main...HEAD -- '*.go' '*.ts' '*.tsx' \| xargs -r wc -l \| awk '$2!="total" && $1>300'` vacío (verificado 2026-10-25 sobre todos los trackeados) |
| no-any-ts | sin `any` explícito en TypeScript | `npx oxlint -D typescript/no-explicit-any` en `web`, exit 0 (verificado 2026-10-25: 0 errores en el repo; la regla dispara sobre un `const x: any`) |
| gofmt | código Go formateado | `gofmt -l .` vacío (verificado 2026-10-25) |
| estilo-go | guía de estilo de Uber para Go (github.com/uber-go/guide) | guia — sin gate medible: /sdd-run la sigue al generar, la juzga el reviewer |
| no-panic | ningún `panic(` en Go fuera de `cmd/` y de tests. Excepción documentada: `web/embed.go` (`fs.Sub` del embed; solo falla con un build roto) | `git grep -n 'panic(' -- '*.go' ':!cmd/' ':!*_test.go' ':!web/embed.go'` vacío (verificado 2026-10-25; sin la excepción hoy hay 1 hit) |
| coding-policies | .sdd/coding-policies.md (go, typescript, react) | guia — sin gate: /sdd-run la sigue al generar, la juzga el reviewer |

## Decisiones humanas
- 2026-10-25 — Datos para verificar sin humano: solo fixtures dentro del repo (`testdata/` sintético y respuestas de `gh` grabadas), sin red. El workspace real y GitHub real quedan para pruebas manuales.
- 2026-10-25 — Políticas activas: commits convencionales `tipo(scope)?: resumen`; máximo 300 líneas por archivo Go/TS; sin `any` en TS; gofmt limpio; guía de estilo de Uber como guía; sin `panic()` fuera de `cmd/`, con excepción de `web/embed.go`. No activadas: tamaño de PR, dependencias nuevas, coverage.
- 2026-10-25 — Generadores encadenados al terminar el contrato: `/coding-policies` y `/design-system` (superficie web sin sistema).
- 2026-10-29 — Refresco tras aterrizar el visor: políticas mantenidas sin cambios (no activadas: coverage Go con baseline 74,8 %, gates de diseño, tamaño de PR, dependencias). `npm test` se suma al job Front del CI y los cuatro bullets base desactualizados de `web/DESIGN_SYSTEM.md` se corrigen a mano, los dos en el mismo chore PR.

## Gaps
- Front sin tooling de coverage (`@vitest/coverage-v8` no instalado): coverage del front no activable. Sumarlo y reofrecer con `/sdd-init --update`.
- Coverage Go no activado; baseline 74,8 % (2026-10-29). Activable con `/sdd-init --update`.
- `internal/workspace` sin tests propios: lo cubren de forma indirecta los de `internal/api`.
- Sin e2e de UI en el repo. En esta máquina hay `playwright-core` 1.63 en la caché de npx (`~/.npm/_npx/40a319e23a391f24/node_modules/playwright-core`) con Chromium 1243 ya descargado: sirve para probes manuales con un script de Node (capturas, conteos del DOM) contra el binario levantado; no es un escalón del contrato ni corre en CI.
- Secret `CLAUDE_CODE_OAUTH_TOKEN` no cargado en `chichex/sdd`: `claude-review.yml` falla hasta correr `gh secret set CLAUDE_CODE_OAUTH_TOKEN --repo chichex/sdd`.
- Excepción de `no-panic` en `web/embed.go`: se elimina cuando `Dist()` devuelva error en vez de panic.
- check `Version` todavía no marcado como requerido en branch protection: lo hace el humano. Hasta entonces, un merge sin bump deja `release.yml` rojo sin publicar nada.
