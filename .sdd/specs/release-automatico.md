# Spec — Release automático al mergear a main
<!-- Generada por /sdd-spec el 2026-10-29. Fuente: pedido libre. Estado: aprobada -->
<!-- SDD-Tracking: version=1; type=spec; state=approved; issue=none; grill=none; superseded-by=none -->

## Contexto
Hoy la versión del binario es la constante `version = "0.0.0-dev"` en `cmd/sdd/main.go:29`, que `/api/health` devuelve (`internal/api/api.go:64`) y el banner imprime (`cmd/sdd/banner.go:45`); el Makefile compila con `go build -o bin/sdd ./cmd/sdd` sin `-ldflags`. No hay tags, releases ni archivo de versión, y el CI (`.github/workflows/ci.yml`) solo buildea y testea. El contrato (`.sdd/project.md`, Limites) prohíbe crear releases y tags `v*` sin humano, y en Gaps anota la distribución por plataforma como pendiente. El pedido: un archivo `VERSION` como única fuente de la semver, un release de GitHub con binarios por cada merge a `main`, y un check en cada PR que obligue a subir la versión. Tres decisiones ya confirmadas por el usuario: `VERSION` plano; el release publica tag, binarios y checksums; el check exige una versión mayor que el último tag `v*`.

## Comportamiento esperado
- **CA-1 (ALTA)** — Existe `VERSION` en la raíz del repo con exactamente una línea `MAJOR.MINOR.PATCH` (sin `v`, sin sufijos), valor inicial `0.1.0`, con salto de línea final. Un test lee el archivo real y lo parsea con `internal/semver`.
- **CA-2 (ALTA)** — Paquete `internal/semver`: `Parse` acepta `0.1.0`, `1.2.3`, `0.10.0` y rechaza `v1.2.3`, `1.2`, `1.2.3-rc1`, `1.2.3+build`, `01.2.3`, `1.2.3 ` y vacío, con un error que nombra el valor; `Compare` ordena numéricamente por major, minor y patch (`0.10.0 > 0.9.1`, `1.0.0 > 0.99.99`); `Latest(tags)` devuelve la mayor versión entre tags con forma `v<semver>` e ignora los que no matchean (`latest`, `v1.2`, `release-1`); con ninguno válido devuelve «sin versión».
- **CA-3 (ALTA)** — `go run ./cmd/versioncheck`, corrido en la raíz: lee `VERSION`, lista los tags `v*` del repo (`git tag --list 'v*'`) y sale 0 imprimiendo `VERSION <v> > último tag <tag>` (o `VERSION <v>, sin tags previos`) cuando la versión es mayor; sale 1 con un mensaje que dice qué falló cuando `VERSION` falta, no es semver válida, o es igual o menor que el último tag. El núcleo es puro (`check(version, tags)`) y se testea table-driven; un test de integración crea un repo git en `t.TempDir()` con tags reales y corre el listado de tags contra él. Sin `panic`.
- **CA-4 (MEDIA)** — El binario toma la versión en build: `version` pasa de `const` a `var` en `cmd/sdd`, reemplazable con `-ldflags "-X main.version=<v>"`; `make build` inyecta `<VERSION>-dev` (por ejemplo `0.1.0-dev`) y el release inyecta `<VERSION>` exacta. Un flag nuevo `--version` imprime `sdd <version>` en stdout y termina con exit 0 sin escanear nada ni abrir puertos. `/api/health` y el banner muestran la misma versión inyectada.
- **CA-5 (MEDIA)** — `make release-assets` compila, con `CGO_ENABLED=0` y `-ldflags "-s -w -X main.version=<VERSION>"`, seis binarios en `dist/release/`: `sdd_<v>_linux_amd64`, `sdd_<v>_linux_arm64`, `sdd_<v>_darwin_amd64`, `sdd_<v>_darwin_arm64`, `sdd_<v>_windows_amd64.exe`, `sdd_<v>_windows_arm64.exe`, más `dist/release/checksums.txt` en formato `sha256sum` (verificable con `sha256sum -c`). Requiere el front ya buildeado (`web/dist`), falla con exit distinto de 0 si `VERSION` es inválida, y `dist/` queda en `.gitignore`.
- **CA-6 (MEDIA)** — `ci.yml` suma el job `Version` en `pull_request`: checkout con historia y tags completos y `go run ./cmd/versioncheck`. Falla si `VERSION` no es mayor que el último tag. En el PR de este mismo run el job aparece en `gh pr checks` y está verde (`0.1.0`, sin tags).
- **CA-7 (NULA; estático BAJA)** — Workflow nuevo `.github/workflows/release.yml`: en `push` a `main`, `concurrency: release` (en cola, sin cancelar), `permissions: contents: write`; pasos: checkout con tags, Node 24 + `npm ci` + `npm test` + `npm run build` en `web`, Go desde `go.mod` + `go vet ./...` + `go test ./...`, `go run ./cmd/versioncheck` (si el tag ya existe o la versión es inválida el workflow falla ahí, antes de compilar), `make release-assets` y `gh release create v<VERSION> dist/release/* --title v<VERSION> --generate-notes` con `GH_TOKEN: ${{ github.token }}`, que crea el tag en el commit de merge. Resultado observable: un merge a `main` produce el release `v<VERSION>` con 7 assets y el tag; un segundo push sin bump deja el run rojo en el paso de versioncheck sin publicar nada.
- **CA-8 (BAJA)** — Documentación y contrato: README gana la sección «Versión y releases» (qué es `VERSION`, cómo bumpearla en cada PR, qué pasa al mergear, dónde quedan los binarios, cómo marcar el check como requerido); `.sdd/project.md` cambia en Limites la línea de releases por «Releases y tags `v*`: los crea solo `release.yml` al mergear a `main`; nunca a mano ni desde un run», suma en Ambientes el job `Version` y `release.yml` a lo que tiene que pasar, y en Gaps reemplaza la línea de distribución por «check `Version` todavía no marcado como requerido en branch protection: lo hace el humano».

## Plan de entrega
| Capa | Etapa | CAs | Justificacion |
|---|---|---|---|
| 1 | release | CA-1 a CA-8 | Un solo PR: `VERSION`, el paquete semver, la herramienta de check, los ldflags, el target de assets, los dos workflows y la doc se sostienen juntos; sin tamaño máximo de PR activo y unas 12 rutas tocadas. |

## Fuera de alcance
- Pre-releases, canales beta, versiones con sufijo (`-rc1`) o metadata (`+build`).
- CHANGELOG mantenido a mano: las notas las genera GitHub desde los PRs.
- Paquetes de instalación (Homebrew, apt, scoop, `go install`), firmas o notarización.
- Bump automático de `VERSION` (por conventional commits o similar): el bump es manual en cada PR.
- Cambiar la protección de `main` para exigir el check: lo hace el humano en GitHub.
- Release desde tags empujados a mano o por `workflow_dispatch`.

## Inferencias
| # | Inferencia | Eleccion propuesta | Alternativa razonable | Confianza | Resolucion |
|---|---|---|---|---|---|
| 1 | ¿Forma del archivo de versión? | `VERSION` plano con solo la semver | `release.yml` con campos | — | confirmada |
| 2 | ¿Qué publica cada release? | tag `v<semver>`, 6 binarios y `checksums.txt` | solo tag y notas | — | confirmada |
| 3 | ¿Regla del check del PR? | `VERSION` mayor que el último tag `v*` | distinta de la base del PR | — | confirmada |
| 4 | ¿Versión inicial? | `0.1.0` | `1.0.0` | media | [ASSUMED] |
| 5 | ¿Qué semver se acepta? | solo `MAJOR.MINOR.PATCH` numéricos | semver 2.0 completa con precedencia de pre-release | alta | [ASSUMED] |
| 6 | ¿Dónde vive la lógica de versión? | Go: `internal/semver` puro y `cmd/versioncheck` solo wiring (layout `cmd/` e `internal/` de las coding policies, tests table-driven, sin `panic`) | script `sh` con `sort -V` en `scripts/` | media | [ASSUMED] |
| 7 | ¿Qué versión muestra un build local? | `<VERSION>-dev` desde `make build`; `0.0.0-dev` si se compila sin ldflags (`main.go:29`) | siempre `VERSION` exacta | media | [ASSUMED] |
| 8 | ¿Flag `--version`? | sí, imprime `sdd <version>` y sale; sirve para verificar el binario publicado | solo `/api/health` y el banner | alta | [ASSUMED] |
| 9 | ¿Formato de los assets? | binarios crudos `sdd_<v>_<os>_<arch>[.exe]`, sin tarballs (un archivo por plataforma, Gaps del contrato) | `.tar.gz` y `.zip` por plataforma | media | [ASSUMED] |
| 10 | ¿Notas del release? | `--generate-notes` de GitHub (títulos de PRs mergeados) | CHANGELOG manual | alta | [ASSUMED] |
| 11 | ¿Cómo se cross-compila? | un job con loop en `make release-assets`, reutilizable en local | matrix de jobs y artifacts | alta | [ASSUMED] |
| 12 | ¿`release.yml` re-corre tests? | sí, `npm test`, `go vet` y `go test` antes de compilar; no depende de `ci.yml` | `workflow_run` encadenado al CI | media | [ASSUMED] |
| 13 | ¿Dónde va el check de PR? | job `Version` dentro de `ci.yml` (un solo workflow de PR, nombre estable para marcarlo requerido) | workflow aparte `version-check.yml` | media | [ASSUMED] |
| 14 | ¿Concurrencia de releases? | `concurrency: release` sin `cancel-in-progress`: dos merges seguidos se encolan | sin concurrencia | alta | [ASSUMED] |
| 15 | ¿Quién crea el tag? | `gh release create v<v>` lo crea en el commit de merge con `GITHUB_TOKEN` (`contents: write`) | `git tag` + `git push --tags` separados | alta | [ASSUMED] |
| 16 | ¿De dónde salen los tags para comparar? | `git tag --list 'v*'` del checkout (en CI con `fetch-depth: 0` y `fetch-tags: true`) | `gh api repos/.../tags` | alta | [ASSUMED] |
| 17 | ¿Strip de los binarios? | `-s -w` en los assets del release; `make build` sin strip | sin strip | media | [ASSUMED] |
| 18 | ¿Versiones de las actions? | las del `ci.yml` actual (`checkout@v4`, `setup-node@v4`, `setup-go@v5`) | pinear por SHA | alta | [ASSUMED] |

## Verificabilidad
Mixta. CA-1 a CA-3 ALTA: lógica pura y lectura de archivos, cubiertas por `go test -count=1 ./...` (verificado 2026-10-29 en el contrato) con tests table-driven y un repo git en `TempDir`, como ya hace `internal/scan`. CA-4 y CA-5 MEDIA: exigen `make build` y `make release-assets` y observar el binario (`--version`, `/api/health`, `sha256sum -c`); el contrato llega hasta «app levantada», así que entran en la escalera. CA-6 MEDIA: el job corre sobre el PR que abre el run y se lee con `gh pr checks`, sin tocar ningún workflow a mano. CA-7 NULA: un `push` a `main` solo ocurre al mergear, y el contrato prohíbe disparar workflows; antes del merge queda la lectura estática del YAML (BAJA, sin `actionlint` ni parser YAML en la máquina). CA-8 BAJA: grep sobre README y contrato. Políticas: `max-lineas-archivo` 300 holgado; `no-panic` aplica al tool nuevo; `commits` en patrón; `gofmt`. Ninguna política de tamaño ni dependencias activa; no se suman dependencias (todo stdlib).

## Plan de verificacion
- **CA-1**: `go test -count=1 ./cmd/versioncheck -run TestVersionFile` lee `../../VERSION` y espera `0.1.0`. Alternativa más barata: `grep -Ex '[0-9]+\.[0-9]+\.[0-9]+' VERSION`.
- **CA-2**: `go test -count=1 ./internal/semver` con tablas para `Parse` (válidos e inválidos listados en el CA), `Compare` y `Latest`. Alternativa más exhaustiva: fuzz de `Parse` con `go test -fuzz`.
- **CA-3**: `go test -count=1 ./cmd/versioncheck`: `TestCheck_*` sobre `check(version, tags)` (mayor, igual, menor, inválida, sin tags) y `TestListTags_RepoReal` con `git init` + `git tag` en `t.TempDir()`. Después, en la raíz: `go run ./cmd/versioncheck` → exit 0 y `VERSION 0.1.0, sin tags previos`. Alternativa más barata: solo el test del núcleo puro.
- **CA-4**: `make build && ./bin/sdd --version` → `sdd 0.1.0-dev`, exit 0, sin puerto abierto; `./bin/sdd --port 4322 --dir testdata/workspace --no-github` en background y `curl -sf :4322/api/health | jq -r .version` → `0.1.0-dev`; `go build -ldflags "-X main.version=9.9.9" -o /tmp/sdd-x ./cmd/sdd && /tmp/sdd-x --version` → `sdd 9.9.9`. Alternativa más exhaustiva: test de `cmd/sdd` que ejecute el binario con `go run . --version`.
- **CA-5**: `make release-assets && ls dist/release` → 7 archivos con los nombres del CA; `(cd dist/release && sha256sum -c checksums.txt)` → todo `OK`; `./dist/release/sdd_0.1.0_linux_amd64 --version` → `sdd 0.1.0`; `file dist/release/sdd_0.1.0_darwin_arm64` dice Mach-O arm64 y el `.exe` PE32+. Con `VERSION` roto (`echo x > VERSION`), `make release-assets` sale distinto de 0 (restaurar después). Alternativa más barata: solo `ls` y `sha256sum -c`.
- **CA-6**: tras abrir el PR del run, `gh pr checks <n> --watch` muestra el job `Version` en verde; el run no edita tags ni dispara nada a mano. Alternativa: revisión estática del YAML.
- **CA-7**: estático: `grep` de `on: push`/`branches: [main]`, `concurrency`, `permissions`, el orden de pasos y `gh release create` en `release.yml`; `node -e` no puede parsear YAML sin dependencias, así que no hay parse. Protocolo humano, después de mergear el PR: 1. Actions → «Release» corrió sobre `main` y está verde. 2. `gh release view v0.1.0 --repo chichex/sdd` lista 7 assets y el tag `v0.1.0` apunta al commit de merge. 3. `gh release download v0.1.0 -p 'sdd_0.1.0_linux_amd64' && chmod +x sdd_0.1.0_linux_amd64 && ./sdd_0.1.0_linux_amd64 --version` → `sdd 0.1.0`. 4. Abrir un PR de prueba sin tocar `VERSION`: el job `Version` queda rojo con «0.1.0 no es mayor que v0.1.0». 5. Settings → Branches → `main` → marcar `Version` como required status check. 6. Mergear un PR con `VERSION` en `0.1.1`: aparece el release `v0.1.1`.
- **CA-8**: `grep -n 'Versión y releases' README.md`, `grep -n 'release.yml' .sdd/project.md` en Limites, Ambientes y Gaps.

## Riesgos y gaps
- **Conflicto con el PR #11** (`chore/sdd-init-2026-10-29`): toca `.sdd/project.md` y `ci.yml`, los mismos archivos que CA-6 y CA-8. Aterrizar #11 antes del run, o correr el run con `--base chore/sdd-init-2026-10-29` y aterrizar en orden.
- **CA-7 solo se prueba mergeando**: el primer release real valida el workflow; si falla, el fix es otro PR con bump de `VERSION`. Mitigación: `make release-assets` y `versioncheck` se prueban en local y son los mismos pasos que corre el workflow.
- **PRs abiertos antes del check** (hoy #11) no bumpean `VERSION`: si aterrizan después de esta spec, `release.yml` queda rojo en `versioncheck` sin publicar nada, por diseño. Bumpearlos antes de mergear o aceptar un run rojo.
- **El check no bloquea hasta marcarlo requerido**: decisión del humano en branch protection (fuera de alcance). Mientras tanto un merge sin bump deja `release.yml` rojo.
- **Inferencia 6 (lógica en Go)** define cómo se verifica CA-3: si se prefiere shell, CA-3 baja a MEDIA (test ejecutando el script).
- **Límites del contrato**: la spec cambia la línea «no crear releases ni tags» por decisión del usuario; el run nunca crea tags ni releases, solo el workflow al mergear.
