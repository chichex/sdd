#!/usr/bin/env bash
# Arma un workspace de fixtures para el visor y imprime su ruta. Crea repos
# con `git init` (sin red), copia los logs reales de testdata/events/ al
# common-dir de cada uno y suma los casos de borde del escaneo: un worktree,
# un repo sin remote, una carpeta sin git, una a profundidad 4 y node_modules.
# Uso: bash testdata/mkworkspace.sh [destino]   (sin destino usa mktemp -d)
set -euo pipefail

here=$(cd "$(dirname "$0")" && pwd)
dest=${1:-$(mktemp -d)}
mkdir -p "$dest"
dest=$(cd "$dest" && pwd)
# Aislado de la configuración git del usuario: sin hooks, firmas ni plantillas.
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
g() { git -c user.name=fixture -c user.email=fixture@example.invalid -c init.defaultBranch=main -c commit.gpgsign=false "$@"; }
# Fecha fija para los artefactos: el ts de un trabajo sin historia es su mtime.
stamp=202610010000

project() { # project <dir>
  mkdir -p "$1/.sdd"
  printf '# Contrato de autonomia — %s\n<!-- SDD-Tracking: version=1; type=project; generated-at=2026-10-25 -->\n\n## Stack\nFixture.\n' "$(basename "$1")" >"$1/.sdd/project.md"
}

repo() { # repo <dir relativo> <remote|-> <log de testdata/events|->
  local dir=$dest/$1
  project "$dir"
  g -C "$dir" init -q
  if [ "$2" != - ]; then g -C "$dir" remote add origin "$2"; fi
  g -C "$dir" add .sdd/project.md
  g -C "$dir" commit -q -m init
  if [ "$3" != - ]; then
    mkdir -p "$dir/.git/sdd"
    cp "$here/events/$3" "$dir/.git/sdd/events.jsonl"
  fi
}

artifact() { # artifact <fixture de testdata/workspace> <destino>
  mkdir -p "$(dirname "$2")"
  cp "$here/workspace/$1" "$2"
  touch -t "$stamp" "$2"
}

repo dale-que-sale https://github.com/pramaestudio/dale-que-sale.git dale-que-sale.jsonl
repo waica git@github.com:chichex/waica.git waica.jsonl
g -C "$dest/waica" worktree add -q "$dest/waica-wt" -b wt
artifact waica/spec-issue-203.md "$dest/waica/.sdd/specs/issue-203.md"
artifact waica/spec-issue-203-common.md "$dest/waica/.git/sdd/specs/issue-203.md"
repo platform ssh://git@github.com/pramaestudio/platform.git platform.jsonl
artifact platform/grill-backlog-norte.md "$dest/platform/.git/sdd/grills/2026-10-08-backlog-norte-issues-autonomos.md"
repo chichex/skills https://github.com/chichex/skills skills.jsonl
repo sin-remote - -
project "$dest/sin-git"
repo uno/dos/tres/profundo https://github.com/fixture/profundo.git -
project "$dest/node_modules/paquete"

printf '%s\n' "$dest"
