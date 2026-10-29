# Spec — Visor: lista por etapas en celular
<!-- Generada por /sdd-spec el 2026-10-29. Fuente: pedido libre. Estado: aprobada -->
<!-- SDD-Tracking: version=1; type=spec; state=approved; issue=none; grill=none; superseded-by=none -->

## Contexto
Probado en el teléfono contra `/home/chiche/workspace` (2026-10-29, binario de PR #8, `3368540`), la lista de celular no dice qué es cada fila: la etapa se comunica solo con el punto de 8 px de color (`EtapaDot`, `aria-hidden`, `web/src/components/marcas.tsx:12`, usado en `web/src/components/ListaTrabajos.tsx:28`) y el tipo de ancla solo con un ícono cuyo texto vive en un Tooltip (`marcas.tsx:30-44`), que en touch no abre. En desktop la etapa se lee por el carril (`web/src/components/board/Board.tsx:57-59`); en la lista el carril desaparece y el color queda solo, contra la regla «no comunicar un estado solo con color» de `web/DESIGN_SYSTEM.md`. Además, un trabajo anclado en una spec que vive solo en una rama (la de #8) muestra la ruta `.sdd/specs/…md` como título, porque `title()` cae a la ref del ancla (`internal/model/build.go:109-120`). La lista ya viene ordenada por repo y, dentro del repo, por etapa de más avanzada a menos (`web/src/lib/filtrar.ts` `agrupar`), así que partirla en secciones por etapa es consecutivo y barato. Esta spec sigue a `visor-anclas-traza-filtros`: su base es la rama `sdd/visor-anclas-traza-filtros` (PR #8), no `main`, y el PR resultante se apila sobre #8.

## Diseño
Superficie `web` (`web/src`), con sistema: tokens en `web/src/index.css` (`@theme`), componentes en `web/src/components/ui/` (badge, button, dialog, input, select, sheet, skeleton, sonner, tooltip) y doc `web/DESIGN_SYSTEM.md` (dirección `sobrio-indigo`, con «Ajustes de este proyecto» de la spec anterior). Solo cambia la Lista de trabajos (celular, <768 px); el board, los nodos y el detalle no se tocan.

- Reusa: `EtapaDot`, `ETAPA_INFO` y `etapaInfo` (`web/src/components/etapas.ts`), `EncabezadoRepo` y `GrupoSinTraza` (`web/src/components/Grupos.tsx`), `Badge`, `SubEstadoBadge`, `SinTrazaBadge`, íconos lucide `circle-dot`, `messages-square`, `file-text`, tokens `--color-fg-secondary`, `--color-fg-muted`, `--font-mono`.
- Nuevo: **Encabezado de etapa** (subtítulo de sección dentro del grupo de repo, mismo dibujo que la cabecera del carril: punto de etapa, nombre en `xs` 600 mayúsculas y conteo en mono `--color-fg-muted`; sin fondo; `h4`; 28 px) y **Etiqueta de ancla** (ícono de 14 px `aria-hidden` seguido de la palabra `issue`, `grill` o `spec` en `--font-mono` `xs` `--color-fg-muted`).
- Fila de la lista: pierde el punto de etapa (lo da el subtítulo) y cambia el ícono solo por la Etiqueta de ancla. Sigue siendo una línea de 44 px: etiqueta de ancla, título con elipsis, badge «sin traza» si aplica, badge de sub-estado y antigüedad en mono. Lo único que se trunca es el título.

Referencia a igualar (mock elegido por el usuario el 2026-10-29, 390 px):
```text
CHICHEX/WAICA                                     14   ← Encabezado de repo (h3, sticky)
 ● PR                                              1   ← Encabezado de etapa (h4)
 ⊙ issue  Renaming an Animator blend i…  validado  58 min
 ● IDEA                                           13
 ⊙ issue  Roadmap: grill order for the…  abierto   14 d
 ⊙ issue  Unify the duplicated glb pars…  abierto   14 d

CHICHEX/SDD                                        2
 ● PR                                              2
 ▤ spec   visor-anclas-traza-filtros     validado  13 h
 ▭ grill  Visor del flujo SDD            validado  2 d
```

## Comportamiento esperado
- **CA-1** Función pura `seccionarPorEtapa(trabajos)` en `web/src/lib/filtrar.ts`: recibe los trabajos de un grupo (ya ordenados por `agrupar`) y devuelve secciones `{ etapa, trabajos }` consecutivas, una por etapa presente, en el mismo orden de entrada y sin reordenar ni perder trabajos; con lista vacía devuelve `[]`. — BAJA (ALTA con el runner; ver Verificabilidad)
- **CA-2** En `ListaTrabajos`, dentro de cada grupo de repo (abierto o dentro de «Sin traza (N)» abierto) las filas se dibujan por sección de CA-1, cada una precedida por un Encabezado de etapa: `h4` cuyo nombre accesible es `<Nombre de ETAPA_INFO> <conteo de la sección>` (p. ej. «PR 1», «Idea 13»), con `EtapaDot`. Se dibuja aunque el repo tenga una sola etapa; con el filtro de etapas activo solo aparecen las secciones de las etapas presentes. El board (`Board.tsx`) no cambia. — BAJA (lógica) · NULA (aspecto)
- **CA-3** La fila de la lista ya no lleva `EtapaDot` y, en lugar del ícono solo, lleva la Etiqueta de ancla: ícono `aria-hidden` más la palabra literal `issue`, `grill` o `spec` según `ancla.tipo`, visible como texto. El nombre accesible del botón de la fila empieza con esa palabra seguida del título. El chip de referencia del nodo y los chips del detalle conservan `AnclaIcono` con su `aria-label` y Tooltip (CA-7 de la spec anterior). — BAJA (lógica) · NULA (aspecto)
- **CA-4** Sin título de issue ni H1 de spec o grill, `Titulo` es el slug del ancla en vez de la ref cruda (`internal/model/build.go` `title()`): para una ref de spec `.sdd/specs/<slug>.md` → `<slug>`; para una ref de grill `<YYYY-MM-DD>-<slug>` → `<slug>`; para issues y cualquier otra forma, la ref tal cual. `Ancla.Ref` y el `id` del trabajo no cambian; la búsqueda por `ancla.ref` sigue encontrando la ruta. — ALTA
- **CA-5** En 390 px de ancho la fila sigue siendo una sola línea de 44 px: la etiqueta de ancla, los badges y la antigüedad no se truncan ni bajan de línea; solo el título se corta con elipsis. El Encabezado de etapa no es focusable ni accionable y no se cuenta como control en el recorrido con Tab (CA-12 de la spec anterior sigue valiendo). — NULA (celular)
- **CA-6** `web/DESIGN_SYSTEM.md`, dentro de los markers de «Ajustes de este proyecto», documenta las dos piezas nuevas («Encabezado de etapa» y «Etiqueta de ancla») con sus tokens, y la entrada «Ícono de ancla» dice que en la fila de la lista va acompañado de su palabra. Nada fuera de los markers cambia. — ALTA

## Plan de entrega
| Capa | Etapa | CAs | Justificacion |
|---|---|---|---|
| 1 | lista-etapas | CA-1 a CA-6 | Un solo PR apilado sobre #8 (`--base sdd/visor-anclas-traza-filtros`): tres archivos de front, una función de Go y el doc. El cambio de Go (CA-4) es independiente pero chico; partirlo en dos PRs no agrega verificación. |

## Fuera de alcance
- Cambiar el board de desktop, el `NodoCard`, el `StackCard` o el detalle (siguen con el ícono de ancla y el carril como contexto).
- Etiqueta de etapa por fila (opción B, descartada por el usuario).
- Leer la spec desde la rama del PR para resolver el título real cuando el archivo no está en `main` (el slug es el fallback; el título real aparece al aterrizar).
- Colapsar o filtrar desde los Encabezados de etapa (las chips de etapa del Sidebar ya filtran).
- Cambios de API, de `events.jsonl` o del modelo salvo `title()`.

## Inferencias
| # | Inferencia | Eleccion propuesta | Alternativa razonable | Confianza | Resolucion |
|---|---|---|---|---|---|
| 1 | ¿Cómo se marca la etapa en la lista? | Subtítulos por etapa dentro de cada repo (carriles en vertical) | Etiqueta punto + nombre en cada fila | alta | elegida por usuario: A (2026-10-29) |
| 2 | ¿Se dibuja el subtítulo si el repo tiene una sola etapa? | Sí, siempre: la lectura es uniforme | Omitirlo con una sola sección | media | [ASSUMED] |
| 3 | ¿Estilo y nivel del subtítulo? | Igual a la cabecera del carril (punto, nombre `xs` 600 mayúsculas, conteo mono), sin fondo, `h4` bajo el `h3` del repo, 28 px | Fondo `--color-bg-subtle` como el de repo | media | [ASSUMED] |
| 4 | ¿La fila conserva el punto de etapa? | No: lo da el subtítulo y libera 16 px para el título | Conservarlo | alta | [ASSUMED] |
| 5 | ¿Cómo se muestra el ancla en la fila? | Ícono + palabra `issue`/`grill`/`spec`, mono `xs` `--color-fg-muted` | Ícono + referencia (`#245`, slug) | media | [ASSUMED] |
| 6 | ¿El nodo del board y el detalle cambian? | No: desktop tiene hover y el carril da la etapa | Sumar la palabra también al chip del nodo | alta | [ASSUMED] |
| 7 | ¿Qué slug usa el fallback del título? | Spec: nombre de archivo sin `.sdd/specs/` ni `.md`; grill: sin el prefijo de fecha; resto igual | Mostrar la ruta completa, como hoy | media | [ASSUMED] |
| 8 | ¿El conteo del subtítulo qué cuenta? | Trabajos de esa etapa en ese repo tras etapa y búsqueda (los de la sección) | Conteo global de la etapa | alta | [ASSUMED] |
| 9 | ¿Secciones también dentro de «Sin traza (N)»? | Sí, mismo componente | Lista plana dentro del grupo | alta | [ASSUMED] |
| 10 | ¿El Tooltip del ancla sigue en la fila? | No: la palabra es texto visible; el Tooltip queda en nodo y detalle | Mantenerlo | media | [ASSUMED] |
| 11 | ¿Entrega? | Una capa, PR apilado sobre #8 | Corregir sobre la rama de #8 | alta | [ASSUMED] |

## Verificabilidad
Mixto, anclado en el contrato del 2026-10-25:
- **ALTA: CA-4 y CA-6.** `go test -count=1 ./internal/model` (escalón 2) y greps sobre `web/DESIGN_SYSTEM.md`.
- **BAJA: CA-1, CA-2 y CA-3.** El contrato sigue sin registrar un runner de front (Gaps); por contrato solo llegan `npx tsc -b`, `npm run lint` y el build. La rama base trae Vitest + Testing Library con 120 tests (`cd web && npm test`), que el run usa igual; serían ALTA cuando `/sdd-init --update` registre `npm test`.
- **NULA: CA-5 y el aspecto de CA-2 y CA-3.** «La UI renderizada… hace falta un navegador» (Verificacion autonoma); sin Playwright ni MCP de navegador. Protocolo humano abajo.
- Políticas: `commits`, `max-lineas-archivo` (los archivos tocados tienen entre 32 y 224 líneas), `no-any-ts`, `gofmt` y `no-panic` aplican como gates. Sin dependencias nuevas.

## Plan de verificacion
Mecanismo elegido: test de Go para `title()`; Vitest + Testing Library por rol y texto para la lista (función pura en `filtrar.ts` y `render` con `matchMedia` en `matches: false`, como en `web/src/App.test.tsx:90`); greps para el doc; protocolo humano para aspecto y celular. Alternativas: (más exhaustiva) Playwright con captura de la lista a 390 px contra el mock; (más barata) solo `seccionarPorEtapa` puro y protocolo humano.

Por CA:
- CA-1: `cd web && npm test -- lib`: `seccionarPorEtapa([pr, pr, idea])` → `[{etapa:'pr', trabajos:[pr, pr]}, {etapa:'idea', trabajos:[idea]}]`; `[]` → `[]`; con una sola etapa, una sección.
- CA-2: `npm test -- App` (o un `Lista.test.tsx` nuevo): fixture con un repo con trabajos en dos etapas; dentro de `getByRole('list', { name: 'Trabajos' })`, `getAllByRole('heading', { level: 4 })` da los nombres en orden («PR 1», «Idea 2») y cada fila está bajo su encabezado; con `?etapa=idea` solo queda el `h4` «Idea 2»; con «Todos» y el grupo «Sin traza» abierto hay `h4` dentro del grupo; el board (`matchMedia` en `matches: true`) no tiene `h4` nuevos.
- CA-3: en el mismo test, `getByRole('button', { name: /^issue Renaming/ })` existe y `within(fila).getByText('issue')` es visible; `queryByLabelText('ancla: issue')` dentro de la lista es `null`, y en `Detalle.test.tsx` sigue existiendo en los chips del detalle.
- CA-4: `go test -count=1 ./internal/model -run TestBuild_TituloSlug`: un grupo anclado en `.sdd/specs/foo-bar.md` sin archivo en disco (solo eventos) → `Titulo == "foo-bar"`; un grill `2026-10-01-norte` sin H1 → `"norte"`; un issue sin título → `"#9"`. `TestBuild_TrabajosDeLogsReales` y `TestBuild_SinHistoria` siguen verdes.
- CA-5: humano.
- CA-6: `sed -n '/design-system:ajustes:start/,/design-system:ajustes:end/p' web/DESIGN_SYSTEM.md | grep -c -E 'Encabezado de etapa|Etiqueta de ancla'` = 2, y `git diff sdd/visor-anclas-traza-filtros -- web/DESIGN_SYSTEM.md` toca solo líneas entre los markers.
- Escalera: `gofmt -l .`, `go vet ./...`, `npx tsc -b`, `npm run lint`, `npx oxlint -D typescript/no-explicit-any`, `make test`, `cd web && npm test`, `make build`; app levantada contra el fixture: `./bin/sdd --dir "$(bash testdata/mkworkspace.sh)" --no-github --cache-dir "$(mktemp -d)"` y `curl -sf localhost:4321/api/workspace | jq '[.trabajos[].titulo | select(startswith(".sdd/"))] | length'` = 0.
- Gates de políticas: los de `.sdd/project.md`, con `origin/sdd/visor-anclas-traza-filtros` como base en vez de `origin/main` para `commits` y `max-lineas-archivo`.

Protocolo de prueba humana (CA-5 y aspecto de CA-2 y CA-3):
1. Con el binario de la rama (`make build && ./bin/sdd --dir /home/chiche/workspace --host 0.0.0.0`), abrir la lista en el teléfono con el filtro default: bajo `CHICHEX/WAICA` aparece «● PR 1» y debajo la fila «⊙ issue Renaming an Animator…», después «● IDEA 13» con sus filas; bajo `CHICHEX/SDD`, «● PR 2» con las filas «▤ spec visor-anclas-traza-filtros» y «▭ grill Visor del flujo SDD».
2. Ninguna fila baja de línea en 390 px; el título se corta con elipsis y los badges y la antigüedad quedan enteros a la derecha.
3. En el Sidebar, activar solo la etapa Idea: cada repo muestra una sola sección «● IDEA n»; limpiar el filtro las devuelve.
4. Elegir «Todos» y abrir «Sin traza (N)»: los repos de adentro también traen sus subtítulos por etapa.
5. En desktop, el board no cambió: siete carriles, nodos con su ícono de ancla y tooltip «ancla: issue» al pasar el mouse.

## Riesgos y gaps
- **Revisar primero: inferencia 5** (ícono + palabra del ancla, no la referencia) e **inferencia 7** (slug del fallback). Definen qué se lee en cada fila.
- **Base distinta de `main`**: `/sdd-run` debe correr con `--base sdd/visor-anclas-traza-filtros`; el PR se apila sobre #8 y `/sdd-land` del PR nuevo aterriza las siete capas. El contrato dice «/sdd-run ramifica desde `origin/main`»: el flag lo pisa.
- **Nombre accesible de la fila**: el botón de la fila gana la palabra del ancla como prefijo. Los tests de la base que busquen filas por `name` exacto deben ajustarse al nombre nuevo, nunca relajarse a `queryAllByRole('button')` sin nombre.
- **Contrato desactualizado**: el front ya tiene Vitest pero el contrato no lo registra; por eso CA-1 a CA-3 quedan BAJA. `/sdd-init --update` después del land.
- **`max-lineas-archivo` 300**: `web/src/App.test.tsx` tiene 174 y `internal/model/build_test.go` 224; si el test de CA-2 no entra cómodo en `App.test.tsx`, va a un `Lista.test.tsx` propio.
- **Fixture**: `testdata/mkworkspace.sh` no tiene un trabajo cuyo título caiga a la ref (sus specs están en disco), así que la prueba del escalón 4 pasa sin discriminar; la evidencia de CA-4 es el test de Go con su repo temporal.
- Con «Todos» y nueve repos, los subtítulos se multiplican (hasta siete por repo). Aceptado por el usuario al elegir A; si molesta, el seguimiento es colapsar por repo, fuera de esta spec.
