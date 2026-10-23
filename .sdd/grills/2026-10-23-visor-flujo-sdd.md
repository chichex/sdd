# Grill — Visor del flujo SDD
<!-- Estado: finalized. Proyecto: /home/chiche/workspace. Fuente: pedido libre. -->
<!-- SDD-Tracking: version=1; type=grill; state=finalized; issue=none; grill=2026-10-23-visor-flujo-sdd; project=%2Fhome%2Fchiche%2Fworkspace -->

## Hechos comprobados
- Los skills escriben un log local append-only, `<git-common-dir>/sdd/events.jsonl`, con 11 eventos (`grill.paused` … `land.merged`). Es por clon y por máquina; los worktrees lo comparten. [verificado: `skills/docs/sdd-events-v1.md`]
- El catálogo de `events-v1` es cerrado e inmutable: un evento nuevo exige `events-v2`. [verificado: mismo doc, "Version compatibility"]
- Hay datos reales en cuatro clones: `dale-que-sale` 17 eventos, `waica` 15, `platform` 5, `skills` 4. Incluyen stacks de hasta 5 PRs, un run `interrupted` reintentado con los mismos PRs, y un `land.merged` por PR. Harnesses en uso: antigravity, claude, opencode. [verificado: `events.jsonl` de cada clon]
- Los eventos no traen título ni estado de review. El mismo issue aparece como `#288` y `pramaestudio/dale-que-sale#288` en un mismo log. [verificado: `dale-que-sale/.git/sdd/events.jsonl`]
- Grill: nace de pedido libre o issue; `paused|finalized`; vive en `<git-common-dir>/sdd/grills/` o `.sdd/grills/`; nunca va a GitHub. Spec: nace de pedido libre, issue o grill; `draft → approved → implemented`, o `superseded`; vive en `<git-common-dir>/sdd/specs/`, `.sdd/specs/` o el body del issue; con dos copias manda la del issue. [verificado: `skills/claude/grill/SKILL.md`, `skills/claude/sdd-spec/SKILL.md`]
- Run: worktree y branch `sdd/<slug>` (o stack `sdd/<slug>/<n>-<etapa>`); importa grill y spec al primer commit; el PR de abajo lleva `Refs #NN`, el top `Closes #NN`; la spec pasa a `implemented` antes del merge. [verificado: `skills/claude/sdd-run/SKILL.md`]
- Gate de merge de `sdd-land`: PR abierto, no draft, `mergeable` no `CONFLICTING`, `reviewDecision` distinto de `CHANGES_REQUESTED`, checks verdes deduplicados por identidad. [verificado: `skills/claude/sdd-land/SKILL.md` Fase 3]
- Linaje machine-readable: `issue=`, `grill=`, `superseded-by=` en markers `SDD-Tracking v1`; parser canónico `parseSddArtifact` y `parseSddEventLine` en `skills/pi-extensions/sdd-artifacts/` (TypeScript). [verificado]
- Prod en repos prama: tag liviano a mano por servicio (`api/v1.0.57` sobre el merge del PR #365), `publish.yml` buildea la imagen, y `deploy.yml` del repo infra se dispara a mano y commitea `versions.env` como bot `prama-deploy`. `prama.yaml` del infra declara `repos.code` y `tag_prefix`. `waica` publica en npm por tag `v*`; `skills` usa releases. Nadie usa environments ni deployments de GitHub. [verificado: workflows y `git for-each-ref`]
- `platform` ya tiene un skill de deploy propio (`/deploy`, `/skill:platform-deploy`) con gates de publish y deploy. [verificado: `platform/.claude/skills/deploy/SKILL.md`]
- Escaneo de `/home/chiche/workspace`: 12 `.sdd/project.md`; dos son worktrees (`platform`, `waica`) y dos son carpetas sin `.git` (`platform-protocolo4`, `platform-sdd-tajada-2-vm`). `chichex/ideas` tiene `.sdd/` pero su remote no existe. [verificado: `find` + `git rev-parse`]
- Issues abiertos: `waica` 15, `platform` 6, `brik` 2, resto 0. Rate limit de GitHub: 5000 por hora REST y GraphQL. `gh` logueado como `chichex` con scope `repo`. Clones con tags al día (`dale-que-sale` tiene 4 tags locales que no están en el remoto). [verificado: `gh api`, `git tag`, `git ls-remote`]
- Stacks del usuario: Next.js + React + TS en `platform/web`, `tracker`, `cpanel/web`; Vite + React + Express en el dashboard de `dale-que-sale`; Go 1.26.3 y Node 24 locales; binarios Go en `cpanel`. Coding-policies cubren Go, TS, Node, React y Next, con reglas de web en celular en `references/react.md`. [verificado]
- El contrato de `skills` dice "Sin superficies UI" y no tiene build. [verificado: `skills/.sdd/project.md`]
- Existía un grill previo sobre un board SDD (`skills/.sdd/grills/2026-10-14-visualizacion-harness.md`). Por decisión del usuario se descartó por completo; este handoff no hereda ninguna de sus decisiones.

## Decisiones resueltas
1. Modelo: el issue es una identidad, no una etapa. Puede existir en cualquier etapa.
2. Modelo: la unidad raíz es el *trabajo*: un hilo que arranca en un pedido (libre o issue) y reúne grill, spec, runs, PRs, merge y publicación. Ancla el issue si existe; si no, el grill o la spec.
3. Modelo: siete etapas: Idea, Grill, Spec, Run, PR, Mergeado, Publicado. Grill es salteable.
4. Modelo: PR validado = el gate de `sdd-land`. Sub-estados: draft, abierto, validado, cambios pedidos, checks rojos.
5. Modelo: Publicado tiene dos sub-estados: taggeado (el merge está en un tag) y en prod.
6. Modelo: la ubicación (borrador local, repo, GitHub) es un atributo del artefacto; no cambia etapa ni estado.
7. Modelo: el linaje es un grafo con bifurcaciones visibles: specs reemplazadas, issues absorbidos y runs fallidos se dibujan como ramas.
8. Modelo: issues abiertos sin artefactos entran como Idea. PRs sin spec quedan afuera.
9. Modelo: todas las etapas muestran sub-estado: grill pausado/finalizado; spec borrador/aprobada/implementada/reemplazada; run corriendo/fallado/con PR; PR y Publicado como arriba.
10. Alcance: v1 = solo el grafo del flujo. Sin bandeja de pendientes ni métricas.
11. Alcance: uso en vivo y retrospectivo.
12. Alcance: todos los repos juntos. Se descubren escaneando los subdirectorios de la carpeta donde corre, por la existencia de `.sdd/project.md`, deduplicando worktrees por common-dir y salteando carpetas sin git.
13. Vista: layout de carriles por etapa. Siete columnas fijas; cada trabajo es una franja; las ramas van dentro de la franja.
14. Vista: dos niveles: tablero compacto y detalle expandido.
15. Vista: stacks como un nodo por capa encadenado 1 → N, cada uno con sub-estado; Mergeado también por capa.
16. Vista: terminados con filtro activos / todos, default activos. Un trabajo en prod sale del tablero por default.
17. Vista: orden por última actividad; repo como badge y filtro.
18. Vista: tiempo como Gantt en el detalle, una barra por etapa sobre el eje de tiempo, ramas fallidas o reemplazadas como barras paralelas. El tablero conserva los carriles.
19. Fuente: los trabajos anteriores al log entran con estado actual derivado de markers y GitHub, sin historia de transiciones.
20. Fuente: refresh con archivos locales observados en vivo; GitHub por polling cada 2 minutos; botón de refresh manual.
21. Fuente: "en prod" lo escribe SDD en el log. El visor no conoce infra ni `prama.yaml`. Hasta que exista `events-v2`, "en prod" queda sin dato.
22. Fuente: "taggeado" se lee del git local del clon, sin fetch. El visor nunca escribe en los clones.
23. Fuente: Idea = todos los issues abiertos de cada repo.
24. Superficie: web app local, un proceso que corre en la carpeta del workspace y sirve la UI en el browser.
25. Superficie: repo nuevo propio. Arquitectura preparada para pipeline: modelo de dominio separado de la UI, acciones como punto de extensión, solo lectura en v1.
26. Superficie: GitHub vía `gh` CLI como subproceso.
27. Superficie: distribución como binario Go, un archivo por plataforma, sin runtime.
28. Superficie: front en Vite + React + TypeScript, buildeado y embebido con `go:embed`, servido por `net/http` junto con una API JSON.
29. Diseño: sin GitHub (sin red, `gh` sin login, repo sin remote) degrada a solo local con aviso visible: Idea vacía, PRs sin sub-estado, antigüedad del snapshot. Nunca bloquea.
30. Diseño: kit UI Tailwind + shadcn/ui. El grafo y el Gantt son SVG propio.
31. Diseño: dirección visual fijada con `/design-system` en el repo nuevo, después de `/sdd-init` y antes de `/sdd-spec`.
32. Diseño: plataformas desktop y celular por LAN. Teclado básico, contraste AA, tema claro y oscuro.
33. Diseño: en celular, bajo cierto ancho el tablero pasa a lista (tarjeta por trabajo con repo, etapa y sub-estado); el detalle muestra el grafo y el Gantt con scroll horizontal.
34. Diseño: el server escucha en localhost por default; un flag habilita LAN y muestra la URL con IP y QR. Sin auth.

## Ramas pendientes
Ninguna.

## Diseño
- Superficies: una web nueva, sin sistema de diseño. Inventario vacío: no hay componentes, tokens, mocks ni Claude Design.
- Pantallas y flujos: tablero por carriles (siete columnas, una franja por trabajo, ramas dentro de la franja); detalle de un trabajo (grafo expandido con fechas y links, más Gantt por etapa); lista en celular bajo el breakpoint, con el detalle accesible al tocar.
- Estados: carga progresiva (los repos aparecen a medida que se escanean); vacío (ningún `.sdd/project.md` bajo la carpeta, con el mensaje de cómo inicializar uno); degradado sin GitHub (aviso visible, Idea vacía, PRs sin sub-estado); snapshot viejo con antigüedad visible.
- Componentes: ninguno para reusar. Kit Tailwind + shadcn/ui para filtros, badges y paneles; grafo y Gantt en SVG propio con React.
- Dirección visual: pendiente de `/design-system` en el repo nuevo. Tema claro y oscuro.
- Plataformas y accesibilidad: desktop y celular por LAN; navegación básica por teclado; contraste AA.
- Web en celular: aplican las reglas de `coding-policies/references/react.md` (no bloquear el zoom, safe-area, teclado virtual). Breakpoint lo decide el design-system.
- Capturas de referencia: ninguna. No existe mock ni referencia renderizable ni adjunta.

## Handoff
**Tema y alcance.** Una herramienta local, de solo lectura, que muestra cómo cada *trabajo* avanza por el flujo SDD en todos los repos de una carpeta: Idea → Grill → Spec → Run → PR → Mergeado → Publicado. Se dibuja como grafo con ramas en carriles por etapa, con un detalle tipo Gantt. La v1 es solo esa vista. Vive en un repo nuevo, se distribuye como binario Go con el front embebido, y queda preparada para crecer a un pipeline de desarrollo más automatizado.

**Decisiones.** Las 34 de `## Decisiones resueltas`, agrupadas por bloque: Modelo (1 a 9), Alcance (10 a 12), Vista (13 a 18), Fuente (19 a 23), Superficie (24 a 28), Diseño (29 a 34).

**Restricciones y no-objetivos.**
- No ejecuta acciones de los skills: no aprueba specs, no lanza runs, no mergea, no deploya.
- No escribe en los clones ni en GitHub.
- No se acopla a `prama.yaml`, repos infra ni al estándar Prama.
- No agrega UI ni build a `skills`. No toca `events-v1` ni `SDD-Tracking v1`.
- Sin bandeja, sin métricas, sin PRs fuera de SDD, sin auth en LAN.

**Supuestos explícitos.**
- Nombre del repo a definir; ejemplo: `chichex/sdd-viewer`.
- Escaneo con profundidad acotada (ejemplo: 3), salteando `node_modules` y `.git`. Carpetas con `.sdd/project.md` pero sin `.git` se listan como "sin git" y no se procesan.
- `#NN` se resuelve contra el remote `origin` del clon; `#288` y `owner/repo#288` son el mismo trabajo.
- Agrupación del trabajo: por issue cuando hay; sin issue, por referencia de grill o ruta de spec. Links: `issue=`, `grill=`, `superseded-by=`, los campos de cada evento y `Closes/Refs` en los bodies de PR.
- Mergeado se lee del evento `land.merged` o de `mergedAt` de GitHub, el que exista. Cubre los merges por cola que no registran evento.
- Se cachean los snapshots de GitHub en el directorio de datos del usuario, con antigüedad visible, para arranque rápido y uso offline.
- Intervalo de polling configurable; observación de archivos con fsnotify.
- El parser de eventos y markers se porta a Go, con los vectores de conformance de los docs como tests.
- Librería de grafo y Gantt, breakpoint móvil y mecanismo de releases del binario: los decide la spec o el design-system.

**Riesgos y diferidos.**
- `events.jsonl` es por máquina: un run hecho en otra máquina no aparece.
- Trabajo aparte en `skills`, con spec propia: `events-v2` con `publish.tagged` y `publish.deployed`, y `sdd-land` (o un skill nuevo) que taggee y deploye según el repo, usando el `/deploy` de `platform` como precedente. Hasta entonces "en prod" no se muestra.
- Monorepos: un PR puede quedar en varios tags (`api/v*`, `web/v*`). Taggeado se muestra por tag; cómo agregar lo decide la spec.
- Tags locales no pusheados pueden marcar "taggeado" antes de tiempo. Aceptado.
- Bandeja, métricas y pipeline automatizado son etapas futuras; la arquitectura las contempla, la v1 no las trae.

**Contexto para la sesión que escriba la spec.** Orden: 1) crear el repo; 2) `/sdd-init`; 3) `/design-system`; 4) `/sdd-spec --from-grill /home/chiche/workspace/.sdd/grills/2026-10-23-visor-flujo-sdd.md`. Hoy no hay repo ni contrato, así que la spec no puede correr todavía. Leer `skills/docs/sdd-events-v1.md`, `skills/docs/sdd-tracking-v1.md`, `skills/pi-extensions/sdd-artifacts/sdd-events.ts` e `index.ts` (parser a portar), la Fase 3 de `skills/claude/sdd-land/SKILL.md` (gate), y usar los cuatro `events.jsonl` reales como fixtures. La spec de `events-v2` y deploy es otra, en `skills`.
