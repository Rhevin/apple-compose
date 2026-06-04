#!/bin/sh
set -e

REPO="Rhevin/apple-compose"
BIN="apple-compose"
INSTALL_DIR="/usr/local/bin"

# ── Checks ────────────────────────────────────────────────────────────────────

if [ "$(uname -s)" != "Darwin" ]; then
  echo "error: apple-compose only runs on macOS" >&2
  exit 1
fi

if [ "$(uname -m)" != "arm64" ]; then
  echo "error: apple-compose requires Apple Silicon (arm64)" >&2
  exit 1
fi

# ── Resolve version ───────────────────────────────────────────────────────────

if [ -z "$VERSION" ]; then
  echo "Fetching latest release..."
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | sed 's/.*"tag_name": *"\(.*\)".*/\1/')
fi

if [ -z "$VERSION" ]; then
  echo "error: could not determine latest version" >&2
  exit 1
fi

echo "Installing apple-compose ${VERSION}..."

# ── Download ──────────────────────────────────────────────────────────────────

ARCHIVE="${BIN}_${VERSION#v}_darwin_arm64.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading ${URL}..."
curl -fsSL "$URL" -o "${TMPDIR}/${ARCHIVE}"

# ── Verify checksum ───────────────────────────────────────────────────────────

CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
curl -fsSL "$CHECKSUM_URL" -o "${TMPDIR}/checksums.txt"

cd "$TMPDIR"
grep "${ARCHIVE}" checksums.txt | shasum -a 256 -c -
cd - > /dev/null

# ── Install ───────────────────────────────────────────────────────────────────

tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

if [ -w "$INSTALL_DIR" ]; then
  mv "${TMPDIR}/${BIN}" "${INSTALL_DIR}/${BIN}"
else
  echo "Installing to ${INSTALL_DIR} (may require password)..."
  sudo mv "${TMPDIR}/${BIN}" "${INSTALL_DIR}/${BIN}"
fi

chmod +x "${INSTALL_DIR}/${BIN}"

# ── Done ──────────────────────────────────────────────────────────────────────

echo ""
echo "✓ apple-compose ${VERSION} installed to ${INSTALL_DIR}/${BIN}"
echo ""
echo "Next steps:"
echo "  container system start"
echo "  apple-compose --version"
echo "  apple-compose up --dry-run"
