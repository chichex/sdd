# sdd

Visor local del flujo SDD. Muestra cómo cada *trabajo* avanza por
Idea → Grill → Spec → Run → PR → Mergeado → Publicado en todos los repos de una
carpeta, como grafo con ramas en carriles por etapa y con un detalle tipo Gantt.
Solo lectura. Consume el log `events-v1` y los markers `SDD-Tracking v1` que
escriben los skills de [`chichex/skills`](https://github.com/chichex/skills),
más GitHub vía `gh`.

Estado: esqueleto. Las decisiones de producto están en
[`.sdd/grills/2026-10-23-visor-flujo-sdd.md`](.sdd/grills/2026-10-23-visor-flujo-sdd.md).

## Cómo se construye

Binario Go con el front (Vite + React + TypeScript) embebido con `go:embed`.
Node hace falta solo para buildear el front; el binario no lo necesita.

```sh
make build          # npm ci + vite build + go build → bin/sdd
./bin/sdd --dir ~/workspace
./bin/sdd --dir ~/workspace --host 0.0.0.0   # para abrirlo desde la LAN
make test
```

## Próximos pasos

1. `/sdd-init` para el contrato de autonomía.
2. `/design-system` para la dirección visual.
3. `/sdd-spec --from-grill .sdd/grills/2026-10-23-visor-flujo-sdd.md`.
