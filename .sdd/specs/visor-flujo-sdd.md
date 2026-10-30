# Spec — Visor del flujo SDD
<!-- Generada por /sdd-spec el 2026-10-27. Fuente: grill 2026-10-23-visor-flujo-sdd. Estado: implemented -->
<!-- SDD-Tracking: version=1; type=spec; state=implemented; issue=none; grill=2026-10-23-visor-flujo-sdd; superseded-by=none -->

## Contexto
El grill `.sdd/grills/2026-10-23-visor-flujo-sdd.md` (finalizado, 34 decisiones) define una herramienta local de solo lectura que muestra cómo cada *trabajo* avanza por Idea → Grill → Spec → Run → PR → Mergeado → Publicado en todos los repos de una carpeta. Hoy el repo es un esqueleto: `cmd/sdd/main.go:34-37` sirve `/api/health` y el fallback de SPA, `web/src/App.tsx` solo consulta `/api/health`, `web/src/index.css` es `@import "tailwindcss"` y los únicos tests son los dos de `cmd/sdd/main_test.go`. No existen `internal/`, `testdata/`, tokens en código ni runner de tests del front. La dirección visual ya está fijada en `web/DESIGN_SYSTEM.md` (`sobrio-indigo`), sin materializar.

Fuentes que el visor consume, todas definidas fuera de este repo y que esta spec porta a Go sin modificarlas: `skills/docs/sdd-events-v1.md` (log `<git-common-dir>/sdd/events.jsonl`, 11 eventos, 16 códigos de diagnóstico, 12 vectores de conformance), `skills/docs/sdd-tracking-v1.md` (markers `SDD-Tracking v1` de spec, grill y contrato, 3 vectores, compatibilidad legacy) y la referencia TypeScript `skills/pi-extensions/sdd-artifacts/` (`parseSddEventLine` en `sdd-events.ts:327`, `parseSddArtifact` en `index.ts:726`, 21 fixtures Markdown). El gate de «PR validado» es la Fase 3 de `skills/claude/sdd-land/SKILL.md`, con su filtro jq de deduplicación y 7 rollups grabados en `skills/pi-extensions/sdd-land-gate/fixtures/`. Datos reales para fixtures: los logs de `dale-que-sale` (17 líneas, 3 trabajos), `waica` (21, con un run `interrupted` reintentado con los mismos 5 PRs y un `land.merged` sin issue), `platform` (22, con un stack de 4 PRs mergeado 3/4) y `skills` (4). El workspace real tiene 12 `.sdd/project.md`, 2 worktrees y 2 carpetas sin git; los specs y grills viven casi todos en `.sdd/` del árbol (platform tiene 1 grill en el common-dir).

## Diseño
- **Superficie**: `web` (raíz `web/src`), estado `con sistema` por `web/DESIGN_SYSTEM.md`; tokens y componentes todavía no existen en código (`.sdd/project.md`, `## Diseño`). shadcn/ui decidido (grill D30) y no instalado. Esta spec es el primer run que toca UI: materializa los tokens bajo `@theme` en `web/src/index.css` y los componentes base en `web/src/components/ui/`.
- **Pantallas y flujos**: (1) Board en desktop (≥768 px): siete carriles fijos, una tira por trabajo, ramas dentro de la tira. (2) Lista de trabajos en celular (<768 px) y en desktop con filtro «todos» si el board no entra. (3) Detalle de un trabajo en `/trabajos/<id>`: Dialog en desktop, Sheet desde abajo en celular, con el grafo expandido (fechas y links) y el Gantt por etapa. (4) Navegación: cabecera de una fila en desktop; hamburguesa + Sidebar en celular.
- **Estados**: carga progresiva (los repos aparecen a medida que se escanean; skeleton con la forma del board hasta el primer snapshot), vacío (sin `.sdd/project.md` bajo la carpeta → EmptyState de página con el comando), degradado sin GitHub (Banner `warning` persistente, Idea vacía, PRs `sin dato`, nodos `stale`), error de lectura (Banner `error` con reintento, último dato bueno conservado), snapshot viejo (antigüedad visible en el Refresh).
- **Componentes**: todos nacen en este run siguiendo `web/DESIGN_SYSTEM.md` por nombre: de shadcn/ui `Button`, `Select`, `Badge`, `Dialog`, `Sheet`, `Tooltip`, `Skeleton`, `Toast`; propios `Navegación`, `Sidebar`, `Refresh`, `Segmented`, `Chip de referencia`, `Nodo`, `Carril`, `Tira de trabajo`, `Lista de trabajos`, `Banner`, `EmptyState`, `Grafo` (SVG) y `Gantt` (SVG). Tokens: solo los de la tabla `### Color` del doc, claro y oscuro, con `prefers-color-scheme` y `data-theme`.
- **Wireframe board (desktop)**:

```text
┌ sdd  [Todos los repos ▾]  (Activos|Todos)   ⟳ 2 min   ☼ ◐ ☾ ┐
│ ⚠ Sin GitHub: mostrando solo lo local            [Reintentar] │
├────────┬────────┬────────┬─────────┬─────────┬─────────┬───────┤
│ IDEA 3 │GRILL 1 │ SPEC 2 │ RUN   1 │ PR    5 │ MERGE 5 │ PUB 0 │
├────────┼────────┼────────┼─────────┼─────────┼─────────┼───────┤
│        │● #203  │● #203  │● run 2  │● 1/5 ✓  │● 1/5    │       │  ← tira #203
│        │ final. │ impl.  │ con PR  │● 2/5 ✓  │● 2/5 …  │       │
│        │        │        │○ run 1  │  …      │         │       │  ← rama fallada
│        │        │        │ interr. │         │         │       │
├────────┼────────┼────────┼─────────┼─────────┼─────────┼───────┤
│● #57   │        │        │         │         │         │       │  ← tira Idea
└────────┴────────┴────────┴─────────┴─────────┴─────────┴───────┘
```

- **Wireframe lista (celular)**:

```text
┌ ☰  sdd · waica                       ⟳ 2 min ┐
│ ⚠ Sin GitHub: solo lo local      [Reintentar] │
├───────────────────────────────────────────────┤
│ ● #205 Fade and blend      [mergeado]    1 d  │
│ ● #203 3D shadows          [mergeado]    3 d  │
│ ● #57  Backlog norte…      [abierto]     2 h  │
└───────────────────────────────────────────────┘
```

- **Wireframe detalle**:

```text
┌ #203 3D shadows · chichex/waica                         [×] ┐
│ Grafo                                                        │
│ (Idea)─(Grill final.)─(Spec impl.)─┬─(Run 1 · interrumpido)  │
│                                    └─(Run 2)─(PR 1/5)…(PR 5/5)─(Merge 1/5)…│
│ Gantt      oct 15     oct 17     oct 20     oct 23   ┆ hoy   │
│ Grill      ▬▬                                                │
│ Spec         ▬▬▬                                             │
│ Run 1            ▬  interrumpido                             │
│ Run 2            ▬▬▬                                         │
│ PR 1/5…5/5          ▬▬▬▬▬▬▬▬▬▬▬▬▬▬                           │
│ Mergeado                           ▬                         │
└──────────────────────────────────────────────────────────────┘
```

- **Plataformas y accesibilidad**: desktop y celular por LAN; teclado básico (Tab a todo, flechas en el board, Enter abre el detalle), contraste AA de los tokens, claro y oscuro (grill D32). Web en celular: viewport sin bloquear zoom, inputs ≥16 px, `safe-area-inset-*`, 44 px táctiles (reglas de `coding-policies.md`, sección React «Web en celular y webview», y de `DESIGN_SYSTEM.md`).
- **Capturas de referencia**: ninguna. El handoff no trae mock ni referencia renderizable; la referencia es el doc del design system.

## Comportamiento esperado
Grados: **ALTA** test Go determinista que el contrato corre hoy; **BAJA** solo señales estáticas hoy (typecheck, lint, build), sería ALTA con el runner del front que la capa 4 suma; **NULA** exige humano.

**Parsers (capa 1)**
- **CA-1** `internal/events`: `ParseLine(line []byte)` devuelve exactamente uno de `Event` válido, `UnsupportedVersion` o `Invalid` con uno de los 16 códigos del doc; los 12 vectores de `sdd-events-v1.md` (copiados verbatim a `testdata/vectors/events-v1.jsonl`) parsean válidos y `Serialize(event)` reproduce sus bytes exactos. Un cambio en el bloque de vectores del doc rompe el test. — ALTA
- **CA-2** `internal/events`: cada entrada de una tabla de líneas inválidas produce su código: espacio extra o `A` → `non-canonical`; claves faltantes, extra, repetidas o desordenadas → `missing-key`/`extra-key`/`duplicate-key`/`key-order`; >1024 bytes → `line-too-long`; `"v":2` → `unsupported-version`; `ts` fuera de calendario → `invalid-ts`; harness, evento, issue, reason desconocidos → su código; `%2f` en minúscula, `/` crudo o `none` codificado → `invalid-reference`; `pr` con `0`, negativo o `land.merged` con dos PRs → `invalid-pr`/`invariant-violation`. `ReadLog(r)` salta las inválidas contando cuántas, sigue leyendo y descarta un segmento final sin `LF`. — ALTA
- **CA-3** `internal/tracking`: `Parse(markdown)` devuelve metadata canónica (`kind=canonical`) para los 3 vectores de `sdd-tracking-v1.md` y clasifica los 21 fixtures de `skills/pi-extensions/sdd-artifacts/fixtures/` (copiados a `testdata/tracking/` con un `expected.json`) en canónico válido, ausente, inválido, conflicto o legacy con estado normalizado: preámbulo hasta el primer `##` fuera de fences, BOM conservado, duplicados idénticos → válido + `duplicate-canonical`, divergentes → conflicto, aliases de `Estado:`/`Status:` de la tabla legacy (`aprobada`, `pendiente de ejecución`, `reemplazada por <ref>`…) → estado normalizado con `legacy-explicit`, H2 `Resultado de ejecucion` → `implemented` `legacy-inferred`, tipo por H1 `Spec`/`Grill`/`Contrato de autonomia`. `Serialize` solo para metadata canónica, con round-trip byte a byte de los vectores. — ALTA
- **CA-4** `internal/tracking`: `DecodeRef`/`EncodeRef` cumplen la codificación de referencias: `%2393` ↔ `#93`, `.sdd%2Fspecs%2Fx.md` ↔ `.sdd/specs/x.md`, `session%209%2F%C3%A1` ↔ `session 9/á`; `%2f`, `%C3` truncado, UTF-8 inválido, no-ASCII crudo y `%6E%6F%6E%65` (`none` codificado) son inválidos. — ALTA

**Escaneo, modelo y API local (capa 2)**
- **CA-5** `sdd --dir D` descubre repos: carpetas hasta profundidad 3 bajo `D` con `.sdd/project.md`, salteando `node_modules`, `.git`, `vendor`, `dist` y `build`; `id` = `owner/repo` del remote `origin` (formas `https://github.com/o/r.git`, `git@github.com:o/r.git` y `ssh://git@github.com/o/r.git`), o el nombre de la carpeta si no hay remote; dos worktrees con el mismo `git-common-dir` son un solo repo que lista ambas rutas; una carpeta con `.sdd/project.md` sin `.git` aparece en `repos[]` como `sinGit: true` y no se procesa. Test con un workspace temporal armado con `git init`, `git worktree add`, una carpeta sin git y una a profundidad 4. — ALTA
- **CA-6** Por repo se leen `<common-dir>/sdd/events.jsonl` (líneas inválidas salteadas y contadas en `repos[].lineasInvalidas`), grills de `.sdd/grills/*.md` y `<common-dir>/sdd/grills/*.md`, specs de `.sdd/specs/*.md` y `<common-dir>/sdd/specs/*.md` (ante el mismo nombre gana el árbol), cada Markdown parseado con `internal/tracking`; un directorio ausente es vacío, no error; un archivo ilegible se reporta en `repos[].errores` sin frenar el escaneo. — ALTA
- **CA-7** Agrupación en trabajos: eventos y artefactos se agrupan por ancla con prioridad issue > grill > spec; `#NN` y `owner/repo#NN` del mismo repo son el mismo ancla; un `land.merged` con `issue=none` entra al trabajo cuyo `run.pr-opened` listó ese PR; un PR sin spec ni ancla queda fuera. Con los cuatro logs reales copiados a `testdata/events/<repo>.jsonl` el resultado es: `dale-que-sale` 3 trabajos (#287, #289, #288); `waica` 2 trabajos (#203 con 5 PRs, #205 con 5 PRs) y el PR 249 excluido; `platform` 6 (#54, #101, #102, #57, #56, #100); `skills` 1 (#100). — ALTA
- **CA-8** Etapas y sub-estados derivados solo de lo local: Grill `pausado`/`finalizado`; Spec `borrador`/`aprobada`/`implementada`/`reemplazada` (el marker del artefacto más nuevo manda sobre los eventos; `spec.moved` une las dos identidades; una `reemplazada` es una rama cerrada); Run `corriendo` tras `run.started`, `fallado (aborted|interrupted)` tras `run.failed` como rama cerrada, `con PR` tras `run.pr-opened`; un `run.started` posterior a un `run.failed` abre una rama nueva; PR: un nodo por PR en orden de capa `n/N`, sub-estado `sin dato` sin GitHub; Mergeado: un nodo por PR con `land.merged`; Publicado: ver CA-20. Etapa actual = la más avanzada de la rama viva; `ultimaActividad` = mayor `ts`. Con el fixture de waica, #203 tiene dos ramas de Run (la primera `fallado (interrupted)`) y 5 nodos Mergeado; con platform, #54 queda en Mergeado 3/4 con el PR 98 en `sin dato`. — ALTA
- **CA-9** `GET /api/workspace` responde JSON `{root, scannedAt, scanning, github:{ok,reason,fetchedAt}, repos:[{id,nombre,rutas,sinGit,remote,lineasInvalidas,errores}], trabajos:[…]}` con cada trabajo `{id, repo, titulo, ancla, etapa, subEstado, ultimaActividad, terminado, ramas:[{id, estado, nodos:[{etapa, subEstado, ts, refs:{issue,grill,spec,pr}, url}]}]}`, ordenado por `ultimaActividad` descendente; `?filtro=activos|todos` (default `activos`, que excluye `terminado=true`) y `?repo=<id>` filtran; mientras el escaneo inicial corre, responde el parcial con `scanning:true` y los repos ya leídos. Verificado con `httptest` sobre el workspace de fixtures y con la app levantada. — ALTA
- **CA-10** `GET /api/trabajos/{id}` devuelve el trabajo con, además, `gantt:[{rama, etapa, inicio, fin, estado}]` (inicio = primer evento de la etapa en esa rama, fin = primer evento de la etapa siguiente o `null` si sigue abierta) y `links` (issue, PRs, rutas locales de grill y spec); un id desconocido responde `404` con `{error}`. — ALTA
- **CA-11** Trabajos anteriores al log (grill D19): un grill o spec con marker y sin eventos produce un trabajo con un solo nodo en su etapa y sub-estado del marker, `ts` = mtime del archivo y `sinHistoria: true`; el Gantt de ese trabajo no tiene barras. — ALTA
- **CA-12** Refresh local en vivo: el server relee el log y los artefactos de un repo a lo sumo 2 s después de un cambio (polling de mtime con stdlib, `--watch-interval`, default `2s`) y actualiza `scannedAt`; test: agregar una línea al log del fixture y esperar hasta 10 s a que el nodo aparezca en `/api/workspace`. — ALTA
- **CA-13** Workspace vacío (sin repos bajo `--dir`): `repos: []`, `trabajos: []`, `scanning: false`, y el server sigue sirviendo; sin GitHub que consultar. — ALTA
- **CA-14** LAN (grill D34): con `--host 0.0.0.0` la salida estándar imprime `abrí http://<IP LAN>:<port>` con la primera IPv4 no loopback y un QR ASCII de esa URL (ancho ≤ 33 módulos); con el default `127.0.0.1` imprime solo la URL local y ningún QR; el aviso «sin autenticación» se mantiene. Test sobre la función que arma el mensaje con interfaces inyectadas. — ALTA

**GitHub (capa 3)**
- **CA-15** `internal/github`: todo acceso a GitHub pasa por un `Runner` que ejecuta `gh` como subproceso; los tests usan un fake que sirve respuestas grabadas de `testdata/gh/<repo>/*.json` y falla si recibe una invocación no grabada; ninguna invocación contiene `mutation`, `-X POST`/`PUT`/`PATCH`/`DELETE` fuera de `gh api graphql` de consulta, ni subcomandos `edit`, `create`, `merge`, `close`, `comment`. — ALTA
- **CA-16** Idea (grill D23): cada issue abierto del repo sin artefactos SDD es un trabajo en Idea con `titulo`, número, `url` y `ultimaActividad` = `updatedAt`; un issue abierto que ya ancla un trabajo le aporta título y url sin duplicarlo; un issue cerrado sin artefactos no aparece. — ALTA
- **CA-17** Sub-estado de PR con el gate de `sdd-land`: `draft` si `isDraft`; `cambios pedidos` si `reviewDecision=CHANGES_REQUESTED`; `checks rojos` si el filtro deduplicado da `detenido`; `validado` si está abierto, no draft, `mergeable≠CONFLICTING`, sin `CHANGES_REQUESTED`, checks `verde` y cero threads sin resolver; si no, `abierto`. La deduplicación porta el filtro jq de la Fase 3 (identidad `workflowName / name` o `context`, más reciente por `startedAt`/`completedAt`, `SKIPPED`/`NEUTRAL` nunca supersede un run real, empate → conservador) y los 7 rollups de `skills/pi-extensions/sdd-land-gate/fixtures/` dan el mismo `estado` que ese gate (`brik-117` detenido, `cancelled-ultimo` detenido, `pendiente` pendiente, `skipped-tras-failure` detenido, `neutral-tras-cancelled` detenido, los dos empates detenido). — ALTA
- **CA-18** Mergeado desde GitHub: un PR con `mergedAt` y sin `land.merged` produce el nodo Mergeado con `ts=mergedAt` (cubre la merge queue); con ambos, manda el `ts` del evento y `mergedAt` queda en `refs`. — ALTA
- **CA-19** Vinculación por body: un PR que no figura en ningún `run.pr-opened` pero cuyo body dice `Closes #NN` o `Refs #NN` entra al trabajo del issue NN solo si ese trabajo tiene spec; si no, queda fuera (grill D8). — ALTA
- **CA-20** Publicado/taggeado (grill D22): por cada PR mergeado con `mergeCommit.oid`, `git -C <ruta> tag --contains <sha>` sin fetch; con uno o más tags el nodo Publicado dice `taggeado` y lista todos los tags (monorepo: uno por prefijo); si el SHA no existe en el clon, `subEstado: "sin dato local"`. `en prod` no se produce nunca en v1 (grill D21). Test con un repo temporal con tag. — ALTA
- **CA-21** Polling y refresh manual: GitHub se consulta por repo cada `--github-interval` (default `2m`); `POST /api/refresh` dispara una consulta inmediata y responde `202 {fetchedAt}` al terminar; dos pedidos concurrentes se fusionan en una sola consulta en vuelo. Test con reloj y `Runner` falsos. — ALTA
- **CA-22** Caché: el último snapshot de GitHub de cada repo se persiste en `<cache>/sdd/<id con / → _>.json` (`os.UserCacheDir()`, override `--cache-dir`); al arrancar se sirve ese snapshot con su `fetchedAt` mientras corre la primera consulta; test con `--cache-dir` temporal arranca sin `gh` y sirve lo cacheado. — ALTA
- **CA-23** Degradado (grill D29): sin `gh` en PATH, `gh auth status` fallido, repo sin remote, error de red o `rate limit`, la respuesta lleva `github:{ok:false, reason:<una de: sin-gh|sin-login|sin-remote|red|rate-limit>, fetchedAt:<última caché o null>}`, Idea queda con la caché o vacía, los PRs sin dato fresco llevan `subEstado:"sin dato"` y `stale:true`, lo local se sirve igual; el fallo de un repo no afecta a los demás; el proceso nunca termina por GitHub. — ALTA
- **CA-24** Presupuesto de rate limit: una consulta GraphQL por repo y por poll (issues abiertos, PRs abiertos y los 100 PRs más recientes por `updatedAt` con `isDraft`, `mergeable`, `reviewDecision`, `statusCheckRollup`, `reviewThreads(isResolved)`, `mergedAt`, `mergeCommit`, `body`); el test cuenta invocaciones del fake por poll. — ALTA
- **CA-25** `GET /api/stream` (SSE): emite `event: snapshot` con `{scannedAt, fetchedAt}` cada vez que cambia el modelo (escaneo, refresh local o de GitHub) y un `: ping` cada 30 s; test con `httptest` lee el primer evento tras un cambio en el fixture. — ALTA

**UI: board (capa 4)**
- **CA-26** Tokens y kit: `web/src/index.css` define bajo `@theme` todos los tokens de la tabla `### Color` de `DESIGN_SYSTEM.md` para claro y oscuro (`prefers-color-scheme` por default, `data-theme="light|dark"` manual), familias, radios y sombras; shadcn/ui queda instalado con `components.json` y los componentes listados en `## Diseño`; ningún hex fuera de `web/src/index.css`: `git grep -IhoE '#([0-9a-fA-F]{3}){1,2}' -- 'web/src/**' ':!web/src/index.css' | wc -l` da `0`. — ALTA (gate de grep + build)
- **CA-27** Navegación, Sidebar, Refresh y tema según `DESIGN_SYSTEM.md`: en ≥768 px una fila con nombre, `Select` de repo, Segmented activos·todos, `Refresh` y toggle de tema; en <768 px hamburguesa + «sdd · <repo>» + `Refresh`, y el Sidebar (Repo con conteos, Mostrar, Tema, pie con versión y carpeta) con foco atrapado, cierre por X/Esc/overlay/elección; el tema elegido persiste en `localStorage`; `Refresh` hace `POST /api/refresh`, gira mientras espera y muestra la antigüedad corta de `fetchedAt`. Lógica con Vitest + Testing Library (que esta capa suma con `npm test`); aspecto: humano. — BAJA (lógica, ALTA con el runner) · NULA (aspecto)
- **CA-28** Board desktop: siete carriles fijos con nombre, conteo y punto de color de etapa; una tira por trabajo ordenada por `ultimaActividad`; nodos con punto de etapa, título con elipsis, chip de referencia, antigüedad corta, badge de sub-estado y chip de PR; ramas una debajo de otra dentro de la tira con conectores; stacks como nodos encadenados `n/N` en PR y Mergeado; carril vacío con EmptyState corto; el scroll horizontal vive solo en el board; el filtro «todos» incluye los `terminado`. — BAJA (estructura) · NULA (aspecto)
- **CA-29** Filtros en URL: `?repo=<id>` y `?filtro=activos|todos` se leen al cargar y se escriben al cambiar (`history.replaceState`), así una recarga o un link compartido conservan el filtro; el Select de repo lista los repos del snapshot con su conteo de trabajos. — BAJA (ALTA con el runner)
- **CA-30** Celular (grill D33): por debajo de 768 px el board se reemplaza por la Lista de trabajos (filas de 44 px: punto de etapa, título con elipsis, badge, antigüedad; nada baja de línea); tocar una fila abre el detalle como Sheet; sin scroll horizontal de página; inputs sin zoom de iOS; safe-area respetada. — NULA (protocolo humano en el celular)
- **CA-31** Estados: skeleton con forma de board hasta el primer snapshot con `scanning:false`; Banner `warning` persistente «Sin GitHub: mostrando solo lo local» con «Reintentar» cuando `github.ok=false`, nodos `stale` en gris; Banner `error` cuando `/api/workspace` falla, conservando el último dato; EmptyState de página con `sdd --dir <otra carpeta>` cuando `repos=[]`; Toast de una línea al terminar un refresh manual; el front se suscribe a `/api/stream` y refetchea en cada `snapshot`, con reconexión. — BAJA (ALTA con el runner) · NULA (aspecto)
- **CA-32** Teclado y accesibilidad: Tab recorre todo lo accionable con anillo `:focus-visible`; en el board las flechas mueven el foco entre nodos y Enter abre el detalle; el Select es el custom de shadcn/ui, nunca `<select>` nativo; contraste AA por tokens; `aria-label` en hamburguesa, Refresh y toggle de tema. — NULA (protocolo humano) salvo `aria-label` y ausencia de `<select>` (ALTA por grep)

**UI: detalle (capa 5)**
- **CA-33** Ruta `/trabajos/<id>`: abre el detalle (Dialog ≥768 px, Sheet debajo) sobre el board, con título, chips de repo e issue, y el Grafo SVG: un nodo por etapa alcanzada con fecha corta y link (issue, grill y spec locales por ruta, PRs a GitHub), bifurcaciones como ramas (specs reemplazadas, runs fallados) y el nodo activo en `--color-primary-text`; Esc o cerrar vuelve a `/` conservando filtros; un id inexistente muestra EmptyState «trabajo no encontrado». — BAJA (ruteo, ALTA con el runner) · NULA (aspecto)
- **CA-34** Gantt SVG: una barra por etapa y por rama sobre un eje de fechas en mono con línea de hoy; ramas falladas o reemplazadas como barras paralelas etiquetadas; scroll horizontal dentro del contenedor; `ganttLayout(trabajo, ancho) → barras` es una función pura testeada (posiciones, barras abiertas hasta hoy, trabajos `sinHistoria` sin barras). — BAJA (layout, ALTA con el runner) · NULA (aspecto)

## Plan de entrega
| Capa | Etapa | CAs | Justificacion |
|---|---|---|---|
| 1 | parsers | CA-1 a CA-4 | Port puro de events-v1 y SDD-Tracking v1 con los vectores de conformance como tests; sin I/O, sin UI; `go test` verde solo y base de todo lo demás. |
| 2 | escaneo-api | CA-5 a CA-14 | Descubrimiento de repos, lectura de logs y artefactos, modelo de trabajos y API JSON con fixtures `testdata/`; se sostiene con `--no-github` y ya es útil desde `curl`. |
| 3 | github | CA-15 a CA-25 | Adapter `gh` con respuestas grabadas, gate de PR validado, Idea, merges por `mergedAt`, tags locales, polling, caché, degradado y SSE; sin red en tests. |
| 4 | ui-board | CA-26 a CA-32 | Tokens y kit, navegación, board, lista en celular, filtros y estados contra la API de las capas 2 y 3; suma Vitest + Testing Library. |
| 5 | ui-detalle | CA-33 y CA-34 | Ruta de detalle con Grafo y Gantt en SVG propio; depende del board y de la API de detalle. |

Cinco capas = un stack de cinco PRs en un run (`/sdd-run` con `gh stack`). No hay política de tamaño de PR activa; el corte es por cohesión y por el riesgo de revisar parsers, modelo, GitHub y UI en un solo diff.

## Fuera de alcance
- Acciones sobre los skills: aprobar specs, lanzar runs, mergear, deployar (grill, no-objetivos). El visor no escribe en clones ni en GitHub.
- Bandeja de pendientes, métricas y pipeline automatizado (grill D10): la arquitectura separa dominio (`internal/model`), fuentes (`internal/scan`, `internal/github`) y API para que entren después.
- `en prod` (grill D21): sin dato hasta `events-v2`, spec aparte en `skills`. Publicado solo muestra `taggeado`.
- Infra, `prama.yaml`, repos infra, environments y deployments de GitHub.
- UI o build en `skills`; cambios a `events-v1` o `SDD-Tracking v1`.
- PRs fuera de SDD (sin spec ni ancla) y auth en LAN.
- «Issues absorbidos» como rama del grafo (grill D7): no hay fuente de datos en `events-v1` ni en los markers; queda como gap (ver abajo), no se dibuja en v1.
- Workflow de release del binario (grill D27, Gaps del contrato).
- `fsnotify` y librerías de grafo/Gantt: SVG y polling propios.

## Inferencias
| # | Inferencia | Eleccion propuesta | Alternativa razonable | Confianza | Resolucion |
|---|---|---|---|---|---|
| 1 | Profundidad y exclusiones del escaneo | hasta 3 niveles bajo `--dir`, salteando `node_modules`, `.git`, `vendor`, `dist`, `build` | 2 niveles / flag `--depth` | alta | [ASSUMED] |
| 2 | Id de repo | `owner/repo` del remote `origin` (https, `git@`, `ssh://`); fallback nombre de carpeta | ruta absoluta | alta | [ASSUMED] |
| 3 | Id de trabajo (URL y API) | `<owner>/<repo>/issue/<NN>`, `/grill/<ref>` o `/spec/<ruta codificada>` según el ancla | hash corto del ancla | media | [ASSUMED] |
| 4 | Qué es «terminado» para el filtro activos | Publicado (taggeado), o Mergeado completo (todos los PRs del último run) con issue cerrado o sin issue | literal D16: solo `en prod` termina, así que hasta `events-v2` nada sale del board | media | [ASSUMED] |
| 5 | Observación de archivos en vivo | polling de mtime cada 2 s con stdlib (`--watch-interval`) | `fsnotify` (dependencia nueva) | alta | [ASSUMED] |
| 6 | Cómo llega el cambio al front | SSE `/api/stream` + refetch de `/api/workspace` | polling del front cada 15 s con `ETag` | media | [ASSUMED] |
| 7 | Consulta a GitHub | una query GraphQL por repo y por poll vía `gh api graphql` (issues abiertos, PRs abiertos y 100 recientes con checks, review, threads, merge) | `gh pr list --json` + `gh issue list --json` + GraphQL de threads por PR | media | [ASSUMED] |
| 8 | Respuestas grabadas de `gh` | interfaz `Runner` con fake que sirve `testdata/gh/<repo>/*.json` y falla ante invocaciones no grabadas | binario `gh` falso en PATH | alta | [ASSUMED] |
| 9 | Caché de snapshots | `os.UserCacheDir()/sdd/<id>.json`, override `--cache-dir` | `<dir>/.sdd-viewer/` dentro del workspace | alta | [ASSUMED] |
| 10 | Intervalos | flags `--github-interval` (2m) y `--watch-interval` (2s), más `--no-github` para apagar GitHub (tests y offline) | solo defaults sin flags | alta | [ASSUMED] |
| 11 | Título del trabajo | título del issue > H1 de la spec > H1 del grill > referencia del ancla | solo referencia | alta | [ASSUMED] |
| 12 | Alcance del port de SDD-Tracking | parser completo (canónico + legacy explícito e inferido + diagnósticos) y `Serialize` solo canónico; sin `Upsert` (el visor es solo lectura) | port completo incluido upsert | media | [ASSUMED] |
| 13 | Paquetes Go | `internal/events`, `internal/tracking`, `internal/scan`, `internal/model`, `internal/github`, `internal/api`, `cmd/sdd` solo wiring; archivos ≤300 líneas | un paquete `internal/sdd` | alta | [ASSUMED] |
| 14 | QR en terminal (D34) | dependencia `github.com/skip2/go-qrcode`, justificada en el PR | `qrencode` del sistema si está en PATH, si no solo la URL | media | [ASSUMED] |
| 15 | Detalle en desktop | Dialog modal con ruta `/trabajos/<id>` | panel lateral persistente | alta | [ASSUMED] |
| 16 | Runner de tests del front | Vitest + `@testing-library/react` + jsdom, script `npm test`, agregados en la capa 4; luego `/sdd-init --update` | sin runner (UI solo humana) | alta | [ASSUMED] |
| 17 | Kit shadcn/ui | `npx shadcn init` con `components.json` y añadir `button select badge dialog sheet tooltip skeleton sonner` | copiar solo los componentes a mano | media | [ASSUMED] |
| 18 | Ruteo del front | router propio mínimo (`URL` + `popstate`, dos rutas y query) | `react-router` | media | [ASSUMED] |
| 19 | Merge sin evento | nodo Mergeado con `ts=mergedAt`; con evento, manda el evento | ignorar `mergedAt` | alta | [ASSUMED] |
| 20 | Taggeado | `git tag --contains <mergeCommit>` sin fetch; varios tags se listan todos | solo el primer tag | alta | [ASSUMED] |
| 21 | Fixtures git | los tests arman repos temporales con `git init`/`git worktree add` (git está en local y en `ubuntu-latest`); logs y markers reales copiados a `testdata/` | fixtures con `.git` empaquetados en tar | alta | [ASSUMED] |
| 22 | Fechas | `ts` en UTC; antigüedad corta relativa; Gantt y tooltips en hora local del navegador | todo en UTC | alta | [ASSUMED] |
| 23 | Trabajos anteriores al log | `ts` = mtime del archivo, `sinHistoria: true`, sin barras en el Gantt | sin `ts` | media | [ASSUMED] |
| 24 | Carga progresiva | `scanning:true` con los repos ya leídos; SSE avisa cada repo terminado | esperar el escaneo completo | alta | [ASSUMED] |
| 25 | Nodos `stale` | `stale:true` en nodos de PR/Idea cuando `github.ok=false`, aunque haya caché | `stale` solo si la caché tiene más de 1 h | alta | [ASSUMED] |
| 26 | PRs mergeados a consultar | los 100 más recientes por `updatedAt` por repo; más viejos solo si un evento los nombra | todos los PRs | media | [ASSUMED] |

## Verificabilidad
Mixta, anclada en `.sdd/project.md`:
- **CA-1 a CA-25: ALTA.** Son tests Go deterministas que el contrato ya corre (`make test`, verificado 2026-10-25) sobre fixtures en `testdata/` y repos temporales con `git init`; nada toca red, `/home/chiche/workspace` ni GitHub (Límites y Decisiones humanas). CA-9, CA-12 y CA-13 suman el escalón 4 de la escalera (app levantada contra fixtures, `curl` sobre `/api/*`).
- **CA-26: ALTA.** Gate de grep sobre tokens más `npm run build` y `npx tsc -b`, todos verificados en el contrato.
- **CA-27, CA-28, CA-29, CA-31, CA-33, CA-34 (lógica): BAJA hoy.** El contrato dice «el front no tiene runner de tests» (Gaps). Serían ALTA con Vitest: la capa 4 lo suma y el run reporta esos tests como evidencia adicional, pero el grado se anota con lo que el contrato corre hoy; después del merge, `/sdd-init --update` registra el runner. Las señales que sí corren hoy: `npx tsc -b`, `npm run lint`, `npm run build`, `npx oxlint -D typescript/no-explicit-any`.
- **CA-27, CA-28, CA-30, CA-31, CA-32, CA-33, CA-34 (aspecto, celular, teclado): NULA.** «La UI renderizada… hace falta un navegador. El humano mira la app con `make run`» (Verificacion autonoma del contrato); no hay Playwright ni MCP de navegador. Protocolo humano en el Plan de verificacion.
- **Políticas de generación**: `commits` (gate regex), `max-lineas-archivo` 300 (el port de `index.ts` de 969 líneas y el board deben partirse en archivos cohesivos; inferencia 13), `no-any-ts`, `gofmt`, `no-panic` (los paquetes nuevos devuelven error; `web/embed.go` sigue siendo la excepción) aplican como gates duros a cada capa. Tamaño de PR, coverage y dependencias no están activas: las dependencias nuevas (`go-qrcode`, radix vía shadcn, Vitest, Testing Library) se justifican en el PR por la guía Go y el reviewer las juzga.

## Plan de verificacion
Mecanismo elegido: tests Go con fixtures (`go test -count=1 ./...`) para todo lo de servidor; grep y build para tokens; Vitest + Testing Library (desde la capa 4) para la lógica del front; protocolo humano para aspecto, celular y teclado. Alternativas: (más exhaustiva) sumar Playwright al repo y capturas contra la previsualización del design system, con `/sdd-init --update` para registrarlo; (más barata) sin Vitest, dejando toda la UI como prueba humana.

Por CA:
- CA-1 a CA-4: `go test -count=1 ./internal/events ./internal/tracking` con los vectores en `testdata/vectors/` y `testdata/tracking/expected.json`; `TestParseLine_Vectores`, `TestParseLine_Invalidas` (tabla), `TestParse_Fixtures`, `TestRef_RoundTrip`.
- CA-5 a CA-8, CA-11: `go test -count=1 ./internal/scan ./internal/model` con `testdata/mkworkspace.sh` (crea en `t.TempDir()` los repos con `git init`, copia `testdata/events/*.jsonl` y `testdata/tracking/*.md`, agrega un worktree, una carpeta sin git y una a profundidad 4); aserciones sobre conteos y sub-estados listados en CA-7 y CA-8.
- CA-9, CA-10, CA-13, CA-25: `go test -count=1 ./internal/api` con `httptest`; además app levantada: `go build -o bin/sdd ./cmd/sdd && W=$(bash testdata/mkworkspace.sh) && ./bin/sdd --dir "$W" --no-github --cache-dir "$(mktemp -d)" &` y `curl -sf localhost:4321/api/workspace | jq '.trabajos|length'` = 12, `curl -s -o /dev/null -w '%{http_code}' localhost:4321/api/trabajos/no-existe` = 404; matar el proceso.
- CA-12: test de integración que escribe una línea válida en el log del fixture y espera (≤10 s) el nodo nuevo; también a mano con la app levantada y `curl`.
- CA-14: `go test -count=1 ./cmd/sdd -run TestBanner` con interfaces inyectadas; a mano `./bin/sdd --host 0.0.0.0 --dir "$W"` muestra IP y QR.
- CA-15 a CA-24: `go test -count=1 ./internal/github ./internal/model` con `testdata/gh/` (respuestas grabadas, incluidos los 7 rollups del gate copiados a `testdata/gh/rollups/`), fake `Runner` que cuenta invocaciones y rechaza mutaciones, reloj falso para el polling, `--cache-dir` temporal; repo temporal con tag para CA-20.
- CA-26: `cd web && npm run build && git grep -IhoE '#([0-9a-fA-F]{3}){1,2}' -- 'web/src/**' ':!web/src/index.css' | wc -l` = 0; `grep -rn '<select' web/src` vacío (también CA-32).
- CA-27, CA-29, CA-31, CA-33, CA-34 (lógica): `cd web && npm test` (Vitest) con tests por rol y texto: Sidebar abre/cierra y atrapa foco, filtro escribe la URL, Refresh llama `POST /api/refresh`, banner aparece con `github.ok=false`, `ganttLayout` posiciona barras; más `npx tsc -b`, `npm run lint`, `npx oxlint -D typescript/no-explicit-any`.
- Gates de políticas en cada capa: `git log --format=%s origin/main..HEAD | grep -vE '^(feat|fix|chore|docs|test|refactor|ci|build|perf|style)(\([a-z0-9/_.-]+\))?!?: .+$'` vacío; `git diff --name-only --diff-filter=AM origin/main...HEAD -- '*.go' '*.ts' '*.tsx' | xargs -r wc -l | awk '$2!="total" && $1>300'` vacío; `gofmt -l .` vacío; `git grep -n 'panic(' -- '*.go' ':!cmd/' ':!*_test.go' ':!web/embed.go'` vacío.

Protocolo de prueba humana (CA-27, CA-28, CA-30, CA-31, CA-32, CA-33, CA-34):
1. `make run` en `/home/chiche/workspace/sdd` (escanea `/home/chiche/workspace`) y abrir `http://127.0.0.1:4321` en desktop. Confirmar: siete carriles, tiras de `waica#203` (dos ramas de Run, 5 PRs) y `platform#54` (Mergeado 3/4), nodos de una línea, sin scroll horizontal de página.
2. Cambiar el tema con el toggle; recargar: se conserva. Elegir repo `waica` y filtro «todos»: la URL muestra `?repo=chichex/waica&filtro=todos` y la recarga la respeta.
3. Tab por la cabecera y el board; flechas entre nodos; Enter abre el detalle; Esc lo cierra y vuelve al board con los filtros.
4. En el detalle de `waica#203`: grafo con la bifurcación del run interrumpido y el Gantt con barras paralelas; links a los PRs abren GitHub.
5. Cortar la red (o `gh auth logout` en una shell aparte) y tocar Refresh: aparece el banner warning, los PRs quedan `sin dato` en gris, el board no se vacía. Restaurar y reintentar: el banner desaparece.
6. `./bin/sdd --dir ~/workspace --host 0.0.0.0`: la terminal muestra la URL con IP LAN y el QR; abrir en el celular: lista de trabajos de filas de una línea, hamburguesa que abre el Sidebar, tocar una fila abre el Sheet con grafo y Gantt con scroll horizontal, sin zoom al tocar un input, sin apilado.
7. Agregar una línea a un `events.jsonl` real (por ejemplo correr un `/grill` y pausarlo): el board la refleja en menos de 2 s sin recargar.

## Riesgos y gaps
- **Revisar primero: inferencia 4** (qué cuenta como «terminado»). Define qué muestra el board por default; la lectura literal de D16 dejaría todo visible hasta `events-v2`.
- **Revisar: inferencia 7** (una query GraphQL por repo) y **26** (100 PRs recientes). Fijan el adapter de GitHub y el presupuesto de rate limit (12 repos × 30 polls/h ≈ 360 consultas/h de 5000).
- **Desviación del handoff**: D7 pide dibujar «issues absorbidos» como rama; ni `events-v1` ni `SDD-Tracking v1` registran qué issue absorbió a cuál. v1 no los dibuja. Si se quiere, es un cambio en `skills` (issue-triage o `events-v2`), no acá.
- `events.jsonl` es por máquina: un run hecho en otra máquina no aparece (riesgo aceptado en el grill).
- Tags locales no pusheados pueden marcar `taggeado` antes de tiempo (aceptado en el grill); y sin `git fetch` el merge commit puede no existir en el clon → `sin dato local`.
- **Gap del contrato**: front sin runner de tests. La capa 4 suma Vitest y deja los CA de UI en BAJA hasta `/sdd-init --update`. Sin Playwright ni navegador: aspecto, celular y teclado son prueba humana.
- **Política `max-lineas-archivo` 300**: el port de `index.ts` (969 líneas) y de `sdd-events.ts` (416) exige partir por responsabilidad (preámbulo, legacy, referencias, serialización); el board y el Gantt también.
- **Dependencias nuevas** sin gate activo: `go-qrcode` (Go), radix vía shadcn/ui, Vitest y Testing Library (front). Se justifican en los PRs; la guía Go lo pide.
- `testdata/` necesita `git` en el entorno de tests (local y `ubuntu-latest` lo tienen); si falta, los tests de `internal/scan` se saltean con `t.Skip` y lo dicen.
- Run con stack de 5 PRs: exige la extensión `gh stack` (instalada localmente, v0.2.1); su conducta en vivo está documentada en el contrato de `skills` (Gaps).
- `web/embed.go` conserva la excepción de `no-panic`; `Dist()` con error es un cambio aparte.
- Vectores y fixtures copiados desde `skills`: si los docs normativos cambian, los tests de este repo no lo detectan solos; un test lee el bloque de vectores copiado, no el original.

## Desviaciones
- 2026-10-27 [DEVIATION] Inferencia 13: se suma `internal/workspace` para la orquestación en vivo (escaneo progresivo, polling de mtime, avisos a suscriptores); `internal/api` queda solo con HTTP. No cambia el alcance.
- 2026-10-27 [DEVIATION] Plan de verificacion: `testdata/mkworkspace.sh` no copia `testdata/tracking/*.md` a los repos. Esos fixtures nombran `owner/repo#9` y sumarían trabajos ajenos, rompiendo el conteo de 12 de CA-9. Copia en cambio artefactos sintéticos de `testdata/workspace/` que se suman a trabajos existentes (spec de `#203` en el árbol y en el common-dir, grill de platform en el common-dir). Los 21 fixtures de tracking se verifican en CA-3. No cambia el alcance.
- 2026-10-27 [DEVIATION] CA-8: cuando el log nombra un grill o una spec en eventos de run sin registrar su transición (por ejemplo `waica#203`, que no tiene `grill.*` ni `spec.*`), el trabajo muestra el nodo igual: Grill `finalizado` (una spec sale de un grill finalizado) y Spec `aprobada` (el run la aceptó), sin barra de Gantt porque no hay evento. No cambia el alcance.

- 2026-10-27 [DEVIATION] CA-17: el paréntesis dice «`brik-117` detenido», pero el filtro jq canónico de sdd-land da `verde` para ese rollup (un `CANCELLED` reemplazado por un `SUCCESS` posterior; `jq` sobre `brik-117.json` → `{"estado":"verde","detenidos":[],"supersedidos":1}`, igual que la tabla `JQ_CASES` de `sdd-land-gate.test.ts`). Manda la cláusula principal del CA («dan el mismo `estado` que ese gate»): el test espera `verde`. No cambia el alcance.
- 2026-10-27 [DEVIATION] CA-23: `sin-login` se detecta por la salida de error de `gh` (`gh auth login`, `Bad credentials`, HTTP 401) y no con un `gh auth status` aparte, para sostener la consulta única por repo y por poll de CA-24. Un repo cuyo remote no resuelve en GitHub (`Could not resolve to a Repository`) cuenta como `sin-remote`. `--no-github` reporta `reason: "desactivado"`, que no es una falla. No cambia el alcance.
- 2026-10-27 [DEVIATION] CA-20: un nodo Publicado `sin dato local` no adelanta la etapa del trabajo (no prueba que se haya publicado); `taggeado` sí. Un PR mergeado o cerrado sin merge muestra `mergeado` o `cerrado` como sub-estado de su nodo PR. Las respuestas grabadas de `testdata/gh/` son sintéticas con la forma real de `gh api graphql` (el repo es público y no se copia contenido de repos privados); los 7 rollups del gate sí son copia literal. No cambia el alcance.

- 2026-10-27 [DEVIATION] Inferencia 17: `components.json` se escribió a mano (estilo `new-york`, Radix y Lucide) y se corrió `npx shadcn@4.22.1 add button select badge dialog sheet tooltip skeleton sonner`, en vez de `npx shadcn init`. El `init` 4.22 pide un preset interactivo y su default (Nova) trae la fuente Geist, que el design system prohíbe (sin fuentes externas). El kit usa `cn` (helper oficial de shadcn, repo `shadcn-ui/cn`) y `radix-ui`; se quitó `next-themes`: el Toaster recibe el tema del visor. No cambia el alcance.
- 2026-10-27 [DEVIATION] CA-26: además de los tokens de la tabla, `index.css` define `--color-overlay` con el valor `rgb(0 0 0 / 0.5)` de la sección Dialog / Sheet del doc, y una capa `@theme inline` que mapea los nombres de shadcn (`background`, `popover`, `muted`…) a los tokens de la tabla, como pide «con sus tokens mapeados a los de arriba». No son colores nuevos. No cambia el alcance.
- 2026-10-27 [DEVIATION] CA-31: el skeleton dura hasta el primer snapshot con `scanning:false`, como dice el CA. El escaneo progresivo existe en la API (`scanning:true` con los repos ya leídos), pero la UI no dibuja el parcial. En desktop el board se muestra con los dos filtros (CA-28); la Lista de trabajos es solo de celular. `npm test` no se sumó al workflow de CI: lo registra `/sdd-init --update`. No cambia el alcance.

- 2026-10-27 [DEVIATION] CA-34: `ganttLayout(trabajo, ancho, hoy)` recibe `hoy` explícito para seguir siendo pura (las barras abiertas llegan hasta ese instante). CA-33: los links locales de grill y spec se listan como rutas copiables bajo el grafo, porque el navegador no abre `file://` desde una página http; los nodos con URL (issue, PRs) son links del SVG. No cambia el alcance.

## Resultado de ejecucion (2026-10-27 · HEAD por capa)
Cada capa corrió su verificación final sobre su propio HEAD: capa 1/5 `adf75fd` · capa 2/5 `26cbdba` · capa 3/5 `b50c1d9` · capa 4/5 `d9f9971` · capa 5/5 `2212c48`.

| CA | Capa | Estado | Evidencia |
|---|---|---|---|
| CA-1 | 1/5 | verificado | `go test -count=1 ./internal/events`: `TestParseLine_Vectores` (12 vectores leídos del bloque copiado, 11 eventos del catálogo, round trip byte a byte) verde |
| CA-2 | 1/5 | verificado | `go test -count=1 ./internal/events`: `TestParseLine_Invalidas` (28 casos, los 16 códigos) y `TestReadLog_SaltaInvalidasYSegmentoFinal` verdes |
| CA-3 | 1/5 | verificado | `go test -count=1 ./internal/tracking`: `TestParse_Fixtures` (21/21 contra `testdata/tracking/expected.json`), `TestParse_VectoresRoundTrip`, `TestParse_Preambulo` (BOM, fences, conflictos), `TestParse_AliasesLegacy`, `TestSerialize_SoloCanonica` verdes |
| CA-4 | 1/5 | verificado | `go test -count=1 ./internal/tracking`: `TestRef_RoundTrip` y `TestDecodeRef_Invalidas` (10 formas inválidas) verdes |
| POL-commits | 1/5 | cumplida | gate de `git log --format=%s origin/main..HEAD` vacío |
| POL-max-lineas-archivo | 1/5 | cumplida | gate de `wc -l` vacío (máximo 257, `internal/tracking/legacy.go`) |
| POL-no-any-ts | 1/5 | cumplida | `npx oxlint -D typescript/no-explicit-any`: exit 0 |
| POL-gofmt | 1/5 | cumplida | `gofmt -l .` vacío |
| POL-no-panic | 1/5 | cumplida | `git grep -n 'panic(' …` vacío |
| CA-5 | 2/5 | verificado | `go test -count=1 ./internal/scan`: `TestDiscover_Workspace` (workspace de `testdata/mkworkspace.sh`: 6 repos, worktree unido, sin profundidad 4 ni `node_modules`) y `TestRepoID_FormasDeRemote` verdes |
| CA-6 | 2/5 | verificado | `go test -count=1 ./internal/scan`: `TestRead_FuentesLocales` (2 inválidas contadas, final sin LF descartado, gana el árbol, `rota.md` en errores) y `TestRead_DirectoriosAusentesSonVacios` verdes |
| CA-7 | 2/5 | verificado | `go test -count=1 ./internal/model`: `TestBuild_TrabajosDeLogsReales` (3 + 2 + 6 + 1 = 12, PR 249 fuera, 5 PRs en `#203` y `#205`) verde |
| CA-8 | 2/5 | verificado | `go test -count=1 ./internal/model`: `TestBuild_EtapasYRamas` (`#203` con rama `fallado (interrupted)` y Mergeado 5/5; `#54` Mergeado 3/4 con PR 98 `sin dato`) y `TestBuild_RunNuevoTrasFallo` verdes |
| CA-9 | 2/5 | verificado | `go test -count=1 ./internal/api`: `TestWorkspace_FixtureCompleto`, `TestWorkspace_FiltroActivos`, `TestWorkspace_ParcialMientrasEscanea` verdes; app levantada (`./bin/sdd --dir "$(bash testdata/mkworkspace.sh)"`): `curl …/api/workspace \| jq '.trabajos\|length'` = 12, `scanning` false |
| CA-10 | 2/5 | verificado | `go test -count=1 ./internal/api`: `TestTrabajo_DetalleYNoEncontrado` y `go test ./internal/model` `TestBuild_Gantt` verdes; app levantada: `/api/trabajos/chichex/waica/issue/203` → gantt 4 barras, 5 PRs; `/api/trabajos/no-existe` → 404 |
| CA-11 | 2/5 | verificado | `go test -count=1 ./internal/model`: `TestBuild_SinHistoria` (un nodo, ts = mtime, sin barras) verde |
| CA-12 | 2/5 | verificado | `go test -count=1 ./internal/api`: `TestWorkspace_RefreshLocalEnVivo` (intervalo default 2 s, nodo nuevo antes de 10 s, `scannedAt` avanza) verde; app levantada: línea agregada al log → `chichex/waica/issue/300` visible por `curl` |
| CA-13 | 2/5 | verificado | `go test -count=1 ./internal/api`: `TestWorkspace_Vacio` (`repos: []`, `trabajos: []`, `scanning: false`, `/api/health` 200) verde |
| CA-14 | 2/5 | verificado | `go test -count=1 ./cmd/sdd -run TestBanner`: `TestBanner_LAN` (URL con la primera IPv4 no loopback, QR ≤ 33 módulos, aviso) y `TestBanner_Local` verdes; a mano `./bin/sdd --host 0.0.0.0` imprimió `abrí http://192.168.1.92:4392` y un QR de 33 de ancho |
| POL-commits | 2/5 | cumplida | gate de `git log` vacío |
| POL-max-lineas-archivo | 2/5 | cumplida | gate de `wc -l` vacío tras partir `internal/api/api_test.go` (314 → 160 + 118 + 60, mismas 24 aserciones) |
| POL-no-any-ts | 2/5 | cumplida | `npx oxlint -D typescript/no-explicit-any`: exit 0 |
| POL-gofmt | 2/5 | cumplida | `gofmt -l .` vacío |
| POL-no-panic | 2/5 | cumplida | `git grep -n 'panic(' …` vacío |
| CA-15 | 3/5 | verificado | `go test -count=1 ./internal/github`: `TestReadOnly_RechazaEscrituras` (mutation, `-X POST/DELETE/PATCH`, `pr merge/edit`, `issue comment/create/close`), `TestFetch_NoEjecutaEscrituras` y `TestFetch_RespuestaGrabada` verdes; el doble `internal/github/fake` falla el test ante una invocación no grabada o de escritura |
| CA-16 | 3/5 | verificado | `go test -count=1 ./internal/model`: `TestBuild_IdeaDesdeIssuesAbiertos` (14 trabajos: ideas `waica#210` y `platform#120`; `#57` toma el título del issue sin duplicarse; issue cerrado sin artefactos no aparece) verde |
| CA-17 | 3/5 | verificado | `go test -count=1 ./internal/github`: `TestChecks_RollupsDelGate` (los 7 rollups dan el `estado`, `detenidos` y `supersedidos` del filtro jq; ver desviación de `brik-117`) y `TestSubEstado_GateDeSddLand`; `go test ./internal/model` `TestBuild_SubEstadosDePR` verdes |
| CA-18 | 3/5 | verificado | `go test -count=1 ./internal/model`: `TestBuild_MergeadoDesdeGitHub` (PR 105 sin `land.merged` → Mergeado con ts = `mergedAt`; `#54` con evento conserva su ts y `mergedAt` en refs) verde |
| CA-19 | 3/5 | verificado | `go test -count=1 ./internal/model`: `TestBuild_VinculacionPorBody` (`Closes #57` suma el PR 130; `Refs #120` y `Refs #210` quedan fuera por no tener spec) verde |
| CA-20 | 3/5 | verificado | `go test -count=1 ./internal/scan`: `TestTagsContaining_RepoTemporal` (repo temporal con `api/v1.0.57` y `web/v2.1.0`, SHA ausente = sin dato local) y `go test ./internal/model` `TestBuild_Taggeado` (taggeado con todos los tags, `sin dato local`, nunca `en prod`) verdes |
| CA-21 | 3/5 | verificado | `go test -count=1 ./internal/github`: `TestPoller_IntervaloYUnaConsultaPorRepo` (reloj falso, intervalo 2m), `TestPoller_RefreshConcurrenteSeFusiona` y `TestPoller_RefreshMientrasSeAplicaConsultaDeNuevo`; `go test ./internal/api` `TestRefresh_ConsultaYResponde202` verdes; app levantada: `POST /api/refresh` → 202 `{"fetchedAt":null}` con `--no-github`. Fix `b50c1d9`: un refresh que llegaba mientras se aplicaban los resultados de una consulta ya terminada se fusionaba con ella (intermitente en `TestRefresh_ConsultaYResponde202` bajo carga); loop rojo `go test -count=1 ./internal/github -run TestPoller_RefreshMientrasSeAplica` → «consultas = 1 … want 2»; tras el fix, suite completa 15 corridas sin fallas y `-race` limpio |
| CA-22 | 3/5 | verificado | `go test -count=1 ./internal/github` `TestCache_RutaYRoundTrip` y `go test ./internal/api` `TestCache_ArrancaSinGhYSirveLoCacheado` (sirve la caché con su `fetchedAt` mientras corre la primera consulta y después sin `gh`) verdes |
| CA-23 | 3/5 | verificado | `go test -count=1 ./internal/github` `TestReason_Degradado` y `TestFetch_ErrorDeGhSeClasifica`; `go test ./internal/api` `TestDegradado_UnRepoFallaYLosDemasNo` y `TestDegradado_SinGhSirveLoLocal` verdes; app levantada con `--no-github`: PR 98 `sin dato` y `stale: true`, 12 trabajos locales |
| CA-24 | 3/5 | verificado | `go test -count=1 ./internal/github`: `TestPoller_IntervaloYUnaConsultaPorRepo` (una invocación `gh api graphql` por repo y por poll) y `TestPoller_FallaDeUnRepoNoAfectaAOtros` verdes |
| CA-25 | 3/5 | verificado | `go test -count=1 ./internal/api`: `TestStream_EmiteSnapshotAlCambiar` y `TestStream_PingPeriodico` (rojo comprobado por mutación: sin el ping, el test falla por timeout) verdes; app levantada: `curl -N /api/stream` recibió `event: snapshot` con `scannedAt` y `fetchedAt` tras agregar una línea al log |
| POL-commits | 3/5 | cumplida | gate de `git log` vacío |
| POL-max-lineas-archivo | 3/5 | cumplida | gate de `wc -l` vacío |
| POL-no-any-ts | 3/5 | cumplida | `npx oxlint -D typescript/no-explicit-any`: exit 0 |
| POL-gofmt | 3/5 | cumplida | `gofmt -l .` vacío |
| POL-no-panic | 3/5 | cumplida | `git grep -n 'panic(' …` vacío |
| CA-26 | 4/5 | verificado | `cd web && npm test`: `src/tokens.test.ts` (lee la tabla `### Color` del design system: 29 tokens con su valor claro bajo `@theme` y el oscuro bajo `prefers-color-scheme` y `data-theme="dark"`; familias, radios y sombras) verde; `npm run build` ok; `git grep -IhoE '#([0-9a-fA-F]{3}){1,2}' -- 'web/src/**' ':!web/src/index.css' \| wc -l` = 0; `components.json` y los 8 componentes del kit en `web/src/components/ui/` |
| CA-27 | 4/5 | verificado (lógica) · pendiente humano (aspecto) | `cd web && npm test`: `Sidebar.test.tsx` (abre con Repo, Mostrar, Tema y pie; Esc cierra y devuelve el foco a la hamburguesa; elegir repo cierra y escribe `?repo`), `App.test.tsx` (Refresh hace `POST /api/refresh`, muestra «…» con `aria-busy`, toast al terminar; tema persiste en `localStorage`) verdes. Aspecto: protocolo humano, pasos 1, 2 y 6 |
| CA-28 | 4/5 | verificado (estructura) · pendiente humano (aspecto) | `cd web && npm test`: `Board.test.tsx` (siete carriles con nombre y conteo, «Sin trabajos en Publicado», una tira por trabajo en orden de `ultimaActividad`, ramas, stacks `n/N`, nodo `stale`) verde. Aspecto: protocolo humano, paso 1 |
| CA-29 | 4/5 | verificado | `cd web && npm test`: `filtros.test.tsx` (lee `?repo` y `?filtro` al cargar; cambiar el filtro escribe la URL con `replaceState` sin sumar historia; el Select lista los repos con su conteo y escribe `?repo`) verde |
| CA-30 | 4/5 | pendiente humano | NULA: protocolo humano paso 6 (celular por LAN: lista de una línea, Sheet, sin zoom, safe-area). Implementado: `ListaTrabajos` (<768 px), cabecera de celular, `viewport-fit=cover`, inputs ≥16 px en `pointer: coarse` |
| CA-31 | 4/5 | verificado (lógica) · pendiente humano (aspecto) | `cd web && npm test`: `App.test.tsx` (skeleton hasta `scanning:false`, Banner warning con Reintentar → `POST /api/refresh`, Banner error que conserva el último dato, EmptyState con `sdd --dir <otra carpeta>`, refetch por `snapshot` de `/api/stream` y reconexión tras un corte) verde. Aspecto: protocolo humano, paso 5 |
| CA-32 | 4/5 | verificado (aria-label, sin select nativo, flechas) · pendiente humano (Tab, foco, contraste) | `grep -rn '<select' web/src` vacío; `aria-label` «Menú», «Actualizar» y radiogroup «Tema» en el fuente; `App.test.tsx` lo confirma por rol; `Board.test.tsx` (flechas mueven el foco, Enter abre el detalle) verde. Resto: protocolo humano, paso 3 |
| POL-commits | 4/5 | cumplida | gate de `git log` vacío |
| POL-max-lineas-archivo | 4/5 | cumplida | gate de `wc -l` vacío |
| POL-no-any-ts | 4/5 | cumplida | `npx oxlint -D typescript/no-explicit-any`: exit 0 |
| POL-gofmt | 4/5 | cumplida | `gofmt -l .` vacío |
| POL-no-panic | 4/5 | cumplida | `git grep -n 'panic(' …` vacío |
| CA-33 | 5/5 | verificado (ruteo) · pendiente humano (aspecto) | `cd web && npm test`: `Detalle.test.tsx` (`/trabajos/<id>` abre el Dialog sobre el board con título, chips de repo e issue con link, grafo de 9 nodos con la rama `run-1` y un solo nodo activo, links de PRs y rutas locales; Esc vuelve a `/` conservando `?repo&filtro`; id inexistente → «Trabajo no encontrado»; en celular abre como Sheet; tocar una fila de la lista abre el detalle) verde; app levantada: `GET /trabajos/chichex%2Fwaica%2Fissue%2F203` → 200 (SPA) y `/api/trabajos/chichex%2Fwaica%2Fissue%2F203` → gantt 4, PRs 5. Aspecto: protocolo humano, pasos 3, 4 y 6 |
| CA-34 | 5/5 | verificado (layout) · pendiente humano (aspecto) | `cd web && npm test`: `src/components/detalle/gantt.test.ts` (`ganttLayout`: una fila por barra con las ramas cerradas etiquetadas, posiciones por fecha, barra abierta hasta hoy, sin historia sin barras) y `Detalle.test.tsx` (4 barras y la línea de hoy en el SVG) verdes. Aspecto: protocolo humano, paso 4 |
| POL-commits | 5/5 | cumplida | gate de `git log` vacío |
| POL-max-lineas-archivo | 5/5 | cumplida | gate de `wc -l` vacío |
| POL-no-any-ts | 5/5 | cumplida | `npx oxlint -D typescript/no-explicit-any`: exit 0 |
| POL-gofmt | 5/5 | cumplida | `gofmt -l .` vacío |
| POL-no-panic | 5/5 | cumplida | `git grep -n 'panic(' …` vacío |
| POL-estilo-go | — | guía aplicada | sin gate: la juzga el reviewer |
| POL-coding-policies | — | guía aplicada | sin gate: la juzga el reviewer |
