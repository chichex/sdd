#!/bin/sh
# install.sh: instala o actualiza visor desde el último release de GitHub.
#
#   curl -fsSL https://raw.githubusercontent.com/chichex/visor/main/install.sh | sh
#
# Variables opcionales:
#   VISOR_VERSION      versión exacta a instalar (por ejemplo 0.3.0); default: la última.
#                      Mínimo 0.3.0: antes los binarios se llamaban sdd_*.
#   VISOR_INSTALL_DIR  carpeta destino; default: ~/.local/bin.
#
# Es idempotente: si la versión pedida ya está instalada, no toca nada.
# Verifica el binario contra checksums.txt del release antes de instalarlo.
set -eu

REPO="chichex/visor"
INSTALL_DIR="${VISOR_INSTALL_DIR:-$HOME/.local/bin}"

fail() {
	printf 'install.sh: %s\n' "$*" >&2
	exit 1
}

need() {
	command -v "$1" >/dev/null 2>&1 || fail "falta el comando $1"
}

need curl
need uname
need mktemp

case "$(uname -s)" in
Linux) os=linux ;;
Darwin) os=darwin ;;
*) fail "sistema no soportado: $(uname -s). En Windows bajá el .exe de https://github.com/$REPO/releases" ;;
esac

case "$(uname -m)" in
x86_64 | amd64) arch=amd64 ;;
aarch64 | arm64) arch=arm64 ;;
*) fail "arquitectura no soportada: $(uname -m)" ;;
esac

if [ -n "${VISOR_VERSION:-}" ]; then
	version="${VISOR_VERSION#v}"
else
	# /releases/latest redirige a /releases/tag/v<version>: sin API ni rate limit.
	# La query string única evita que el CDN de GitHub sirva un redirect viejo
	# durante los minutos posteriores a publicar un release.
	url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest?t=$(date +%s)") ||
		fail "no pude consultar el último release de $REPO"
	url="${url%%\?*}"
	version="${url##*/tag/v}"
	[ "$version" != "$url" ] || fail "no reconozco el último release: $url"
fi

asset="visor_${version}_${os}_${arch}"
base="https://github.com/$REPO/releases/download/v$version"
target="$INSTALL_DIR/visor"

if [ -x "$target" ]; then
	current=$("$target" --version 2>/dev/null | awk '{print $2}') || current=
	if [ "$current" = "$version" ]; then
		echo "visor $version ya está instalado en $target"
		exit 0
	fi
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "descargando $asset"
curl -fsSL -o "$tmp/$asset" "$base/$asset" || fail "no existe $base/$asset"
curl -fsSL -o "$tmp/checksums.txt" "$base/checksums.txt" || fail "no pude bajar $base/checksums.txt"

if command -v sha256sum >/dev/null 2>&1; then
	sum="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
	sum="shasum -a 256"
else
	fail "falta sha256sum o shasum para verificar el binario"
fi

line=$(grep " $asset\$" "$tmp/checksums.txt") || fail "$asset no figura en checksums.txt"
(cd "$tmp" && printf '%s\n' "$line" | $sum -c --status -) || fail "checksum inválido para $asset"

chmod 0755 "$tmp/$asset"
mkdir -p "$INSTALL_DIR"
mv -f "$tmp/$asset" "$target"
echo "visor $version instalado en $target"

case ":$PATH:" in
*":$INSTALL_DIR:"*) ;;
*) echo "agregá $INSTALL_DIR al PATH: export PATH=\"$INSTALL_DIR:\$PATH\"" ;;
esac
