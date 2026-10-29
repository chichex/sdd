# Design system — web
<!-- Generado por /design-system el 2026-10-25. Regenerar con /design-system; "Ajustes de este proyecto" se preserva. -->
<!-- design-system: generated=2026-10-25; surface=web; direction=sobrio-indigo -->

## Dirección
- Audiencia y tono: producto de uso diario, sobrio. Un solo usuario (desarrollador) mira el board varias veces por día: jerarquía clara, poco adorno, nada compite con los datos.
- Paleta base: neutros zinc con acento índigo `#4F46E5`. El chrome es neutro; el color lo llevan las etapas y los sub-estados (validado, cambios pedidos, checks rojos).
- Tipografía: sistema (`system-ui`) para texto y mono del sistema para referencias (`#42`, slugs, ramas, SHAs). Sin fuentes externas: el binario sirve todo embebido y funciona sin red.
- Densidad, radios y modo: compacta (base 4 px, filas de 32 px en desktop y 44 px en touch), radios de 6 px, claro y oscuro siguiendo el sistema con toggle manual.

## Tokens
Los tokens se materializan como variables CSS bajo `@theme` de Tailwind 4 en `web/src/index.css`, con los valores de abajo para claro y oscuro. Los componentes usan solo tokens; nunca un hex.

### Color
| Token | Claro | Oscuro | Uso |
|---|---|---|---|
| `--color-bg` | `#FFFFFF` | `#09090B` | fondo de la página |
| `--color-bg-subtle` | `#FAFAFA` | `#18181B` | fondo de carriles y paneles |
| `--color-bg-muted` | `#F4F4F5` | `#27272A` | filas alternas, hover, skeleton |
| `--color-fg` | `#18181B` | `#FAFAFA` | texto principal (17:1 / 19:1) |
| `--color-fg-secondary` | `#52525B` | `#A1A1AA` | texto secundario, labels (7:1 / 8:1) |
| `--color-fg-muted` | `#71717A` | `#71717A` | metadatos y placeholders (4,6:1 / 4,9:1, mínimo AA) |
| `--color-fg-disabled` | `#A1A1AA` | `#52525B` | solo texto deshabilitado, nunca información |
| `--color-border` | `#E4E4E7` | `#27272A` | bordes de carriles, cards, inputs |
| `--color-border-strong` | `#D4D4D8` | `#3F3F46` | separadores con más peso, bordes en hover |
| `--color-primary` | `#4F46E5` | `#4F46E5` | fondo de botón primario, selección |
| `--color-primary-fg` | `#FFFFFF` | `#FFFFFF` | texto sobre primario (6,3:1) |
| `--color-primary-text` | `#4338CA` | `#818CF8` | links, texto y bordes de acento (8:1 / 7,5:1) |
| `--color-primary-subtle` | `#EEF2FF` | `#1E1B4B` | fondo de fila seleccionada, badge de acento |
| `--color-ring` | `#6366F1` | `#818CF8` | anillo de foco |
| `--color-success-text` | `#15803D` | `#4ADE80` | validado, mergeado, en prod (5:1 / 11:1) |
| `--color-success-subtle` | `#F0FDF4` | `#052E16` | fondo del badge de éxito |
| `--color-warning-text` | `#B45309` | `#FBBF24` | cambios pedidos, draft con reviews, aviso de degradación (5,1:1 / 12:1) |
| `--color-warning-subtle` | `#FFFBEB` | `#451A03` | fondo del badge y banner de aviso |
| `--color-error-text` | `#B91C1C` | `#F87171` | checks rojos, run fallido, conflicto (6,5:1 / 7,5:1) |
| `--color-error-subtle` | `#FEF2F2` | `#450A0A` | fondo del badge y banner de error |
| `--color-info-text` | `#0369A1` | `#38BDF8` | abierto, en curso, notas (5,9:1 / 9:1) |
| `--color-info-subtle` | `#F0F9FF` | `#082F49` | fondo del badge de info |
| `--color-stage-idea` | `#D4D4D8` | `#3F3F46` | barra de Gantt y punto de la etapa Idea |
| `--color-stage-grill` | `#C4B5FD` | `#6D28D9` | etapa Grill |
| `--color-stage-spec` | `#A5B4FC` | `#4338CA` | etapa Spec |
| `--color-stage-run` | `#7DD3FC` | `#0369A1` | etapa Run |
| `--color-stage-pr` | `#FCD34D` | `#B45309` | etapa PR |
| `--color-stage-merged` | `#86EFAC` | `#15803D` | etapa Mergeado |
| `--color-stage-published` | `#6EE7B7` | `#047857` | etapa Publicado |

Los colores de etapa son de relleno (barras, puntos, bordes de nodo). El texto nunca va sobre ellos: la etiqueta va al lado, en `--color-fg`. Los sub-estados usan siempre los semánticos con su texto y su fondo `subtle` del mismo par.

### Tipografía
- Familias: `--font-sans: ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif`; `--font-mono: ui-monospace, "SFMono-Regular", Menlo, Consolas, "Liberation Mono", monospace`.
- Escala (tamaño / interlineado): `xs` 12/16 px (metadatos, badges), `sm` 13/18 px (filas del board, celdas), `base` 14/20 px (texto por defecto, inputs, botones), `lg` 16/24 px (título de trabajo en el detalle), `xl` 20/28 px (título de sección), `2xl` 24/32 px (título de página).
- Pesos: 400 texto, 500 labels y botones, 600 títulos y el nombre del trabajo. Nunca 700.
- Referencias (`#42`, `sdd/<slug>`, SHAs, fechas y duraciones) en `--font-mono` a un tamaño menos que el texto que las rodea, con `font-variant-numeric: tabular-nums` para que las columnas del Gantt y las fechas alineen.

### Espaciado
Base 4 px: `0` 0, `1` 4, `2` 8, `3` 12, `4` 16, `5` 20, `6` 24, `8` 32, `10` 40, `12` 48, `16` 64 px.
- Padding interno de card, nodo y celda: `3` (12 px). Separación entre nodos de un carril: `2`. Entre carriles: `4`. Gutter de página: `4` (16 px) en celular y `6` en desktop.
- Altura de fila y control: 32 px en desktop; 44 px de área táctil en touch (`@media (pointer: coarse)`), aunque el control dibuje 32.

### Radios
`--radius-sm` 4 px (badges, chips, barras de Gantt), `--radius-md` 6 px (botones, inputs, nodos, cards; default), `--radius-lg` 8 px (paneles, diálogos, sheet), `--radius-full` 9999 px (avatar, punto de estado, toggle).

### Sombras
- `--shadow-sm`: `0 1px 2px rgb(0 0 0 / 0.05)` para cards y nodos en claro; en oscuro no hay sombra, el borde `--color-border` hace el trabajo.
- `--shadow-md`: `0 4px 12px rgb(0 0 0 / 0.08)` en claro y `0 4px 12px rgb(0 0 0 / 0.4)` en oscuro, para popovers, tooltips y menús.
- `--shadow-lg`: `0 12px 32px rgb(0 0 0 / 0.12)` en claro y `0 12px 32px rgb(0 0 0 / 0.5)` en oscuro, para diálogos y el sheet de detalle en celular.

## Componentes base
Base: shadcn/ui sobre Tailwind 4, con sus tokens mapeados a los de arriba. Los componentes propios del dominio (board, nodo, grafo, Gantt) se dibujan con SVG o CSS propios, siguiendo los mismos tokens.

- **Button**: variantes `primary` (`--color-primary` + `--color-primary-fg`), `secondary` (borde `--color-border-strong`, fondo `--color-bg`), `ghost` (sin borde, hover `--color-bg-muted`), `destructive` (`--color-error-text` como fondo, texto blanco). Tamaños `sm` 28 px e `md` 32 px; en touch, área de 44 px. Radio `md`, peso 500, texto `base`. Con icono a la izquierda, separación `2`.
- **Input**: 32 px, radio `md`, borde `--color-border`, fondo `--color-bg`, placeholder `--color-fg-muted`; en foco, anillo `--color-ring`. Error: borde `--color-error-text` y mensaje `xs` debajo. Se usa en la búsqueda.
- **Select / Combobox**: siempre custom (el Select de shadcn/ui sobre Radix), nunca un `<select>` nativo: el nativo ignora el tema y en iOS abre su propio picker. Trigger con la misma caja que el Input, valor a la izquierda y chevron de 16 px en `--color-fg-secondary`; menú con fondo `--color-bg`, borde `--color-border`, radio `md`, `--shadow-md` y padding 4; opciones de 32 px (44 en touch), radio `sm`, hover `--color-bg-muted`, seleccionada con `--color-primary-subtle` y `--color-primary-text`. Se usa en el filtro de repos.
- **Toggle / Segmented** (activos · todos, claro · oscuro · sistema): 28 px, radio `md`, un solo borde `--color-border`, opción activa con `--color-primary-subtle` y texto `--color-primary-text`, sin borde propio. El selector de tema lleva ícono (sol, luna, auto) y etiqueta; en celular vive en el Sidebar, a ancho completo.
- **Badge** (sub-estado): texto `xs` peso 500, padding `1`/`2`, radio `sm`, par semántico `*-text` sobre `*-subtle`. Variantes: `neutral` (draft, pausado, sin spec), `info` (abierto, en curso), `success` (validado, mergeado, en prod), `warning` (cambios pedidos), `error` (checks rojos, fallido, conflicto). Siempre con texto; el color nunca va solo.
- **Chip de referencia**: `--font-mono` `xs`, fondo `--color-bg-muted`, radio `sm`; para `#42`, `owner/repo`, ramas y SHAs. Clickeable abre GitHub en otra pestaña.
- **Card**: fondo `--color-bg`, borde `--color-border`, radio `md`, `--shadow-sm`, padding `3`.
- **Nodo** (unidad del board): card de 12 px de padding con un solo borde `--color-border` y tres filas que nunca se apilan: título `sm` 600 en una línea con elipsis, precedido por un punto de 8 px del color de su etapa; chip de referencia a la izquierda y antigüedad corta en mono a la derecha; badge de sub-estado y, si hay, chip del PR. Lo que no entra se trunca con elipsis, nunca baja de línea. Sin franjas ni bordes extra. Estados `default`, `hover` (borde `--color-border-strong`), `selected` (borde `--color-primary-text` y fondo `--color-primary-subtle`; el anillo queda solo para el foco de teclado), `stale` (texto `--color-fg-muted`).
- **Carril** (columna de etapa): cabecera `xs` 600 en mayúsculas con el nombre de la etapa, el conteo y un punto del color de etapa; cuerpo con fondo `--color-bg-subtle`, radio `lg`, padding `2`. Siete carriles de 240 px en desktop con scroll horizontal del contenedor; por debajo de 768 px el board no se muestra como carriles sino como la Lista de trabajos.
- **Tira de trabajo**: fila que atraviesa los carriles con los nodos del mismo trabajo; las ramas van una debajo de la otra dentro de la tira, unidas por conectores de 1 px en `--color-border-strong`.
- **Lista de trabajos** (celular, y desktop con el filtro «todos»): una fila de 44 px por trabajo con punto de etapa, título con elipsis, badge de sub-estado y antigüedad corta en mono, todo en una sola línea; borde inferior `--color-border`; tocar abre el detalle como sheet. Nada baja de línea.
- **Grafo** (detalle): SVG propio. Nodos de 32 px, aristas de 1,5 px en `--color-border-strong`, bifurcaciones con curva; el nodo activo usa `--color-primary-text`.
- **Gantt** (detalle): SVG propio. Barras de 20 px de alto con radio `sm` rellenas con el color de la etapa, etiqueta fuera de la barra en `--color-fg` `sm`, eje de fechas en mono `xs`, línea de hoy en `--color-primary-text` discontinua. Scroll horizontal dentro del contenedor.
- **Table / List**: filas de 32 px (44 en touch), borde inferior `--color-border`, hover `--color-bg-muted`, cabecera `xs` 600 `--color-fg-secondary`.
- **Tabs**: subrayado de 2 px en `--color-primary-text`, texto `base` 500; inactivas en `--color-fg-secondary`.
- **Dialog / Sheet**: fondo `--color-bg`, radio `lg`, `--shadow-lg`, overlay `rgb(0 0 0 / 0.5)`. En celular el detalle abre como sheet desde abajo con respeto de safe-area; el Sidebar usa el mismo overlay y la misma sombra.
- **Tooltip / Popover**: fondo `--color-fg` con texto `--color-bg` para el tooltip (`xs`); popover con `--color-bg`, borde y `--shadow-md`.
- **Toast**: esquina inferior derecha (inferior centrada en celular), `--color-bg`, borde, `--shadow-md`, icono semántico a la izquierda; se cierra solo a los 4 s salvo los de error.
- **Banner**: barra de ancho completo bajo la cabecera, par semántico `warning` o `error`, con texto de hasta dos líneas y una acción (reintentar) que queda siempre a la derecha, nunca debajo. Lo usa el modo degradado sin GitHub.
- **Navegación** (cabecera): 48 px, fondo `--color-bg`, borde inferior. En desktop (≥768 px) una sola fila: nombre de la app, Select de repo, segmented activos · todos, Refresh, toggle de tema. En celular la cabecera tiene solo tres piezas: botón hamburguesa (ícono `menu`, 44 px, `aria-label` «Menú») a la izquierda, nombre de la app seguido del repo activo en una línea con elipsis, y el Refresh a la derecha. El Select, el filtro y el tema se mudan al Sidebar; nada baja a una segunda fila.
- **Sidebar** (menú lateral, solo celular): panel de 280 px (máximo 85 % del ancho) que entra desde la izquierda sobre el overlay `rgb(0 0 0 / 0.5)`; fondo `--color-bg`, borde derecho `--color-border`, `--shadow-lg`, respeta `safe-area-inset-top`, `-bottom` y `-left`. Cabecera con el nombre de la app y botón de cerrar (ícono `x`, 44 px). Tres secciones con título `xs` 600 en mayúsculas `--color-fg-secondary`: **Repo** (una fila de 44 px por repo, tilde a la izquierda y conteo de trabajos en mono a la derecha; la activa con `--color-primary-subtle` y `--color-primary-text`), **Mostrar** (activos · todos como filas con tilde) y **Tema** (segmented a ancho completo con íconos y etiquetas); al pie, versión del binario y carpeta escaneada en mono `xs`. Es un `role="dialog"` `aria-modal` con el foco atrapado; cierra con la X, Esc, tocando el overlay o al elegir repo o filtro, y devuelve el foco a la hamburguesa. El body no scrollea mientras está abierto. Entra con una transición de 150 ms, sin transición con `prefers-reduced-motion`. En desktop no existe: sus controles van en la cabecera.
- **Refresh** (cabecera): Button `ghost` de 32 px (44 en touch) sin borde, con el ícono `refresh-cw` de 16 px y la antigüedad corta en mono `xs`, los dos en `--color-fg-secondary`; hover con `--color-bg-muted` y texto `--color-fg`. La antigüedad es la etiqueta del botón: «2 min» significa «actualizado hace 2 min, tocá para refrescar», y el tooltip trae la hora absoluta. Mientras refresca, el ícono gira y la etiqueta muestra «…» (sin giro con `prefers-reduced-motion`, solo opacidad 0,5); el board no se vacía. Si la última actualización falló, ícono y etiqueta pasan a `--color-error-text` y el Banner `error` explica. Nunca lleva borde ni fondo propio: es una acción secundaria y no compite con el Select ni con el segmented.
- **Skeleton**: bloques `--color-bg-muted` con radio `sm` y pulso suave; respeta `prefers-reduced-motion`.
- **EmptyState**: ícono en `--color-fg-muted`, texto `sm` y, si aplica, un comando para copiar en mono.

## Estados y patrones
- **Carga**: la primera carga muestra skeletons con la forma del board (siete carriles, tres nodos cada uno); un refresh en curso solo hace girar el ícono del Refresh y no vacía el board. Lo resuelven Skeleton y Refresh.
- **Vacío**: un carril sin trabajos muestra un EmptyState corto («Sin trabajos en Spec»); un workspace sin repos muestra un EmptyState de página con el comando para correr el binario en otra carpeta.
- **Error**: fallo de lectura local o de GitHub → Banner `error` con reintento, el board conserva el último dato bueno y el Refresh muestra su antigüedad en `error`. Un nodo con `run.failed` o checks rojos lleva Badge `error`.
- **Degradado**: GitHub inaccesible → Banner `warning` persistente «Sin GitHub: mostrando solo lo local», los nodos que dependen de GitHub quedan en `stale`.
- **Éxito**: refresh manual terminado → Toast `success` de una línea; validado, mergeado y en prod → Badge `success`.
- **Deshabilitado**: opacidad 0,5, cursor `not-allowed`, texto `--color-fg-disabled`; nunca solo color: el control conserva su etiqueta.
- **Foco**: anillo de 2 px en `--color-ring` con `outline-offset` de 2 px, solo con `:focus-visible`. Todo lo clickeable se alcanza con Tab; el board se recorre con flechas y Enter abre el detalle.
- **Selección**: nodo seleccionado con `--color-primary-subtle` y anillo; en el detalle el nodo activo del grafo usa `--color-primary-text`.
- **Antigüedad**: en nodos, listas y cabecera la fecha va relativa y corta («3 h», «20 min», «1 d») en mono, en una sola línea; «hace 3 h» solo en toasts y tooltips, junto con la fecha absoluta.

## Reglas
- **MUST** Usar solo tokens de este doc (variables CSS bajo `@theme`); ningún hex ni color de Tailwind directo en componentes. Porqué: claro y oscuro salen de un solo lugar y el gate de hex fuera de tokens se puede medir.
- **MUST** Cumplir contraste AA (4,5:1 texto, 3:1 texto grande y bordes de control) en los dos modos. Porqué: decisión E32 del grill; el board se usa en celular a plena luz.
- **MUST** No comunicar un estado solo con color: badge con texto, icono o patrón. Porqué: daltonismo y modo oscuro cambian la percepción del color.
- **MUST** Soportar claro y oscuro con `prefers-color-scheme` como default y `data-theme` para el toggle manual; `body` con fondo explícito. Porqué: el sistema del celular y el del desktop pueden diferir.
- **MUST** Mantener 44 px de área táctil en `pointer: coarse` aunque la fila dibuje 32 px, respetar `safe-area-inset-*` y evitar el zoom al enfocar inputs (`font-size` ≥ 16 px en inputs en celular). Porqué: se abre por LAN desde el teléfono.
- **MUST** No producir scroll horizontal de página: el scroll horizontal vive solo dentro del board, el grafo y el Gantt. Porqué: decisión E33 del grill; en celular el board se vuelve lista.
- **MUST** Todo lo accionable es alcanzable con teclado y tiene foco visible con `:focus-visible`. Porqué: decisión E32 del grill.
- **MUST** En celular nada se apila: las filas de nodo, lista, cabecera y banner son de una sola línea, y lo que no entra se trunca con elipsis, pasa a ícono o se muda al Sidebar. Porqué: el board se lee de un vistazo en 390 px de ancho y un elemento que baja de línea rompe la fila.
- **MUST** Un solo borde por elemento, de 1 px en `--color-border`: nunca borde más franja, ni borde más anillo salvo el foco de teclado; la selección se marca con color de borde y de fondo. Porqué: los bordes dobles y las franjas con radio se ven sucios, sobre todo en pantallas chicas.
- **MUST** Ningún control nativo con estilo del sistema: Select, checkbox, switch y date picker son componentes del kit con los tokens del doc. Porqué: el nativo ignora el tema claro y oscuro y cambia entre plataformas.
- **MUST** No cargar fuentes ni assets externos. Porqué: el binario se sirve embebido y tiene que funcionar sin red (decisión E29).
- **SHOULD** Respetar `prefers-reduced-motion`: sin pulsos ni transiciones mayores a 150 ms cuando está activo. Porqué: accesibilidad y costo cero.
- **SHOULD** Mostrar fechas relativas con la absoluta en tooltip, en mono y con números tabulares. Porqué: el board se lee de un vistazo y el Gantt necesita alineación.
- **SHOULD** Preferir un componente de shadcn/ui antes que uno propio, salvo board, nodo, grafo y Gantt. Porqué: decisión E30 del grill; menos superficie que mantener.

## Ajustes de este proyecto
<!-- design-system:ajustes:start -->
Piezas del visor sumadas por la spec `visor-anclas-traza-filtros` (2026-10-28). Usan solo los tokens de este doc.

- **Barra de filtros**: barra bajo la cabecera, 40 px en desktop, fondo `--color-bg`, borde inferior `--color-border`, gutter `6`. Lleva el Input de búsqueda (32 px, 280 px de ancho, ícono `search` de 16 px en `--color-fg-muted` y botón `ghost` de limpiar con ícono `x` y `aria-label` «Limpiar búsqueda») y las chips de las siete etapas. En celular (<768 px) solo el Input, a ancho completo, de 44 px y `font-size` 16 px; las etapas se mudan al Sidebar.
- **Chip de etapa**: botón de 28 px, radio `md`, un solo borde `--color-border`, texto `xs` 500 con el nombre de la etapa y el conteo en `--font-mono`; activa con `--color-primary-subtle` y `--color-primary-text`, con `aria-pressed` y multi-selección. En el Sidebar, sección Etapas, es una fila de 44 px con punto de etapa, nombre, conteo y tilde que no cierra el menú.
- **Badge sin traza**: `Badge` `neutral` (`--color-fg-secondary` sobre `--color-bg-muted`) con el texto literal «sin traza», para trabajos sin eventos en `events.jsonl`. Va en la tercera fila del nodo después del sub-estado, en la fila de la lista y junto a los chips del detalle, que además suma una nota `sm` en `--color-fg-muted`.
- **Encabezado de repo**: `h3` con `owner/repo` en `--font-mono` `xs` 600 mayúsculas `--color-fg-secondary`, conteo a la derecha en `--color-fg-muted`, fondo `--color-bg-subtle`, radio `sm`, 28 px de alto. En el board cruza los siete carriles; en la lista es `sticky` arriba. Se omite si la respuesta trae un solo repo.
- **Grupo colapsado**: «Sin traza (N)» al final del board y de la lista, solo con el filtro «todos»: `Button` `ghost` de 32 px con chevron de 16 px (`chevron-right` cerrado, `chevron-down` abierto) y `aria-expanded`. Cerrado por default; con N = 0 no se dibuja. Abierto, muestra sus trabajos agrupados por repo.
- **Ícono de ancla**: ícono lucide de 14 px en `--color-fg-muted` según el tipo (`circle-dot` issue, `messages-square` grill, `file-text` spec), dentro del chip de referencia del nodo y del chip del ancla del detalle (fuera del link): `role="img"` con `aria-label` «ancla: <tipo>» y Tooltip con el mismo texto. En la fila de la lista va acompañado de su palabra (`issue`, `grill` o `spec`), `aria-hidden` y sin Tooltip.

Piezas sumadas por la spec `visor-lista-etapas` (2026-10-29), solo para la Lista de trabajos de celular. Reemplazan el punto de etapa de la fila de la Lista de trabajos y el ícono suelto de ancla.

- **Encabezado de etapa**: subtítulo de sección dentro del grupo de repo, un `h4` por etapa presente, de la más avanzada a la menos. Mismo dibujo que la cabecera del carril: punto de etapa (`--color-stage-*`), nombre en `xs` 600 mayúsculas y texto `--color-fg`, y conteo a la derecha en `--font-mono` `xs` `--color-fg-muted` con números tabulares; sin fondo, 28 px de alto, gutter `4`. Su nombre accesible es «<Etapa> <conteo>». Se dibuja aunque el repo tenga una sola etapa y también dentro de «Sin traza (N)». No es focusable ni accionable: no entra al recorrido con Tab.
- **Etiqueta de ancla**: ícono lucide de 14 px `aria-hidden` seguido de la palabra `issue`, `grill` o `spec` en `--font-mono` `xs` `--color-fg-muted`, separados por `1`; no se encoge. Abre la fila de la lista, que sigue siendo una línea de 44 px: etiqueta, título con elipsis, badge «sin traza» si aplica, badge de sub-estado y antigüedad en mono. Lo único que se trunca es el título. La palabra es texto visible (en touch no hay hover) y abre el nombre accesible de la fila.
<!-- design-system:ajustes:end -->
