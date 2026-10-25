# Contrato de autonomia — sdd
<!-- Generado por /sdd-init el 2026-10-25. Refrescar con /sdd-init --update. -->
<!-- SDD-Tracking: version=1; type=project; generated-at=2026-10-25 -->

## Stack
- Go 1.26 (`go.mod`; local go1.26.3), módulo `github.com/chichex/sdd`, sin dependencias externas. Binario único en `cmd/sdd` sobre `net/http`; `web/embed.go` embebe `web/dist` con `//go:embed all:dist` (`web/dist/.gitkeep` permite compilar sin front buildeado).
- Front en `web/`: Vite 8 + React 19 + TypeScript 6 + Tailwind 4 (`@tailwindcss/vite`), lint con oxlint. Package manager npm (`web/package-lock.json`), Node 24 (CI y local 24.15.0). Node hace falta solo para buildear el front; el binario no lo necesita.
- Repo único, no monorepo: `cmd/`, `web/`, `.sdd/`, `.github/workflows/`. El Makefile orquesta: `make web` (npm ci + build) → `make build` (go build) → `make run`.
- Estado: esqueleto (`/api/health` y fallback de SPA). Las decisiones de producto están en `.sdd/grills/2026-10-23-visor-flujo-sdd.md`; la primera spec sale de ahí con `/sdd-spec --from-grill`.

## Diseño
| Superficie | Raices | Tokens | Componentes | Docs y mocks | Claude Design | Estado |
|---|---|---|---|---|---|---|
| web | web/src | — (en DESIGN_SYSTEM.md) | — (en DESIGN_SYSTEM.md) | web/DESIGN_SYSTEM.md | — | con sistema |

shadcn/ui está decidido en el grill (E30) pero no instalado. La dirección visual y los tokens están en `web/DESIGN_SYSTEM.md` (dirección `sobrio-indigo`); el primer run que toque UI los materializa en código.

## Comandos
| Accion | Comando | cwd | Estado | Duracion | Notas |
|---|---|---|---|---|---|
| deps front | `npm ci` | web | verificado 2026-10-25 | ~1 s con cache | Prerequisito del build del front; incluido en `make web`. |
| test | `make test` | . | verificado 2026-10-25 | 0,1 s | `go vet ./... && go test ./...`; 2 tests en `cmd/sdd` pasan; `web` sin tests. `go test -count=1 ./...` evita el cache. |
| coverage (Go) | `go test -coverprofile=/tmp/cover.out ./... && go tool cover -func=/tmp/cover.out \| tail -1` | . | verificado 2026-10-25 | 0,1 s | baseline 35,0 % total (cmd/sdd 38,9 %, web 0,0 %). El front no tiene tooling de coverage. |
| format (Go) | `gofmt -l .` | . | verificado 2026-10-25 | <1 s | salida vacía = ok |
| typecheck (TS) | `npx tsc -b` | web | verificado 2026-10-25 | 0,5 s | `noEmit`; también corre dentro de `npm run build` |
| lint (TS) | `npm run lint` | web | verificado 2026-10-25 | 0,1 s | oxlint con `web/.oxlintrc.json`; 0 errores |
| build front | `make web` | . | verificado 2026-10-25 | 1,3 s | `npm ci && npm run build`; deja `web/dist/index.html` y `web/dist/assets/`. Solo `npm run build` en `web`: 0,6 s |
| build | `make build` | . | verificado 2026-10-25 | ~1,5 s | `make web` + `go build -o bin/sdd ./cmd/sdd` (0,2 s solo, medidos por separado); binario de 9,2 MB |
| run | `./bin/sdd --dir ..` | . | verificado 2026-10-25 | — | listo cuando `curl -sf http://127.0.0.1:4321/api/health` responde (~0,5 s). `make run` hace build + run con `DIR ?= ..`. Flags: `--host` (default 127.0.0.1), `--port` (4321), `--dir` (.). |
| dev front | `npm run dev` | web | verificado 2026-10-25 | — | Vite en `http://localhost:5173`; escucha solo en `[::1]`, así que `127.0.0.1:5173` no responde. Proxy `/api` → `127.0.0.1:4321`: necesita el binario corriendo. Listo cuando `curl -sf http://localhost:5173/` responde (~0,2 s). |

## Ambientes
- Un solo ambiente: local. No hay staging ni prod; el producto es una herramienta local que se distribuye como binario (releases de GitHub, todavía sin workflow).
- Env vars: ninguna. Toda la configuración entra por flags. No hay `.env` ni docker-compose.
- Servicios externos: ninguno hoy. El diseño prevé `gh` CLI como subproceso (GitHub, solo lectura) y lectura de clones git locales. Para probar a mano: `make run` escanea `/home/chiche/workspace` (12 repos con `.sdd/project.md`).
- Datos para verificación autónoma: solo fixtures dentro del repo (Decisiones humanas 2026-10-25): carpetas sintéticas en `testdata/` y respuestas de `gh` grabadas. Los tests no usan red ni el workspace real.
- Git: branch default `main` (`gh repo view`). Remote `origin` = `https://github.com/chichex/sdd.git`, protocolo https. `origin/HEAD` no está seteado localmente (`git remote set-head origin -a` lo arregla); /sdd-run ramifica desde `origin/main`.
- Capacidad de PR: remote configurado y `gh` autenticado como `chichex` con scopes `repo` y `workflow`: /sdd-run puede pushear branches y abrir PRs.
- CI (`.github/workflows/ci.yml`, en push a `main` y en PR): job Front (`npm ci`, `npm run build`, sube `web-dist`) y job Binario (`go vet`, `go test`, `go build` con ese dist). Eso es lo que tiene que pasar. Review a pedido: `gh workflow run claude-review.yml --repo chichex/sdd -f pr=<N>`; el secret `CLAUDE_CODE_OAUTH_TOKEN` todavía no está cargado (ver Gaps).

## Verificacion autonoma
Escalera, de menor a mayor confianza:
1. Estática: `gofmt -l .` vacío, `go vet ./...`, y en `web` `npx tsc -b` y `npm run lint`.
2. Unit: `go test -count=1 ./...`. Hoy cubre solo `spaHandler` con `fstest.MapFS`. El front no tiene runner de tests.
3. Build: `make build` produce `bin/sdd` con el front embebido; `web/dist/index.html` existe.
4. App levantada: `./bin/sdd --dir <fixture de testdata/>` en background, `curl` sobre `/api/*` y sobre rutas del SPA (fallback a `index.html`), y matar el proceso. Siempre contra fixtures, nunca contra `/home/chiche/workspace` ni GitHub.
5. E2E de UI: no disponible. No hay Playwright en el repo ni MCP de navegador en la sesión.

No se puede verificar sin humano:
- La UI renderizada (carriles, grafo, Gantt, modo oscuro, celular por LAN): hace falta un navegador. El humano mira la app con `make run`.
- El comportamiento contra GitHub real y contra el workspace real: queda fuera por decisión humana; se prueba a mano.
- La degradación sin red y el rate limit: solo con respuestas grabadas.

## Limites
Sin confirmación humana, NO:
- Deploy ni publicación: no crear releases, tags `v*` ni subir binarios.
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

## Gaps
- Front sin runner de tests: coverage del front no activable. Sumar Vitest (o similar) cuando la spec de UI lo pida y reofrecer coverage con `/sdd-init --update`.
- Coverage Go no activado; baseline 35,0 % (2026-10-25). Activable con `/sdd-init --update`.
- Sin e2e de UI: no hay Playwright en el repo ni MCP de navegador en la sesión. La UI la valida un humano.
- `testdata/` y las respuestas grabadas de `gh` no existen todavía: las crea la primera spec que las necesite.
- Secret `CLAUDE_CODE_OAUTH_TOKEN` no cargado en `chichex/sdd`: `claude-review.yml` falla hasta correr `gh secret set CLAUDE_CODE_OAUTH_TOKEN --repo chichex/sdd`.
- `origin/HEAD` sin setear localmente: `git remote set-head origin -a`.
- Excepción de `no-panic` en `web/embed.go`: se elimina cuando `Dist()` devuelva error en vez de panic.
- Distribución del binario (un archivo por plataforma, grill D27): sin workflow de release todavía.
