# Spec — Visor: categorías de trabajos y grafo vertical en celular
<!-- Generada por /sdd-spec el 2026-10-29. Fuente: pedido libre. Estado: implementada -->
<!-- SDD-Tracking: version=1; type=spec; state=implemented; issue=none; grill=none; superseded-by=none -->

## Contexto
Probado en el teléfono contra `/home/chiche/workspace` (2026-10-29, binario de PR #9, `6022605`), el filtro «Mostrar» (`activos` · `con-traza` · `todos`, `internal/api/api.go:66-79`, `web/src/components/opciones.ts`) no deja quedarse con «lo que tiene traza y está incompleto». Dos causas. Primera, de semántica: un issue abierto sin trabajo SDD entra como Idea con un grupo vacío (`internal/model/group.go` `ideas()`), y `SinHistoria` exige «tiene ítems y ninguno es evento» (`internal/model/build.go:61`), así que un grupo vacío cuenta como «con traza»: hoy `activos` devuelve 27 trabajos, de los cuales 19 son issues sin ningún evento, y los que de verdad tienen eventos y no terminaron son 6 (5 en PR y 1 mergeado sin cerrar). Segunda, de modelo: el control mezcla dos ejes (traza y terminado) en tres valores y no permite combinar. Por otro lado, el grafo del detalle dibuja cada nodo a 120 px fijos con 64 px de margen (`web/src/components/detalle/grafo.ts` `GRAFO_PASO`, `X0`) dentro de un `overflow-x-auto` (`Grafo.tsx:21`): cuatro nodos miden 488 px, en el sheet de 390 px desbordan, aparece la barra de scroll y el primer nodo queda cortado. El Gantt no tiene el problema. Esta spec sigue a `visor-lista-etapas`: su base es la rama `sdd/visor-lista-etapas` (PR #9), no `main`, y el PR resultante se apila sobre #9. Reemplaza el filtro de CA-1 y CA-3 de `visor-anclas-traza-filtros` (sin mergear) y la decisión 16 del grill `2026-10-23-visor-flujo-sdd` («activos / todos»), por decisión del usuario del 2026-10-29; la decisión 23 (Idea = issues abiertos) se mantiene.

## Diseño
Superficie `web` (`web/src`), con sistema: tokens en `web/src/index.css` (`@theme`), componentes en `web/src/components/ui/` y doc `web/DESIGN_SYSTEM.md` (dirección `sobrio-indigo`, con «Ajustes de este proyecto» de las dos specs anteriores).

- Reusa: las chips de etapa de `BarraFiltros.tsx` (`ChipsEtapa`: botón de 28 px, `aria-pressed`, nombre y conteo en mono) y las filas con tilde del `Sidebar.tsx` (`Fila` con `toggle`), `EtapaDot`, `etapaInfo`, `fechaCorta`, tokens `--color-fg`, `--color-fg-secondary`, `--color-fg-muted`, `--color-primary-text`, `--color-border-strong`, `--color-stage-*`.
- Nuevo: **Chips de categoría** («Mostrar»): mismo dibujo que las chips de etapa, en la Barra de filtros de desktop antes de las etapas, separadas por un divisor vertical de 1 px `--color-border`; `role="group"` `aria-label` «Mostrar». En celular, sección **Mostrar** del Sidebar con filas de 44 px tildables (tilde, nombre, conteo) que no cierran el menú, como Etapas. El `Segmented` «Mostrar» de la cabecera de desktop desaparece.
- Nuevo: **Grafo vertical** (solo celular, <768 px): un nodo por fila, círculo de 16 px de radio a la izquierda en una columna, y a la derecha, en una sola línea, la etiqueta (`xs`, `--color-fg`, 600 y `--color-primary-text` si es el activo), el sub-estado (`xs`, `--color-fg-secondary`) y la fecha corta (mono `xs`, `--color-fg-muted`); aristas verticales de 1,5 px en `--color-border-strong`; cada rama cerrada en una columna a la derecha de la principal, con una curva desde su nodo de bifurcación. El SVG ocupa el ancho del contenedor (`width="100%"`, `viewBox`) y el alto crece con los nodos; sin scroll horizontal. En desktop el grafo horizontal no cambia.

Referencias a igualar (mocks elegidos por el usuario el 2026-10-29):
```text
Sidebar (celular)                         Barra de filtros (desktop)
MOSTRAR                                   [Buscar…] │ En curso 6  Ideas 19  Terminados 10  Sin traza 125 │ Idea 19  Grill 0 …
 ✓ En curso                      6
 ✓ Ideas                        19        Grafo (celular, 358 px)
   Terminados                   10         ● Spec         borrador    15 oct
   Sin traza                   125         │
                                           ● Run          con PR      15 oct
URL: (default, se omite)                   │
     ?mostrar=en-curso                     ● PR 1/1       mergeado    15 oct
     ?mostrar=en-curso,terminado           │
                                           ◉ Mergeado 1   mergeado    15 oct
                                           rama cerrada: columna a la derecha, curva desde su bifurcación
```

## Comportamiento esperado
- **CA-1** `Trabajo` gana `Categoria` (`json:"categoria"`) con exactamente uno de `idea`, `en-curso`, `terminado`, `sin-traza`: `idea` si el grupo no tiene ítems (issue abierto sin trabajo SDD); si no, `sin-traza` si ningún ítem es evento; si no, `terminado` si `Terminado`; si no, `en-curso`. `Terminado` y `SinHistoria` no cambian de valor. — ALTA
- **CA-2** `GET /api/workspace?mostrar=<ids separados por coma>` devuelve solo los trabajos cuya `categoria` está en la lista; sin el parámetro, la lista es `en-curso,idea`; un id desconocido o una lista vacía responde 400 con `mostrar debe ser una lista de idea, en-curso, terminado o sin-traza`. `?filtro=` deja de existir y se ignora. `?repo=` sigue aplicando antes. — ALTA
- **CA-3** La respuesta de `/api/workspace` suma `categorias`: un objeto con las cuatro claves y el conteo de trabajos de cada categoría tras aplicar `repo` y antes de aplicar `mostrar` (así una categoría sin tildar muestra cuántos esconde). Suma siempre las cuatro claves, con 0 donde no hay. — ALTA
- **CA-4** `Filters.mostrar: Categoria[]` reemplaza a `filtro`: `readFilters` lee `?mostrar=` aceptando solo ids conocidos, sin repetidos y en el orden canónico `en-curso, idea, terminado, sin-traza`; sin el parámetro o sin ningún id válido cae al default `['en-curso','idea']`; `filtersQuery` omite el default y escribe la lista con comas literales; `?filtro=` en un link viejo se ignora. Los cinco filtros `?repo=&mostrar=&etapa=&q=` se conservan al abrir y cerrar el detalle y sobreviven a una recarga. — BAJA (ALTA con el runner; ver Verificabilidad)
- **CA-5** Control «Mostrar»: cuatro opciones con etiqueta y conteo («En curso», «Ideas», «Terminados», «Sin traza», en ese orden, conteos de `categorias`), multi-selección con `aria-pressed`; en desktop como chips en la Barra de filtros (grupo «Mostrar» antes de «Etapas»), y la cabecera ya no tiene el `Segmented` «Mostrar»; en celular como filas tildables de la sección Mostrar del Sidebar, que no se cierra al tocar una. Siempre queda al menos una categoría elegida: tocar la última activa no la destilda. — BAJA (lógica) · NULA (aspecto)
- **CA-6** `useWorkspace` pide `/api/workspace?mostrar=<lista>` con la selección actual y refetchea al cambiarla; `armarVista` separa al final, en el grupo colapsado «Sin traza (N)», los trabajos con `categoria === 'sin-traza'` siempre que estén en la respuesta (ya no depende de `todos`); los de `idea` van en línea con los demás. El badge «sin traza» sigue dependiendo de `sinHistoria`. — BAJA (ALTA con el runner)
- **CA-7** `grafoLayout(t, 'vertical')` (o parámetro equivalente): la rama viva es una columna (`x` constante, `y` creciente con un paso fijo), cada rama cerrada es una columna a la derecha (`x` mayor) cuyo primer nodo se une con una arista curva a su nodo de bifurcación (el mismo que en horizontal), `ancho` es constante (no depende de la cantidad de nodos) y `alto` crece con la fila más baja. La variante horizontal devuelve exactamente lo de hoy. — BAJA (ALTA con el runner)
- **CA-8** En celular (`desktop=false`, el mismo `useMediaQuery('(min-width: 768px)')` de `App.tsx`) el `Grafo` se dibuja vertical: el `svg` tiene `width="100%"` y `viewBox`, su contenedor no tiene `overflow-x-auto` y por cada nodo muestra etiqueta, sub-estado y fecha como `<text>`; `data-grafo-nodo`, `data-rama` y `data-activo` se conservan. En desktop el grafo horizontal y su contenedor no cambian. — BAJA (lógica) · NULA (aspecto)
- **CA-9** `web/DESIGN_SYSTEM.md`, dentro de los markers de «Ajustes de este proyecto», documenta «Chips de categoría» y «Grafo vertical» con sus tokens, y la entrada «Grupo colapsado» pasa a decir que aparece cuando la categoría Sin traza está elegida. Nada fuera de los markers cambia. — ALTA

## Plan de entrega
| Capa | Etapa | CAs | Justificacion |
|---|---|---|---|
| 1 | categorias-grafo | CA-1 a CA-9 | Un solo PR apilado sobre #9 (`--base sdd/visor-lista-etapas`). Las categorías (CA-1 a CA-6) y el grafo (CA-7 y CA-8) son independientes, pero partirlos sumaría un noveno PR a un stack que ya tiene ocho y obligaría a `gh stack` sobre una base que no es `main`; los dos grupos se verifican con `go test` + `npm test` y el protocolo humano. |

## Fuera de alcance
- Cambiar qué cuenta como `Terminado` o como `SinHistoria`, o cómo se derivan etapas y sub-estados.
- Leer la spec desde la rama del PR, colapsar las Ideas en un grupo, o guardar la selección fuera de la URL.
- Tocar el grafo horizontal de desktop, el Gantt o el board.
- Permitir cero categorías elegidas (ver inferencia 5).

## Inferencias
| # | Inferencia | Eleccion propuesta | Alternativa razonable | Confianza | Resolucion |
|---|---|---|---|---|---|
| 1 | ¿Qué reemplaza al filtro «Mostrar»? | Cuatro categorías excluyentes y exhaustivas, multi-selección con conteos | Arreglar la semántica y dejar los tres valores · dos llaves Terminados / Sin traza | alta | elegida por usuario: categorías (2026-10-29) |
| 2 | ¿Qué hacer con el grafo en celular? | Vertical, sin scroll horizontal | Sacarlo en celular · compactar el horizontal | alta | elegida por usuario: vertical (2026-10-29) |
| 3 | ¿Ids, orden y etiquetas? | `en-curso` «En curso», `idea` «Ideas», `terminado` «Terminados», `sin-traza` «Sin traza», en ese orden | Ids en plural | media | [ASSUMED] |
| 4 | ¿Default? | `en-curso,idea`; omitido en la URL y en la request cuando es el default | Solo `en-curso` | media | [ASSUMED] |
| 5 | ¿Se puede destildar todo? | No: queda al menos una; tocar la última no hace nada | Permitir vacío y mostrar el EmptyState «Ningún trabajo coincide» | media | [ASSUMED] |
| 6 | ¿Qué pasa con `?filtro=`? | Desaparece de la API y del front; un link viejo cae al default | Mantenerlo como alias de `mostrar` | alta | [ASSUMED] |
| 7 | ¿De dónde salen los conteos de las categorías? | Del server (`categorias`), por `repo` y antes de `mostrar` | Pedir siempre todo y contar en el cliente | media | [ASSUMED] |
| 8 | ¿Dónde viven las chips en desktop? | Barra de filtros, grupo «Mostrar» antes de «Etapas», con divisor; el `Segmented` de la cabecera se va | Multi-selección en la cabecera | media | [ASSUMED] |
| 9 | ¿La sección Mostrar del Sidebar cierra el menú? | No: filas `toggle` como Etapas | Cerrar al tocar, como Repo | alta | [ASSUMED] |
| 10 | ¿Cuándo aparece «Sin traza (N)»? | Cuando la categoría Sin traza está elegida; Ideas van en línea | Grupo colapsado también para Terminados | alta | [ASSUMED] |
| 11 | ¿`sinHistoria` cambia para las Ideas? | No: sigue `false`, el badge «sin traza» no cambia | Ponerlo `true` | alta | [ASSUMED] |
| 12 | ¿Geometría del grafo vertical? | Columna de nodos a `x` 24, paso de fila 56 px, textos a la derecha en tres columnas fijas; ramas cerradas a 40 px a la derecha; `viewBox` de 340 de ancho, `width="100%"` con `preserveAspectRatio` `xMinYMin meet` | Textos debajo del nodo como en horizontal | media | [ASSUMED] |
| 13 | ¿Cómo sabe el Grafo que es celular? | Prop `vertical` desde `Detalle` (que ya recibe `desktop`) | `useMediaQuery` dentro del Grafo | alta | [ASSUMED] |
| 14 | ¿Entrega? | Una capa, PR apilado sobre #9 | Dos capas (categorías, grafo) con `gh stack` | media | [ASSUMED] |

## Verificabilidad
Mixto, anclado en el contrato del 2026-10-25:
- **ALTA: CA-1, CA-2, CA-3 y CA-9.** `go test -count=1 ./internal/model ./internal/api` con `httptest` (escalón 2) y greps sobre `web/DESIGN_SYSTEM.md`.
- **BAJA: CA-4, CA-5, CA-6, CA-7 y CA-8.** El contrato sigue sin registrar un runner de front (Gaps); por contrato solo llegan `npx tsc -b`, `npm run lint` y el build. La rama base trae Vitest + Testing Library con 135 tests (`cd web && npm test`), que el run usa igual; serían ALTA cuando `/sdd-init --update` registre `npm test`.
- **NULA: el aspecto de CA-5 y CA-8 y el uso en celular.** «La UI renderizada… hace falta un navegador» (Verificacion autonoma); sin Playwright ni MCP de navegador. Protocolo humano abajo.
- Políticas: `commits`, `max-lineas-archivo` (hoy `api.ts` 232, `filtros.test.tsx` 233, `build.go` 219, `api_test.go` 194: ver Riesgos), `no-any-ts`, `gofmt` y `no-panic` aplican como gates. Sin dependencias nuevas.

## Plan de verificacion
Mecanismo elegido: tests de Go (`model` y `api` con `httptest`); Vitest + Testing Library por rol y texto (funciones puras en `router.ts`, `filtrar.ts` y `grafo.ts`; `render` de `App` con `matchMedia` en `matches: false` para celular, como en `web/src/App.test.tsx:90`); greps para el doc; protocolo humano para aspecto y celular. Alternativas: (más exhaustiva) Playwright con capturas del Sidebar y del detalle a 390 px contra los mocks; (más barata) solo las funciones puras y protocolo humano.

Por CA:
- CA-1: `go test -count=1 ./internal/model -run TestBuild_Categoria`: cuatro grupos en un repo temporal (issue abierto sin ítems → `idea`; spec sin eventos → `sin-traza`; spec con `run.pr-opened` y `land.merged` sin issue → `terminado`; spec con `spec.published` → `en-curso`) y la invariante «exactamente una categoría por trabajo» sobre `TestBuild_TrabajosDeLogsReales`.
- CA-2 y CA-3: `go test -count=1 ./internal/api -run 'TestWorkspace_Mostrar'`, reescribiendo `TestWorkspace_FiltroActivos` como tabla sobre el mismo repo temporal más un issue abierto grabado en el snapshot de GitHub: sin parámetro → {b, issue}; `?mostrar=en-curso` → {b}; `?mostrar=en-curso,terminado` → {a, b}; `?mostrar=sin-traza` → {c}; `?mostrar=idea,en-curso,terminado,sin-traza` → todo; `?mostrar=` y `?mostrar=otro` → 400 con el mensaje nuevo; `?filtro=activos` → se ignora (igual que sin parámetro); `categorias` = `{idea:1, en-curso:1, terminado:1, sin-traza:1}` en todas las respuestas 200 y, con `?repo=`, solo cuenta ese repo.
- CA-4: `cd web && npm test -- lib`: `readFilters('?mostrar=terminado,en-curso,zzz')` → `['en-curso','terminado']`; `readFilters('')` y `readFilters('?mostrar=zzz')` → default; `readFilters('?filtro=con-traza')` → default; `filtersQuery` omite el default y escribe `?mostrar=en-curso,terminado`; `trabajoPath` conserva `mostrar`.
- CA-5 y CA-6: `npm test -- mostrar` (archivo nuevo `web/src/mostrar.test.tsx`, reemplazando el `describe('filtro Mostrar con traza')` de `filtros.test.tsx`): en desktop hay un grupo «Mostrar» con cuatro botones en orden, `aria-pressed` true en En curso e Ideas, conteos de `categorias`, y no existe el `Segmented` «Mostrar»; tocar «Ideas» escribe `?mostrar=en-curso`, pide `/api/workspace?mostrar=en-curso` y las tiras de etapa Idea desaparecen; tocar «En curso» siendo la única activa no cambia nada; tocar «Sin traza» hace aparecer «Sin traza (N)» cerrado; en celular las cuatro filas están en el Sidebar y tocar una no lo cierra; abrir y cerrar el detalle conserva `mostrar`.
- CA-7: `npm test -- grafo` (`web/src/components/detalle/grafo.test.ts`, nuevo): con el fixture de `Detalle.test.tsx` (9 nodos, rama `run-1`), la variante vertical da `x` igual para todos los nodos de `principal`, `y` estrictamente creciente, `x` mayor para `run-1`, una arista `curva` que une el nodo de bifurcación con el primero de `run-1`, y `ancho` igual con 2 y con 9 nodos; la horizontal es `toEqual` a la salida actual.
- CA-8: en `Detalle.test.tsx`, el test de celular (`matchMedia` `matches: false`) comprueba `svg[aria-label="Grafo"]` con `width="100%"`, 9 `[data-grafo-nodo]`, 1 `[data-activo]`, 1 `[data-rama="run-1"]`, y que `svg.parentElement` no tiene la clase `overflow-x-auto`; el test de desktop sigue igual.
- CA-9: `sed -n '/design-system:ajustes:start/,/design-system:ajustes:end/p' web/DESIGN_SYSTEM.md | grep -c -E 'Chips de categoría|Grafo vertical'` = 2, y `git diff sdd/visor-lista-etapas -- web/DESIGN_SYSTEM.md` toca solo líneas entre los markers.
- Escalera: `gofmt -l .`, `go vet ./...`, `npx tsc -b`, `npm run lint`, `npx oxlint -D typescript/no-explicit-any`, `make test`, `cd web && npm test`, `make build`; app levantada contra el fixture: `./bin/sdd --dir "$(bash testdata/mkworkspace.sh)" --no-github --cache-dir "$(mktemp -d)" --port 4322` y `curl -sf 'localhost:4322/api/workspace?mostrar=sin-traza' | jq '[.trabajos[].categoria] | unique'` = `["sin-traza"]`, `curl -sf localhost:4322/api/workspace | jq '.categorias | keys | length'` = 4, `curl -s -o /dev/null -w '%{http_code}' 'localhost:4322/api/workspace?mostrar=otro'` = 400.
- Gates de políticas: los de `.sdd/project.md`, con `origin/sdd/visor-lista-etapas` como base en vez de `origin/main` para `commits` y `max-lineas-archivo`.

Protocolo de prueba humana (aspecto de CA-5 y CA-8, celular):
1. Con el binario de la rama (`make build && ./bin/sdd --dir /home/chiche/workspace --host 0.0.0.0`), abrir el Sidebar en el teléfono: Mostrar con «En curso 6» y «Ideas 19» tildadas, «Terminados 10» y «Sin traza 125» sin tildar (conteos del 2026-10-29). Destildar Ideas: la lista queda con los trabajos en PR y mergeado, la URL dice `?mostrar=en-curso`, el Sidebar sigue abierto.
2. Tildar Terminados: aparecen los publicados «taggeado» y los mergeados cerrados. Tildar Sin traza: aparece «Sin traza (125)» colapsado al final.
3. Destildar todo salvo una: la última no se destilda.
4. Abrir `chichex/skills #100` en el teléfono: grafo vertical, cuatro nodos de arriba hacia abajo con etiqueta, sub-estado y fecha a la derecha, ninguna barra de scroll y ningún nodo cortado; el activo con borde índigo. Abrir un trabajo con run fallado (en `chichex/waica`): la rama cerrada aparece como columna a la derecha con su curva.
5. En desktop: la Barra de filtros muestra «Mostrar» antes de «Etapas», la cabecera no tiene el segmented, y el detalle sigue con el grafo horizontal.

## Riesgos y gaps
- **Revisar primero: inferencia 5** (mínimo una categoría) e **inferencia 8** (chips en la Barra de filtros, no en la cabecera). Cambian cómo se usa el control.
- **Base distinta de `main`**: `/sdd-run` debe correr con `--base sdd/visor-lista-etapas`; el PR se apila sobre #9 y `/sdd-land` del PR nuevo aterriza las ocho capas. El contrato dice «/sdd-run ramifica desde `origin/main`»: el flag lo pisa.
- **Reemplaza CA-1 y CA-3 de `visor-anclas-traza-filtros`** (sin mergear): `TestWorkspace_FiltroActivos`, `describe('filtro Mostrar con traza')` y los tests de `lib.test.ts` sobre `filtro` se reescriben contra `mostrar`, no se relajan ni se borran sin reemplazo.
- **Desviación del grill 2026-10-23 (decisión 16)**: «activos / todos, default activos» deja de existir; lo decidió el usuario el 2026-10-29 tras probar con su workspace. La decisión 23 (Idea = issues abiertos) sigue: las Ideas siguen visibles por default.
- **Contrato desactualizado**: el front ya tiene Vitest pero el contrato no lo registra; por eso CA-4 a CA-8 quedan BAJA. `/sdd-init --update` después del land.
- **`max-lineas-archivo` 300**: `api.ts` (232) suma `categoria` y `categorias`; `filtros.test.tsx` (233) pierde el `describe` de Mostrar y los tests nuevos van a `mostrar.test.tsx` y `grafo.test.ts`; `build.go` (219) suma la categoría; `api_test.go` (194) crece con la tabla nueva: si se acerca al límite, el test de `mostrar` va a un archivo propio.
- **Ideas sin GitHub**: con `--no-github` o GitHub caído no hay snapshot de issues y la categoría `idea` queda en 0; es el comportamiento de hoy (las Ideas salen de GitHub), no un bug nuevo.
- **Fixture**: `testdata/mkworkspace.sh` y las respuestas grabadas de `gh` traen issues abiertos (`idea`), trabajos con eventos y mergeados; el escalón 4 discrimina `sin-traza` solo si el fixture tiene un artefacto sin eventos (la spec anterior encontró que no lo tiene): la evidencia que discrimina CA-1 y CA-2 es el test de Go con su repo temporal.
- Los conteos del protocolo humano (6, 19, 10, 125) son los del workspace del 2026-10-29 y cambian con cada grill o run nuevo.

## Desviaciones
- 2026-10-29 [DEVIATION] Geometría del grafo vertical (inferencia 12, [ASSUMED]): el texto de cada fila arranca a la derecha de la última columna de nodos (`x` = 24 + 40·K + 24, con K = ramas cerradas) en vez de un `x` fijo, para que el codo de una rama cerrada no cruce el texto de la viva; los nodos de una rama cerrada ocupan sus propias filas debajo de la viva, así cada nodo tiene una línea de texto. No cambia el alcance.
- 2026-10-29 [DEVIATION] La etiqueta y el sub-estado del grafo vertical se recortan con «…» según el ancho de su columna (el texto completo queda en el `<title>` del nodo): la spec no preveía el desborde de un sub-estado largo como `fallado (interrupted)`. No cambia el alcance.
- 2026-10-29 [DEVIATION] CA-1, plan de verificación: la invariante «exactamente una categoría por trabajo» vive en `TestBuild_CategoriaUnicaEnLogsReales` (los mismos logs reales de `TestBuild_TrabajosDeLogsReales`, sin y con GitHub) en vez de dentro de ese test, para no mezclarla con la prueba de anclas. No cambia el alcance.

## Resultado de ejecucion (2026-10-29 · HEAD 99011dd)
| CA | Capa | Estado | Evidencia |
|---|---|---|---|
| CA-1 | 1/1 | verificado | `go test -count=1 ./internal/model -run TestBuild_Categoria -v`: 3 tests ok (cuatro grupos idea, sin-traza, terminado y en-curso; la idea no cambia `sinHistoria`; una categoría válida y coherente con las banderas en los 12 trabajos de logs reales y en los 14 con GitHub). Rojo previo: 33 FAIL por `categoria = ""`. |
| CA-2 | 1/1 | verificado | `go test -count=1 ./internal/api -run 'TestWorkspace_Mostrar' -v`: 3 tests ok (11 casos de la tabla: default, `en-curso`, `en-curso,terminado`, `sin-traza`, las cuatro, `?filtro=activos` y `?filtro=todos` ignorados, `?repo=`; 6 consultas inválidas → 400 con el mensaje nuevo). Rojo previo: el server ignoraba `mostrar` y devolvía el filtro viejo. |
| CA-3 | 1/1 | verificado | Mismo comando: `categorias` = `{idea:1, en-curso:1, terminado:1, sin-traza:1}` en todas las respuestas 200, con `?repo=vacio` cuatro ceros. Rojo previo: `categorias = map[]`. |
| CA-4 | 1/1 | verificado | `cd web && npm test -- lib`: 46/46 (`lib.test.ts` 22 con `readFilters`, `filtersQuery`, `trabajoPath` y `writeFilters` sobre `mostrar`, más `filtrar.test.ts` 24). Rojo previo: 11 FAIL en `lib.test.ts`. El runner no está registrado en el contrato (BAJA). |
| CA-5 | 1/1 | verificado (lógica) · pendiente humano (aspecto) | `npm test -- mostrar`: 16/16 (grupo «Mostrar» con cuatro botones en orden y conteos de `categorias`, antes de «Etapas»; sin `Segmented` «Mostrar»; la última activa no se destilda; Sidebar con cuatro filas que no lo cierran). Aspecto: protocolo humano 1 a 3 y 5. |
| CA-6 | 1/1 | verificado | `npm test -- mostrar` y `npm test -- filtrar` (24/24): pide `/api/workspace` (default) o `?mostrar=<lista>`, refetchea al cambiar, «Sin traza (N)» cerrado cuando la categoría está en la respuesta, ideas en línea, badge por `sinHistoria`. Rojo previo: 16 FAIL en `mostrar.test.tsx` (app sin cablear) y 6 en `filtrar.test.ts`. |
| CA-7 | 1/1 | verificado | `npm test -- grafo`: 18/18 (`x` constante e `y` creciente en la viva, cerrada a la derecha con arista curva desde `principal-1` como en horizontal, `ancho` 340 con 2 y con 8 nodos; la horizontal `toEqual` a la salida anterior). Rojo previo: 5 FAIL; la caracterización horizontal pasa antes y después por construcción. |
| CA-8 | 1/1 | verificado (lógica) · pendiente humano (aspecto) | `npm test -- Detalle`: 31/31 (celular: `svg` con `width="100%"`, `viewBox` `0 0 340 N`, contenedor sin `overflow-x-auto`, 9 `[data-grafo-nodo]`, 1 `[data-activo]`, 1 `[data-rama="run-1"]`, tres `<text>` por nodo; desktop `width` 968 y `overflow-x-auto`). Rojo previo: `expected '968' to be '100%'`. Aspecto: protocolo humano 4 y 5. |
| CA-9 | 1/1 | verificado | `sed -n '/design-system:ajustes:start/,/design-system:ajustes:end/p' web/DESIGN_SYSTEM.md \| grep -c -E 'Chips de categoría\|Grafo vertical'` = 2 (antes 0); `git diff sdd/visor-lista-etapas -- web/DESIGN_SYSTEM.md` toca la línea 131 y las 138-142, entre los markers (124 y 143). |
| POL-commits | 1/1 | verificado | `git log --format=%s origin/sdd/visor-lista-etapas..HEAD \| grep -vE '<patrón del contrato>'` vacío (7 commits al medir). |
| POL-max-lineas-archivo | 1/1 | verificado | `git diff --name-only --diff-filter=AM origin/sdd/visor-lista-etapas...HEAD -- '*.go' '*.ts' '*.tsx' \| xargs -r wc -l \| awk '$2!="total" && $1>300'` vacío (el más largo: `web/src/lib/api.ts`, 272). |
| POL-no-any-ts | 1/1 | verificado | `cd web && npx oxlint -D typescript/no-explicit-any`: exit 0. |
| POL-gofmt | 1/1 | verificado | `gofmt -l .`: vacío. |
| POL-no-panic | 1/1 | verificado | `git grep -n 'panic(' -- '*.go' ':!cmd/' ':!*_test.go' ':!web/embed.go'`: vacío. |

Regresión y escalera, todo en HEAD 99011dd: `gofmt -l .` vacío; `go vet ./...` exit 0; `cd web && npx tsc -b` exit 0; `npm run lint` exit 0 (3 warnings, los mismos que corre la base); `make test`: ok en todos los paquetes (63 tests de primer nivel, 58 en la base; 249 con subtests, 201 en la base); `cd web && npm test`: 12 archivos, 176/176 (135 en la base); `make build` exit 0. App levantada (`./bin/sdd --dir <testdata/mkworkspace.sh> --no-github --cache-dir <tmp> --port 4322`, proceso matado): `/api/workspace` trae `categorias` con 4 claves; `?mostrar=otro` → 400 con el mensaje nuevo; `?filtro=todos` se ignora. Con el fixture tal cual, `?mostrar=sin-traza` devuelve `[]` y no `["sin-traza"]`: el fixture no tiene un artefacto sin eventos (lo anticipaban los Riesgos). Probe complementario sobre el mismo fixture generado en un directorio temporal con una spec sin eventos agregada: `?mostrar=sin-traza` → `["sin-traza"]`, `categorias` = `{en-curso:12, idea:0, sin-traza:1, terminado:0}`. Sin verificar: el aspecto de CA-5 y CA-8 y el uso en celular (sin navegador en la sesión); los conteos y el comportamiento contra GitHub y el workspace real.
