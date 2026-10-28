# Spec — Visor: anclas, traza y filtros
<!-- Generada por /sdd-spec el 2026-10-28. Fuente: pedido libre. Estado: implemented -->
<!-- SDD-Tracking: version=1; type=spec; state=implemented; issue=none; grill=none; superseded-by=none -->

## Contexto
El stack `sdd/visor-flujo-sdd` (PRs #2 a #6, sin mergear) implementa la spec `visor-flujo-sdd`. Probado contra `/home/chiche/workspace` (2026-10-28), el board no se entiende: de 138 trabajos «activos», 115 no tienen traza en `events.jsonl` (los logs arrancan el 15 a 17 de octubre), los grills sueltos, las specs sueltas y los issues se intercalan ordenados solo por última actividad, y «activos» casi iguala a «todos» (138 vs 153) porque `terminado` exige tag o merge completo con issue cerrado (`internal/model/build.go:68`). Hoy el modelo ya calcula `SinHistoria` por trabajo (`internal/model/build.go:61`) y la API lo serializa (`internal/model/trabajo.go:72`), pero el front solo lo parsea (`web/src/lib/api.ts:40`) y no lo dibuja. El único filtro del server es `filtro=activos|todos` más `repo` (`internal/api/api.go:66-79`); el front guarda `repo` y `filtro` en la URL (`web/src/lib/router.ts`) y los muestra en un `Segmented` (`web/src/components/opciones.ts`) y en el Sidebar. Esta spec corrige los tres problemas sobre el código del stack: su base es la rama `sdd/visor-flujo-sdd/5-ui-detalle`, no `main`, y el PR resultante se apila sobre #6.

## Diseño
Superficie `web` (contrato `## Diseño`), con sistema: tokens bajo `@theme` en `web/src/index.css`; kit shadcn/ui en `web/src/components/ui/` (`badge`, `button`, `dialog`, `select`, `sheet`, `skeleton`, `sonner`, `tooltip`); piezas propias en `web/src/components/` (`marcas.tsx`: `EtapaDot`, `SubEstadoBadge`, `ChipRef`; `Segmented.tsx`; `RepoSelect.tsx`; `ListaTrabajos.tsx`; `Sidebar.tsx`; `Navegacion.tsx`; `board/Board.tsx`, `board/Tira.tsx`, `board/NodoCard.tsx`; `detalle/Detalle.tsx`); doc `web/DESIGN_SYSTEM.md` (dirección `sobrio-indigo`). Sin capturas de referencia: el diseño se describe acá y lo valida el protocolo humano.

Pantallas y estados:
- **Board desktop (≥768 px)**: cabecera actual de una fila (`Navegacion.tsx`) + **Barra de filtros** nueva de 40 px debajo (fondo `--color-bg`, borde inferior `--color-border`, gutter `6`): `Input` de búsqueda (shadcn `input`, 32 px, ancho 280 px, ícono `search` de 16 px y botón de limpiar `x`) y siete **Chips de etapa** (botón de 28 px, radio `md`, un solo borde `--color-border`, texto `xs` 500 con el nombre y el conteo en mono; activa con `--color-primary-subtle` y `--color-primary-text`, `aria-pressed`). El board se parte en **grupos por repo**: un **Encabezado de repo** por grupo que cruza las siete columnas (`owner/repo` en mono `xs` 600 mayúsculas `--color-fg-secondary`, conteo a la derecha, fondo `--color-bg-subtle`, radio `sm`, altura 28 px), y debajo sus tiras. Con el filtro «todos», al final un **Grupo colapsado** «Sin traza (N)»: `Button ghost` de 32 px con chevron (`chevron-right` cerrado, `chevron-down` abierto), `aria-expanded`, que al abrir muestra sus grupos por repo.
- **Lista celular (<768 px)**: Barra de filtros con solo el `Input` a ancho completo (44 px, `font-size` 16 px); las chips de etapa viven en el Sidebar, sección nueva **Etapas** (una fila de 44 px por etapa con punto de color, nombre, conteo y tilde; multi-selección, no cierra al elegir). La lista se parte con el mismo Encabezado de repo (fila de 28 px, `position: sticky` arriba) y el mismo Grupo colapsado al final.
- **Tira y nodo**: cada `ChipRef` de referencia del nodo lleva delante el **Ícono de ancla** (lucide 14 px en `--color-fg-muted`: `circle-dot` issue, `messages-square` grill, `file-text` spec; `aria-label` «ancla: issue|grill|spec» y Tooltip). En la tercera fila del nodo, después del badge de sub-estado, el **Badge «sin traza»** (`Badge` `neutral`) cuando `sinHistoria`.
- **Fila de la lista**: punto de etapa, Ícono de ancla, título con elipsis, Badge «sin traza» si aplica, badge de sub-estado, antigüedad. Una sola línea; lo que no entra se trunca.
- **Detalle**: los chips de la cabecera llevan el Ícono de ancla y, si `sinHistoria`, el Badge «sin traza» más una **Nota** `sm` en `--color-fg-muted` debajo de los chips: «Sin traza en events.jsonl: las etapas salen de los archivos y de GitHub, y las fechas son de modificación».
- **Vacío por filtros**: `EmptyState` «Ningún trabajo coincide con los filtros» con `Button secondary` «Limpiar filtros».

```text
Desktop
┌ sdd  [repo ▾]  (Activos | Con traza | Todos)                 ⟳ 2 min  (☀ ☾ ⊙) ┐
├ [🔍 Buscar trabajo, #NN o repo     ×]  [Idea 15] [Grill 52] [Spec 64] [Run 1] [PR 4] [Mergeado 3] [Publicado 8] ┤
│ CHICHEX/WAICA ······································································ 23 │
│  ┌ ⊙ #203 ┐──┌ ▭ grill ┐──┌ ▭ spec ┐──┌ run ┐──┌ 5/5 ┐──┌ 5/5 ┐                      │
│ PRAMAESTUDIO/PLATFORM ·············································································· 9 │
│  ...                                                                                                   │
│ ▸ Sin traza (115)                                                     ← solo con «Todos», colapsado  │
└───────────────────────────────────────────────────────────────────────────────────────┘

Celular                             Sidebar
┌ ☰  sdd · todos los repos    ⟳ 2 min ┐   ┌ sdd                      × ┐
├ [🔍 Buscar                       ×] ┤   │ REPO    …                  │
│ CHICHEX/WAICA                     23 │   │ MOSTRAR Activos ✓ · …      │
│ ● ⊙ Issue #203: …   [5/5]   3 d      │   │ ETAPAS  ● Grill 52  ✓      │
│ ● ▭ 3D spot lights  [sin traza] 1 d  │   │         ● Spec 64          │
│ ▸ Sin traza (115)                    │   │ TEMA    (☀ ☾ ⊙)            │
└──────────────────────────────────────┘   └────────────────────────────┘
```

## Comportamiento esperado
- **CA-1** `GET /api/workspace?filtro=con-traza` devuelve solo trabajos con `sinHistoria=false` (terminados incluidos); `filtro=activos` (default) devuelve solo `terminado=false` y `sinHistoria=false`; `filtro=todos` devuelve todo; cualquier otro valor responde 400 con `filtro debe ser activos, con-traza o todos`. El modelo (`Terminado`, `SinHistoria`) no cambia. — ALTA
- **CA-2** Función pura `agrupar(trabajos)` en `web/src/lib/`: devuelve grupos por `repo` ordenados por id ascendente; dentro de cada grupo, trabajos ordenados por etapa de más avanzada a menos (índice de `ETAPA_INFO` descendente) y, a igual etapa, por `ultimaActividad` descendente. — BAJA (ALTA con el runner; ver Verificabilidad)
- **CA-3** El filtro «Mostrar» tiene tres opciones, en este orden: `Activos`, `Con traza`, `Todos`, en el `Segmented` de desktop y en la sección Mostrar del Sidebar; `readFilters` acepta `filtro=con-traza` y cae a `activos` ante un valor desconocido; `filtersQuery` lo escribe; `useWorkspace` lo pasa al server. — BAJA (ALTA con el runner)
- **CA-4** Badge «sin traza» (`Badge` `neutral`, texto literal «sin traza») visible cuando `sinHistoria=true` en: la tercera fila del `NodoCard`, la fila de `ListaTrabajos` y los chips de cabecera del `Detalle`; en el Detalle además la Nota del `## Diseño`. Con `sinHistoria=false` ninguno aparece. — BAJA (lógica) · NULA (aspecto)
- **CA-5** Con `filtro=todos`, los trabajos con `sinHistoria=true` no se mezclan con los demás: van al final en el Grupo colapsado «Sin traza (N)» (N = cantidad tras aplicar etapa y búsqueda), cerrado por default, con `aria-expanded`; al abrirlo muestra esos trabajos con el mismo agrupado por repo. Con `activos` y `con-traza` el grupo no existe. Con N = 0 no se dibuja. — BAJA (lógica) · NULA (aspecto)
- **CA-6** Board y lista se dibujan por grupos de repo según CA-2, con un Encabezado de repo por grupo (`owner/repo` y conteo) que en el board cruza las siete columnas y en la lista es sticky; si la respuesta trae trabajos de un solo repo, el encabezado se omite. Los siete carriles y sus conteos de nodos siguen siendo los del total visible. — BAJA (lógica) · NULA (aspecto)
- **CA-7** Ícono de ancla por `ancla.tipo` (`circle-dot` issue, `messages-square` grill, `file-text` spec) en el chip de referencia del nodo, en la fila de la lista y en los chips del detalle, con `aria-label` «ancla: <tipo>» y Tooltip con el mismo texto; nunca solo el ícono sin el texto accesible. — BAJA (lógica) · NULA (aspecto)
- **CA-8** Chips de etapa: una por etapa de `ETAPA_INFO`, con nombre y conteo de trabajos en esa etapa dentro de la respuesta actual (sin aplicar etapa ni búsqueda); multi-selección con `aria-pressed`; la selección se escribe en la URL como `etapa=<ids separados por coma, en el orden de ETAPA_INFO>` y se omite si está vacía; sin selección se muestran todas; con selección, board y lista muestran solo trabajos cuya `etapa` esté en ella. En desktop viven en la Barra de filtros; en celular, en la sección Etapas del Sidebar, que no se cierra al tocar una chip. Un id desconocido en la URL se ignora. — BAJA (ALTA con el runner)
- **CA-9** Búsqueda: `Input` con placeholder «Buscar trabajo, #NN o repo» y botón de limpiar con `aria-label` «Limpiar búsqueda»; cada cambio escribe `q=<texto>` en la URL con `replaceState` (sin entradas de historia) y filtra en el cliente por substring sin distinguir mayúsculas ni acentos sobre `titulo`, `repo`, `ancla.ref` e `id`; `q` vacío se omite de la URL. `font-size` computado ≥16 px en celular. — BAJA (ALTA con el runner)
- **CA-10** Si repo y filtro devuelven trabajos pero etapa y búsqueda dejan cero visibles, board y lista muestran el `EmptyState` «Ningún trabajo coincide con los filtros» con el botón «Limpiar filtros», que vacía `etapa` y `q` (no toca `repo` ni `filtro`). Sin trabajos en la respuesta se mantienen los estados actuales (carriles vacíos, EmptyState de página). — BAJA (ALTA con el runner)
- **CA-11** Los cuatro filtros `?repo=&filtro=&etapa=&q=` se leen al cargar, se conservan en `trabajoPath` al abrir el detalle y al cerrarlo, y sobreviven a una recarga; el skeleton y el stream (`/api/stream`) no los pisan. — BAJA (ALTA con el runner)
- **CA-12** Teclado y accesibilidad: chips, Input, botón de limpiar y Grupo colapsado se alcanzan con Tab y muestran `:focus-visible`; los encabezados de repo son `h3` con el nombre del repo; ningún `<select>` nativo (`grep -rn '<select' web/src` vacío) y ningún hex fuera de `index.css` (gate de CA-26 de la spec anterior = 0). — NULA (recorrido humano) salvo los dos greps (ALTA)
- **CA-13** `web/DESIGN_SYSTEM.md` documenta, dentro de `<!-- design-system:ajustes:start -->` y `<!-- design-system:ajustes:end -->` y con sus tokens, las piezas nuevas: «Barra de filtros», «Chip de etapa», «Badge sin traza», «Encabezado de repo», «Grupo colapsado» e «Ícono de ancla»; el resto del doc y sus markers quedan intactos. — ALTA (grep)

## Plan de entrega
| Capa | Etapa | CAs | Justificacion |
|---|---|---|---|
| 1 | filtros-traza | CA-1 a CA-13 | Un solo PR apilado sobre #6 (`--base sdd/visor-flujo-sdd/5-ui-detalle`): el cambio de API es un valor más del mismo parámetro, el modelo no se toca, y el front es un conjunto cohesivo de filtros y marcas que se verifica con `go test` + `npm test` + protocolo humano. Partirlo dejaría una capa de API sin consumidor. |

## Fuera de alcance
- Cambiar qué cuenta como `terminado` (inferencia 4 de la spec anterior) o el agrupado por ancla (`internal/model/group.go`).
- Persistir filtros fuera de la URL (localStorage, perfil) o filtros por sub-estado, fecha o autor.
- Backfill de la traza para trabajos previos al 15 de octubre: el visor marca la ausencia, no la inventa.
- Filtrado de etapa y búsqueda en el server, paginación, o cambios en `/api/trabajos/{id}`, el stream y el poll de GitHub.
- Cambios al doc generado de `web/DESIGN_SYSTEM.md` fuera de «Ajustes de este proyecto».
- Mergear el stack (`/sdd-land`, humano) y `/sdd-init --update`.

## Inferencias
| # | Inferencia | Eleccion propuesta | Alternativa razonable | Confianza | Resolucion |
|---|---|---|---|---|---|
| 1 | Marcar la falta de traza | Badge «sin traza» en tira, lista y detalle sobre `sinHistoria` | Solo en el detalle | alta | elegida por usuario: badge en los tres |
| 2 | Qué significa «activos» | con traza y no terminado | literal del grill D16 (solo terminado) | alta | elegida por usuario: con traza y no terminado |
| 3 | Dónde queda lo sin traza | en «todos», en un grupo colapsado «Sin traza (N)» al final | ocultarlo del todo | alta | elegida por usuario: grupo colapsado |
| 4 | Orden del board | agrupado por repo y, dentro, por etapa, con ícono de ancla por tipo | solo por última actividad | alta | elegida por usuario: repo → etapa + ícono |
| 5 | Filtros nuevos | chips de etapa y búsqueda por texto, en la URL | un Select de etapa | alta | elegida por usuario: chips + búsqueda |
| 6 | Cómo se ofrece «solo con traza» | tercera opción `con-traza` del Segmented «Mostrar» (`opciones.ts:8`), server-side | toggle aparte junto al Segmented | media | [ASSUMED] |
| 7 | Dónde se filtran etapa y búsqueda | en el cliente, sobre la respuesta de `/api/workspace`; `repo` y `filtro` siguen en el server (`api.go:66-79`) | parámetros `etapa` y `q` en la API | alta | [ASSUMED] |
| 8 | Orden entre repos y dentro del repo | repos por id ascendente; dentro, etapa más avanzada primero y luego `ultimaActividad` desc (`build.go:30` queda igual) | repos por última actividad; Idea primero | media | [ASSUMED] |
| 9 | Encabezado con un solo repo | se omite cuando la respuesta trae un solo repo (por ejemplo con `?repo=`) | mostrarlo siempre | alta | [ASSUMED] |
| 10 | Estado del Grupo colapsado | colapsado por default, estado local de React (no en la URL); existe solo con `todos` | abierto en la URL (`sintraza=1`) | media | [ASSUMED] |
| 11 | Íconos de ancla | lucide `circle-dot`, `messages-square`, `file-text` dentro del `ChipRef` (`marcas.tsx:17`) | letra I/G/S en el chip | media | [ASSUMED] |
| 12 | Ubicación de chips y búsqueda | Barra de filtros bajo la cabecera (desktop: Input + chips; celular: solo Input); chips de celular en el Sidebar, sección Etapas sin cerrar al elegir | chips en las cabeceras de carril | media | [ASSUMED] |
| 13 | Semántica de la búsqueda | substring, sin mayúsculas ni acentos (NFD), sobre `titulo`, `repo`, `ancla.ref`, `id`; sin debounce (la lista es local) | solo título; prefijo | alta | [ASSUMED] |
| 14 | Conteo de cada chip | trabajos por `etapa` en la respuesta actual, sin aplicar etapa ni `q` | conteo tras aplicar `q` | media | [ASSUMED] |
| 15 | Badge en el nodo | tercera fila del `NodoCard`, después del sub-estado; si no entra se trunca (regla «nada se apila») | fila propia | alta | [ASSUMED] |
| 16 | Input del kit | `npx shadcn@latest add input` sin `init`, como los 8 componentes existentes (`components.json` a mano) | `<input>` con clases propias | alta | [ASSUMED] |
| 17 | Dónde se documentan las piezas nuevas | «Ajustes de este proyecto» de `web/DESIGN_SYSTEM.md` | regenerar el doc con `/design-system` | alta | [ASSUMED] |
| 18 | Fixture para `con-traza` | extender `TestWorkspace_FiltroActivos` (`api_test.go:47`) con una spec `.md` sin eventos (trabajo `sinHistoria`) y asertar los tres filtros | fixture nuevo en `testdata/` | alta | [ASSUMED] |
| 19 | Grupo «Sin traza» en celular | mismo botón y mismo agrupado que en desktop, dentro de la lista | ocultarlo en celular | alta | [ASSUMED] |

## Verificabilidad
Mixto, anclado en el contrato del 2026-10-25:
- **ALTA: CA-1 y CA-13**, más los greps de CA-12. `go test -count=1 ./internal/api` con `httptest` (escalón 2 del contrato) y grep sobre `web/DESIGN_SYSTEM.md` y `web/src`.
- **BAJA: CA-2, CA-3, CA-5, CA-6, CA-8, CA-9, CA-10, CA-11 y la lógica de CA-4 y CA-7.** El contrato sigue diciendo «el front no tiene runner de tests» (Gaps), así que por contrato solo llegan `npx tsc -b`, `npm run lint` y el build. De hecho la rama base trae Vitest + Testing Library con 69 tests (`web/src/*.test.tsx`, `npm test`), que el run usa igual; serían ALTA cuando `/sdd-init --update` registre `npm test`.
- **NULA: el aspecto de CA-4 a CA-7, CA-12 (recorrido con Tab) y el uso en celular.** «La UI renderizada… hace falta un navegador» (Verificacion autonoma); sin Playwright ni MCP de navegador. Protocolo humano abajo.
- Políticas: `commits`, `max-lineas-archivo` (los archivos tocados tienen entre 20 y 232 líneas; la lógica nueva va a `web/src/lib/filtrar.ts` y componentes propios para no acercar `App.tsx` ni `Sidebar.tsx` al límite), `no-any-ts`, `gofmt` y `no-panic` aplican como gates. Sin política de tamaño de PR ni de dependencias: `input` de shadcn se justifica en el PR.

## Plan de verificacion
Mecanismo elegido: test de API en Go con `httptest`; Vitest + Testing Library por rol y texto para la lógica del front (funciones puras en `web/src/lib/filtrar.ts` y comportamiento con `render`); greps para tokens y doc; protocolo humano para aspecto, celular y teclado. Alternativas: (más exhaustiva) sumar Playwright con capturas del board en los tres filtros; (más barata) sin tests de componentes, solo `filtrar.ts` puro y protocolo humano.

Por CA:
- CA-1: `go test -count=1 ./internal/api -run 'TestWorkspace_Filtro'`: en el repo temporal de `TestWorkspace_FiltroActivos` sumar `.sdd/specs/c.md` con marker `state=approved` y sin eventos (`sinHistoria`), y asertar `activos` = {b}, `con-traza` = {a, b}, `todos` = {a, b, c}, `?filtro=otro` = 400 con el mensaje nuevo.
- CA-2, CA-9 (normalización), CA-8 (parse/serialize de `etapa`): `cd web && npm test -- lib` sobre `filtrar.ts` y `router.ts`: `agrupar` (orden de repos, etapa desc, empate por fecha), `coincide(q, trabajo)` («Ñandú» ≈ «nandu», «#203», `owner/repo`), `readFilters('?etapa=spec,grill,zzz')` → `['grill','spec']`, `filtersQuery` omite vacíos.
- CA-3, CA-8, CA-9, CA-10, CA-11: `npm test -- filtros`: cambiar a «Con traza» pide `/api/workspace?filtro=con-traza` y escribe la URL; tocar la chip «Grill» escribe `?etapa=grill` y oculta las tiras de otras etapas; escribir «203» en «Buscar» deja una tira; «Limpiar filtros» vacía `etapa` y `q` y conserva `repo`; abrir y cerrar el detalle conserva los cuatro parámetros.
- CA-4, CA-5, CA-6, CA-7: `npm test -- Board` y `npm test -- App`: fixture con dos repos y un trabajo `sinHistoria`; en `todos` existe el botón «Sin traza (1)» con `aria-expanded=false` y la tira no está en el DOM hasta abrirlo; con `activos` el botón no existe; hay un `h3` por repo en orden alfabético, ninguno con un solo repo; `getByText('sin traza')` en nodo, fila y detalle; `getByLabelText('ancla: issue')`.
- CA-12: `grep -rn '<select' web/src` vacío; `git grep -IhoE '#([0-9a-fA-F]{3}){1,2}' -- 'web/src/**' ':!web/src/index.css' | wc -l` = 0; el recorrido con Tab es humano.
- CA-13: `sed -n '/design-system:ajustes:start/,/design-system:ajustes:end/p' web/DESIGN_SYSTEM.md | grep -c -E 'Barra de filtros|Chip de etapa|Badge sin traza|Encabezado de repo|Grupo colapsado|Ícono de ancla'` = 6, y `git diff --stat sdd/visor-flujo-sdd/5-ui-detalle -- web/DESIGN_SYSTEM.md` toca solo líneas entre los markers.
- Escalera: `gofmt -l .`, `go vet ./...`, `npx tsc -b`, `npm run lint`, `npx oxlint -D typescript/no-explicit-any`, `make test`, `make build`; app levantada contra el fixture: `./bin/sdd --dir "$(bash testdata/mkworkspace.sh)" --no-github --cache-dir "$(mktemp -d)"` y `curl -sf 'localhost:4321/api/workspace?filtro=con-traza' | jq '[.trabajos[].sinHistoria] | any'` = false.
- Gates de políticas: los de `.sdd/project.md`, con `origin/sdd/visor-flujo-sdd/5-ui-detalle` como base en vez de `origin/main` para `commits` y `max-lineas-archivo`.

Protocolo de prueba humana (aspecto de CA-4 a CA-7, CA-12 y celular):
1. Con el binario de la rama (`make build && ./bin/sdd --dir /home/chiche/workspace --host 0.0.0.0`), abrir el board en desktop con el filtro default: ver 23 trabajos (con traza y no terminados) partidos por repo con encabezados `OWNER/REPO · N`, dentro de cada repo los más avanzados primero; chips con conteos en la Barra de filtros; ningún «sin traza».
2. Elegir «Todos»: aparece «▸ Sin traza (115)» al final, cerrado; abrirlo muestra los trabajos viejos agrupados por repo con el badge «sin traza» en cada nodo. Elegir «Con traza»: los terminados (`waica#203`) vuelven, el grupo desaparece.
3. Activar las chips Grill y Spec: la URL muestra `?etapa=grill,spec` y el board deja solo esas etapas; recargar la conserva.
4. Escribir «203» en Buscar con «Todos»: queda `waica#203`; limpiar con la X. Buscar «zzz»: EmptyState «Ningún trabajo coincide con los filtros» y «Limpiar filtros» restaura.
5. Abrir un trabajo sin traza (por ejemplo `platform#53`): badge «sin traza» junto a los chips y la nota debajo; el ícono de ancla (`circle-dot`) tiene tooltip «ancla: issue». Esc vuelve con los filtros.
6. Celular por Tailscale: Barra de filtros con el Input a ancho completo sin zoom al enfocar; Sidebar con la sección Etapas (filas con tilde que no cierran el menú); lista con encabezados de repo sticky y «Sin traza (N)» al final; nada baja de línea.
7. Tab desde la cabecera: Select → Mostrar → Buscar → X → chips → grupo colapsado → nodos; cada parada con anillo visible; Enter en una chip la activa; Enter en «Sin traza» la abre.

## Riesgos y gaps
- **Revisar primero: inferencia 8** (orden dentro del repo: más avanzada primero) e **inferencia 12** (chips en una Barra de filtros nueva; en celular en el Sidebar). Definen cómo se ve el board de un vistazo.
- **Desviación del grill 2026-10-23 (D16)**: «activos» deja de ser «no terminado» y pasa a «con traza y no terminado», decidido por el usuario el 2026-10-28 tras probar contra su workspace. La spec anterior queda `implemented`; esta no la reemplaza, la extiende.
- **Base distinta de `main`**: `/sdd-run` debe correr con `--base sdd/visor-flujo-sdd/5-ui-detalle`; el PR se apila sobre #6 y `/sdd-land` del PR nuevo aterriza las seis capas. El contrato dice «/sdd-run ramifica desde `origin/main`»: el flag lo pisa.
- **Contrato desactualizado**: el front ya tiene Vitest (69 tests en la base) pero el contrato no lo registra; por eso los CA de front quedan BAJA. `/sdd-init --update` después del land.
- **`max-lineas-archivo` 300**: `web/src/lib/api.ts` tiene 232 y no debería crecer (no hay campos nuevos); `App.tsx` (103) y `Sidebar.tsx` (92) suman props: la lógica de filtrado y agrupado va a `filtrar.ts` y los controles a componentes propios.
- **Dependencia nueva**: componente `input` de shadcn (radix no agrega paquetes). Sin política de dependencias activa; se justifica en el PR.
- **Fixture**: `testdata/mkworkspace.sh` ya trae artefactos sin eventos (`platform/.git/sdd/grills/2026-10-08-backlog-norte…`), así que el fixture del escalón 4 tiene al menos un `sinHistoria`; el test de API usa su repo temporal propio (inferencia 18).
- Conteos del protocolo humano (23, 115, 15 ideas) son los del workspace del 2026-10-28 y cambian con cada grill o run nuevo.

## Desviaciones
- 2026-10-28 [DEVIATION] Riesgos y escalón 4 del Plan de verificacion: el fixture de `testdata/mkworkspace.sh` no tiene ningún trabajo `sinHistoria` (el grill `backlog-norte` de platform se agrupa en un issue con eventos): `activos`, `con-traza` y `todos` devuelven 12 trabajos cada uno, así que la prueba `jq '[.trabajos[].sinHistoria] | any'` = `false` pasa sin discriminar. La evidencia que discrimina CA-1 es el test de Go con la spec `c.md` sin eventos. El fixture no se toca: sumarle un trabajo sin traza cambiaría los 12 trabajos que cuentan los tests de la spec anterior. No cambia el alcance.
- 2026-10-28 [DEVIATION] CA-6: «los del total visible» se lee como todos los trabajos que pasan etapa y búsqueda, incluidos los del grupo «Sin traza» aunque esté cerrado; los conteos de los carriles no cambian al abrir o cerrar el grupo. No cambia el alcance.
- 2026-10-28 [DEVIATION] CA-7 e inferencia 11: en el detalle el ícono va dentro de la caja del chip del ancla pero fuera del `<a>`, para que el link conserve su nombre accesible (`#57`); el chip del repo no lleva ícono porque no es un ancla. No cambia el alcance.
- 2026-10-28 [DEVIATION] CA-9: el botón de limpiar aparece solo cuando la búsqueda tiene texto; en el recorrido con Tab del protocolo (paso 7) la parada «X» existe solo con texto escrito. No cambia el alcance.
- 2026-10-28 [DEVIATION] `## Diseño`, sección Etapas del Sidebar: la tilde va a la izquierda, como en las filas de Repo y Mostrar (mismo componente `Fila`), y no a la derecha como en el mock. No cambia el alcance.
- 2026-10-28 [DEVIATION] Inferencia 16: `npx shadcn@latest add input` generó el input de shadcn (36 px, ring propio, `md:text-sm`); se adaptó a los tokens del doc como los 8 componentes previos (32 px, `--color-border`, `--color-bg`, placeholder `--color-fg-muted`, foco con el anillo global de `:focus-visible`). Sin paquetes npm nuevos. No cambia el alcance.

## Resultado de ejecucion (2026-10-28 · HEAD e7180be)
Verificación final sobre `e7180be`, base `origin/sdd/visor-flujo-sdd/5-ui-detalle` (`856a3d7`). Spec aprobada al correr (`draft` con `--assume`). `npm test` no figura en el contrato: se usó igual como mecanismo de los CA de front (Vitest + Testing Library de la base).

| CA | Capa | Estado | Evidencia |
|---|---|---|---|
| CA-1 | 1/1 | verificado | `go test -count=1 ./internal/api -run 'TestWorkspace_Filtro'`: `TestWorkspace_FiltroActivos` con 4 subtests verdes (activos por default y explícito = {b}, `con-traza` = {a, b}, `todos` = {a, b, c}) y `?filtro=otro` → 400 «filtro debe ser activos, con-traza o todos». App levantada contra el fixture en `:4321`: `curl -sf 'localhost:4321/api/workspace?filtro=con-traza' \| jq '[.trabajos[].sinHistoria] \| any'` = `false` (ver desviación del fixture); `?filtro=otro` → 400 con el mensaje |
| CA-2 | 1/1 | verificado | `npm test -- lib`: 33/33; `agrupar` en `src/lib/filtrar.test.ts` (repos por id ascendente; etapa más avanzada primero y empate por `ultimaActividad` desc; vacío) |
| CA-3 | 1/1 | verificado | `npm test -- lib` (`readFilters('?filtro=con-traza')`, desconocido → `activos`, `filtersQuery`) y `npm test -- filtros`: orden Activos · Con traza · Todos, elegir «Con traza» escribe `?filtro=con-traza` y pide `/api/workspace?filtro=con-traza`, fila en el Sidebar |
| CA-4 | 1/1 | verificado (lógica) · pendiente humano (aspecto) | `npm test -- Board`: badge «sin traza» en la fila del sub-estado del nodo, después de él, y ausente con traza. `npm test -- App`: fila de la lista; detalle con badge y nota; sin ellos con traza. Aspecto: protocolo, pasos 2 y 5 |
| CA-5 | 1/1 | verificado (lógica) · pendiente humano (aspecto) | `npm test -- Board`: «Sin traza (1)» al final con `aria-expanded=false`, tira fuera del DOM hasta abrirlo, h3 por repo al abrirlo; ausente con `activos` y `con-traza`; N tras `q`; N = 0 no se dibuja. `npm test -- App` (lista) y `npm test -- lib` (`armarVista`). Aspecto: protocolo, paso 2 |
| CA-6 | 1/1 | verificado (lógica) · pendiente humano (aspecto) | `npm test -- Board`: h3 `o/a` y `o/b` con su conteo y sus tiras debajo; sin h3 con `?repo=`; carril «Grill 2» con el grupo sin traza cerrado. `npm test -- App`: h3 en la lista. Cruce de carriles y sticky: protocolo, pasos 1 y 6 |
| CA-7 | 1/1 | verificado (lógica) · pendiente humano (aspecto) | `npm test -- Board`: `ancla: issue`, `ancla: grill` y `ancla: spec` en el chip del nodo; tooltip `ancla: spec` al pasar el puntero. `npm test -- App`: fila y detalle; el link `#57` conserva su nombre. Íconos: protocolo, paso 5 |
| CA-8 | 1/1 | verificado | `npm test -- filtros`: 7 chips en orden con el conteo de la respuesta; «Grill» → `?etapa=grill` y solo sus tiras; multi-selección → `?etapa=idea,pr` con conteos fijos y vaciar vuelve todo; `zzz` ignorado; la sección Etapas del Sidebar no se cierra. `npm test -- lib`: `readFilters('?etapa=spec,grill,zzz')` → `['grill', 'spec']` |
| CA-9 | 1/1 | verificado (lógica) · pendiente humano (`font-size` en celular) | `npm test -- filtros`: placeholder, `?q=57` con `replaceState` y sin pedir al server, «Limpiar búsqueda»; en celular solo la búsqueda. `npm test -- lib`: `coincide` («nandu» ≈ «Ñandú», mayúsculas, `#57`, `owner/repo`, id, vacío). El `font-size` computado ≥ 16 px lo ve el protocolo, paso 6: jsdom no aplica el CSS de Tailwind |
| CA-10 | 1/1 | verificado | `npm test -- filtros`: EmptyState «Ningún trabajo coincide con los filtros»; «Limpiar filtros» deja `?repo=o%2Fa&filtro=todos`; sin trabajos en la respuesta, carriles vacíos |
| CA-11 | 1/1 | verificado | `npm test -- filtros`: los cuatro se leen al cargar, viajan en `trabajoPath`, vuelven al cerrar con Esc y sobreviven a desmontar y volver a montar; el skeleton y un `snapshot` del stream no los tocan |
| CA-12 | 1/1 | verificado (greps) · pendiente humano (Tab) | `grep -rn '<select' web/src` vacío (exit 1); `git grep -IhoE '#([0-9a-fA-F]{3}){1,2}' -- 'web/src/**' ':!web/src/index.css' \| wc -l` = 0; h3 por repo en `npm test -- Board`. Recorrido con Tab: protocolo, paso 7 |
| CA-13 | 1/1 | verificado | `sed -n '/design-system:ajustes:start/,/design-system:ajustes:end/p' web/DESIGN_SYSTEM.md \| grep -c -E '…'` = 6; `git diff -U0 sdd/visor-flujo-sdd/5-ui-detalle -- web/DESIGN_SYSTEM.md`: un solo hunk `@@ -124,0 +125,8 @@`, entre los markers (líneas 124 y 133); `npm test -- tokens` 31/31 |
| POL-commits | 1/1 | cumplida | gate sobre `git log --format=%s origin/sdd/visor-flujo-sdd/5-ui-detalle..HEAD` vacío |
| POL-max-lineas-archivo | 1/1 | cumplida | gate de `wc -l` sobre `origin/sdd/visor-flujo-sdd/5-ui-detalle...HEAD` vacío (máximo 233, `web/src/filtros.test.tsx`) |
| POL-no-any-ts | 1/1 | cumplida | `npx oxlint -D typescript/no-explicit-any`: exit 0 (3 advertencias previas, ninguna en archivos tocados) |
| POL-gofmt | 1/1 | cumplida | `gofmt -l .` vacío |
| POL-no-panic | 1/1 | cumplida | `git grep -n 'panic(' -- '*.go' ':!cmd/' ':!*_test.go' ':!web/embed.go'` vacío |

Escalera hasta el escalón 4: `gofmt -l .` vacío, `go vet ./...` exit 0, `npx tsc -b` exit 0, `npm run lint` exit 0 (3 advertencias previas), `make test` y `go test -count=1 ./...` verdes, `npm test` 120/120 (51 nuevos sobre los 69 de la base), `make build` (`bin/sdd` y `web/dist/index.html`), app levantada contra el fixture con `--no-github` (las rutas del SPA con `?repo=&filtro=&etapa=&q=` responden 200 `text/html`). Guías aplicadas sin gate: `estilo-go` (Uber) y `.sdd/coding-policies.md`. Sin E2E de navegador: el aspecto, el celular y el teclado quedan en el protocolo humano.
