#!/usr/bin/env bash
# Cross-compila PearDesk.exe da Linux con mingw-w64 + FFmpeg Windows libs.
# Su Ubuntu/Debian:  sudo bash build-windows.sh
# Con Docker:
#   docker run --rm -v "$(pwd)":/src -w /src ubuntu:22.04 bash build-windows.sh
set -euo pipefail

echo "=== PearDesk Windows EXE builder ==="

# ── 1. System dependencies ────────────────────────────────────────────────────
if command -v apt-get &>/dev/null; then
  apt-get update -qq
  apt-get install -y --no-install-recommends \
    gcc-mingw-w64-x86-64 pkg-config wget curl ca-certificates \
    libx11-dev libxrandr-dev libxcursor-dev libxi-dev \
    libxinerama-dev libgl1-mesa-dev libxtst-dev \
    libfontconfig1-dev libfreetype6-dev
fi

# ── 2. Install Go ─────────────────────────────────────────────────────────────
GO_VERSION="1.25.5"
if ! command -v go &>/dev/null || [[ "$(go version)" != *"go${GO_VERSION}"* ]]; then
  echo "Installing Go ${GO_VERSION}..."
  wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf /tmp/go.tar.gz
  rm /tmp/go.tar.gz
fi
export PATH="/usr/local/go/bin:$PATH"
go version

# ── 3. Download pre-built FFmpeg Windows libs (required for CGO video codec) ──
#
# The Windows build needs FFmpeg headers + import libs compiled for mingw-w64.
# We pull the official gyan.dev release and extract only what CGO needs.
FFMPEG_WIN="ffmpeg-7.1.1-full_build-shared"
FFMPEG_URL="https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-lgpl-shared.zip"
FFMPEG_ZIP="/tmp/ffmpeg-win.zip"
FFMPEG_DIR="/tmp/ffmpeg-win"

if [ ! -d "${FFMPEG_DIR}" ]; then
  echo "Downloading FFmpeg Windows shared libs..."
  wget -q "${FFMPEG_URL}" -O "${FFMPEG_ZIP}"
  mkdir -p "${FFMPEG_DIR}"
  unzip -q "${FFMPEG_ZIP}" -d "${FFMPEG_DIR}"
  # Flatten: the zip contains one top-level directory
  INNER=$(find "${FFMPEG_DIR}" -maxdepth 1 -type d | tail -n1)
  if [ "${INNER}" != "${FFMPEG_DIR}" ]; then
    mv "${INNER}"/* "${FFMPEG_DIR}/"
    rmdir "${INNER}"
  fi
fi

FFMPEG_INC="${FFMPEG_DIR}/include"
FFMPEG_LIB="${FFMPEG_DIR}/lib"

# ── 4. Build binary ───────────────────────────────────────────────────────────
echo "Building PearDesk.exe..."
mkdir -p dist

CGO_ENABLED=1 \
GOOS=windows \
GOARCH=amd64 \
CC=x86_64-w64-mingw32-gcc \
PKG_CONFIG_PATH="" \
CGO_CFLAGS="-I${FFMPEG_INC}" \
CGO_LDFLAGS="-L${FFMPEG_LIB} -lavcodec -lavutil -lswscale -lx264" \
GOTOOLCHAIN=local \
  go build \
    -ldflags="-s -w -H=windowsgui" \
    -o dist/PearDesk.exe \
    ./cmd/peardesk

echo ""
echo "✓ Done: dist/PearDesk.exe"
ls -lh dist/PearDesk.exe
echo ""
echo "NOTA: copia nella stessa cartella dell'exe le DLL da ${FFMPEG_DIR}/bin/:"
echo "  avcodec-*.dll  avutil-*.dll  swscale-*.dll"
