# Spec — Vista consistente de GitHub y trabajos en el workspace
<!-- Generada por /sdd-spec el 2026-10-29. Fuente: pedido libre. Estado: aprobada -->
<!-- SDD-Tracking: version=1; type=spec; state=approved; issue=none; grill=none; superseded-by=none -->

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
