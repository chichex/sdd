# Spec — Vista consistente de GitHub y trabajos en el workspace
<!-- Generada por /sdd-spec el 2026-10-29. Fuente: pedido libre. Estado: implementada -->
<!-- SDD-Tracking: version=1; type=spec; state=implemented; issue=none; grill=none; superseded-by=none -->

## Contexto
`TestDegradado_UnRepoFallaYLosDemasNo` (`internal/api/github_test.go:149`) falla de forma intermitente: en el CI de `main` (`0ba1dca`, run 40983780319) y en el PR #12, y en local 4 de 6 veces con una sola CPU (`taskset -c 0`). El test espera a que `/api/workspace` traiga `github.fetchedAt` y entonces exige que el PR 239 de waica sea `sin dato` y `stale`. La causa está en `internal/workspace`: `Snapshot()` (`workspace.go:171`) calcula el estado de GitHub en vivo con `githubStatus()` a partir de `s.gh`, `s.ghStatus` y `s.polled`, pero devuelve `s.trabajos` del último `rebuild()`. `applyGitHub` (`github.go:47`) guarda cada snapshot de repo bajo su propio lock, recién después pone `s.polled = true` y por último reconstruye: entre medio, una lectura ve `fetchedAt` de platform con trabajos armados antes del poll, donde nada es stale (`inputFor`, `github.go:116`). No es un problema del test: la API puede servir por unos milisegundos un estado de GitHub que no corresponde a los trabajos que muestra.

## Comportamiento esperado
- **CA-1 (ALTA)** — `Snapshot()` devuelve `GitHub`, `RepoGitHub` y `Trabajos` calculados en el mismo `rebuild()`: el estado de GitHub se computa y se guarda en el servicio dentro de la misma sección crítica en que se guardan los trabajos, y `Snapshot()` lee los tres campos ya guardados sin recalcular nada. `applyGitHub` aplica todos los resultados del poll y `polled = true` bajo un solo lock antes de reconstruir. Invariante observable: para cualquier `Snapshot()`, si `RepoGitHub[id].OK` es falso tras el primer poll, todos los nodos con PR de ese repo son `stale`; y si `GitHub.FetchedAt` no está vacío, `polled` ya era verdadero al armar esos trabajos. Test nuevo en `internal/workspace` que aplica un resultado con un repo caído y otro ok mientras un goroutine toma `Snapshot()` en loop y comprueba el invariante en cada lectura; corre verde con `-race`.
- **CA-2 (ALTA)** — `TestDegradado_UnRepoFallaYLosDemasNo` queda con sus aserciones y su condición de espera (`FetchedAt != nil`) intactas y pasa 20 de 20 veces con `taskset -c 0 go test -count=20 -race -run 'TestDegradado' ./internal/api`. Los demás `TestDegradado_*` (caché mientras corre la primera consulta, sin `gh`) siguen verdes: el `fetchedAt` de la caché cargada por `loadCaches` sigue siendo visible antes del primer poll.
- **CA-3 (ALTA)** — Regresión completa verde: `go test -count=1 -race ./...` y `make test`; sin cambios en `web/`. `Refresh()` sigue devolviendo el `fetchedAt` global vigente.

## Plan de entrega
| Capa | Etapa | CAs | Justificacion |
|---|---|---|---|
| 1 | fix | CA-1 a CA-3 | Un solo PR sobre `main`: el cambio vive en `internal/workspace` (dos funciones y un test nuevo) y se verifica con `go test`. |

## Fuera de alcance
- Cambiar qué significa `stale` o cómo se calcula por nodo (`internal/model`).
- Tocar el contrato de la API (`internal/api/views.go`) o el front.
- Otros tests intermitentes que no sean este.

## Inferencias
| # | Inferencia | Eleccion propuesta | Alternativa razonable | Confianza | Resolucion |
|---|---|---|---|---|---|
| 1 | ¿Se arregla el servicio o el test? | El servicio: la vista inconsistente es un bug real de la API, el test solo lo detecta | Que el test espere hasta que el nodo sea stale (oculta el bug) | alta | [ASSUMED] |
| 2 | ¿Dónde se calcula el estado de GitHub? | En `rebuild()`, guardado en campos del servicio junto a `trabajos`; `Snapshot()` solo lee | Calcular en `Snapshot()` y reconstruir trabajos bajo el mismo lock | media | [ASSUMED] |
| 3 | ¿Cómo se aplica el poll? | Todos los repos y `polled` bajo un único lock, después un solo `rebuild()` | Mantener locks por repo y reconstruir al final | alta | [ASSUMED] |
| 4 | ¿`loadCaches` dispara reconstrucción? | Solo si hace falta para que el `fetchedAt` cacheado sea visible antes del primer poll; si el primer `rebuild()` del escaneo ya ocurre después de `loadCaches`, no se agrega | `rebuild()` explícito al final de `loadCaches` | media | [ASSUMED] |
| 5 | ¿Forma del test nuevo? | Test de paquete (`package workspace`) con un goroutine lector y el invariante por lectura, bajo `-race` | Test por la API con `taskset` repetido | alta | [ASSUMED] |
| 6 | ¿Base del PR? | `main` (`0ba1dca`), independiente de #11 y #12 | Apilarlo sobre #12 | alta | [ASSUMED] |

## Verificabilidad
ALTA en los tres CA: todo se observa con `go test -count=1 ./...` (verificado 2026-10-29 en el contrato) y `-race`, sin levantar la app. El invariante de CA-1 se comprueba en un test de paquete que nunca falla en falso: solo falla si alguna lectura viola el invariante. CA-2 usa la repetición con una CPU, que hoy reproduce la falla en 4 de 6 corridas. Políticas: `no-panic`, `gofmt`, `max-lineas-archivo` (`workspace.go` tiene ~200 líneas, `github.go` ~170), `commits`; sin dependencias nuevas.

## Plan de verificacion
- **CA-1**: `go test -count=1 -race ./internal/workspace -run TestSnapshot_Consistente` (o el nombre elegido) → verde; rojo primero, sobre el código actual, para probar que detecta la carrera (con `taskset -c 0` si hace falta). Alternativa más barata: solo el test por la API de CA-2.
- **CA-2**: `taskset -c 0 go test -count=20 -race -run 'TestDegradado' ./internal/api` → `ok` 20/20; `git diff -- internal/api/github_test.go` vacío. Alternativa más exhaustiva: `-count=100` sin `-race`.
- **CA-3**: `go test -count=1 -race ./...` y `make test` → verdes; `gofmt -l .` vacío; `git grep -n 'panic(' -- 'internal/workspace/*.go'` vacío.
- Señal extra, sin disparar nada a mano: el job `Binario` del CI del PR de este run queda verde (`gh pr checks`).

## Riesgos y gaps
- **Caché antes del primer poll**: si el estado de GitHub pasa a calcularse en `rebuild()`, el `fetchedAt` que `loadCaches` carga tiene que entrar en alguna reconstrucción antes del primer poll; `TestDegradado_CacheMientrasCorre` lo cubre (inferencia 4).
- **`-race` en CI**: el contrato corre `go test ./...` sin `-race`; el test nuevo tiene que ser determinista también sin la bandera.
- **Inferencia 1** cambia el servicio y no el test: si se prefiere lo mínimo, el arreglo se reduce a la condición de espera del test y CA-1 se descarta.

## Desviaciones
- 2026-10-29 [DEVIATION] CA-1, invariante: la cláusula «si `GitHub.FetchedAt` no está vacío, `polled` ya era verdadero al armar esos trabajos» vale para un servicio sin caché. Con `CacheDir`, `loadCaches` deja visible un `fetchedAt` cacheado antes del primer poll a propósito (CA-2). El test nuevo corre sin caché y comprueba la cláusula como «con `fetchedAt` presente, el repo caído figura caído y sus PRs son stale». No cambia el alcance.
- 2026-10-29 [DEVIATION] CA-2 y Riesgos: `TestDegradado_CacheMientrasCorre` no existe. El test de la caché mientras corre la primera consulta es `TestCache_ArrancaSinGhYSirveLoCacheado` (`internal/api/github_test.go:111`) y `-run 'TestDegradado'` no lo selecciona: corre `TestDegradado_UnRepoFallaYLosDemasNo` y `TestDegradado_SinGhSirveLoLocal`. Se corrió además con `-run 'TestDegradado|TestCache'`. No cambia el alcance.
- 2026-10-29 [DEVIATION] Inferencia 4: `loadCaches` no reconstruye. Con la caché cargada y sin datos de repo todavía, un `rebuild()` mostraría el `fetchedAt` cacheado sin los trabajos armados con esa caché, y `TestCache_ArrancaSinGhYSirveLoCacheado` exige la idea #210 junto a ese `fetchedAt`. El `fetchedAt` cacheado entra a la vista con el primer `rebuild()` del escaneo. Además el estado de GitHub se calcula en `New()` y al descubrir los repos (`scanAll`), para que `github` y `repos[].github` no queden en cero hasta la primera lectura (con `--no-github` siguen diciendo `desactivado`): esos dos puntos no guardan trabajos. No cambia el alcance.
- 2026-10-29 [DEVIATION] CA-1, plan de verificación: el test nuevo no llama `applyGitHub` a mano. Levanta el servicio con `Run` y un `fake.Runner` con `Fail` y `Gate`, y el poll entra por el `OnResult` del `Poller`; el lector toma `Snapshot()` en loop desde antes del poll. Es el mismo camino que usa la API, un seam más alto. `Snapshot()` clona `RepoGitHub` (copia en la frontera) para que un llamador no pueda mutar el mapa que guarda `rebuild()`. No cambia el alcance.

## Resultado de ejecucion (2026-10-29 · HEAD f30780e)
| CA | Capa | Estado | Evidencia |
|---|---|---|---|
| CA-1 | 1/1 | verificado | `go test -count=1 -race ./internal/workspace -run TestSnapshot -v`: `TestSnapshot_GitHubYTrabajosCoincidenAlAplicarUnPoll` ok, 2 subtests (cae waica / cae platform), 10 rondas cada uno (cada ronda: un servicio nuevo, un lector en loop y un poll con un repo caído). Rojo previo sobre el código de `main`, 20/20 con `go test -count=20 -run TestSnapshot ./internal/workspace`, también con `taskset -c 0` y con `-race`: `github.fetchedAt = "…" pero chichex/waica figura ok: el poll se ve a medias`. Verde en HEAD f30780e con el mismo `-count=20` en las cuatro variantes (normal 5,9 s, `taskset -c 0` 20,8 s, `-race` 12,8 s, ambas 26,6 s). |
| CA-2 | 1/1 | verificado | `taskset -c 0 go test -count=20 -race -run 'TestDegradado' ./internal/api -v`: `ok`, `TestDegradado_UnRepoFallaYLosDemasNo` 20/20 y `TestDegradado_SinGhSirveLoLocal` 20/20. Con `-run 'TestDegradado\|TestCache'` también `TestCache_ArrancaSinGhYSirveLoCacheado` 20/20. `git diff origin/main -- internal/api/github_test.go` vacío. Rojo previo: `taskset -c 0 go test -count=6 -run 'TestDegradado_UnRepoFallaYLosDemasNo$' ./internal/api` falló 4 de 6 con `PR 239 de waica = "sin dato" stale false`; con el arreglo, 6/6. |
| CA-3 | 1/1 | verificado | `go test -count=1 -race ./...`: 8 paquetes ok (más 2 sin tests), 64 tests de primer nivel (63 en la base, 1 nuevo) y 188 subtests, 0 FAIL. `make test` ok. `go test -count=1 ./...` ok. `git diff --name-status origin/main..HEAD` no toca `web/`. `Refresh()` devuelve `ghGlobal.FetchedAt`, el mismo valor que antes calculaba en vivo (`TestRefresh_ConsultaYResponde202` ok). |
| POL-commits | 1/1 | verificado | `git log --format=%s origin/main..HEAD \| grep -vE '<patrón del contrato>'` vacío (4 commits al medir). |
| POL-max-lineas-archivo | 1/1 | verificado | `git diff --name-only --diff-filter=AM origin/main...HEAD -- '*.go' '*.ts' '*.tsx' \| xargs -r wc -l \| awk '$2!="total" && $1>300'` vacío (el más largo: `internal/workspace/workspace.go`, 229). |
| POL-no-any-ts | 1/1 | verificado | `cd web && npx oxlint -D typescript/no-explicit-any`: exit 0 (sin TypeScript tocado). |
| POL-gofmt | 1/1 | verificado | `gofmt -l .`: vacío. |
| POL-no-panic | 1/1 | verificado | `git grep -n 'panic(' -- '*.go' ':!cmd/' ':!*_test.go' ':!web/embed.go'` y `git grep -n 'panic(' -- 'internal/workspace/*.go'`: vacíos. |

Regresión y escalera, en HEAD f30780e salvo lo indicado: `gofmt -l .` vacío; `go vet ./...` exit 0; `cd web && npx tsc -b` exit 0; `npm run lint` exit 0 (3 warnings, en archivos que este run no toca); `cd web && npm test`: 12 archivos, 176/176 (sin cambios en `web/`); `make build` exit 0 (`bin/sdd` de 11 MB con el front embebido). App levantada (`./bin/sdd --port 4322 --dir testdata/workspace --no-github`, proceso matado): `/api/health` 200, `/api/workspace` con `scanning:false` y `github` = `{ok:false, reason:"desactivado", fetchedAt:null}`, `?mostrar=otro` → 400, `/t/abc` → 200 `text/html`, `POST /api/refresh` → 202. Sin verificar: el comportamiento contra GitHub real y contra el workspace real (por contrato solo fixtures). Fuera de alcance y sin tocar: `Snapshot().Repos` sigue leyendo `s.data` en vivo mientras `store()` y `rebuild()` toman el lock por separado, así que `Repos` puede adelantarse a `Trabajos` por un instante; la spec limita CA-1 a `GitHub`, `RepoGitHub` y `Trabajos`.
