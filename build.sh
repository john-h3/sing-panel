#!/bin/bash

set -e

echo "=== Building Sing Box Panel ==="

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Select target architecture. Pass a target such as linux/arm64 to skip the
# interactive prompt. With no target, keep the original interactive mode.
GOOS=""
GOARCH=""
GOARM=""
OUTPUT=""
TARGET=""
ENABLE_BROTLI=0

for argument in "$@"; do
  case "$argument" in
    --brotli)
      ENABLE_BROTLI=1
      ;;
    --help|-h)
      TARGET="$argument"
      ;;
    *)
      if [ -n "$TARGET" ]; then
        echo "只能指定一个目标架构: $argument" >&2
        exit 1
      fi
      TARGET="$argument"
      ;;
  esac
done

if [ "$TARGET" = "--help" ] || [ "$TARGET" = "-h" ]; then
  echo "Usage: $0 [TARGET]"
  echo ""
  echo "TARGET may be one of:"
  echo "  native         current platform (default in interactive mode)"
  echo "  linux/amd64"
  echo "  linux/arm64"
  echo "  linux/arm/v7"
  echo "  darwin/amd64"
  echo "  darwin/arm64"
  echo "  windows/amd64"
  echo ""
  echo "Options:"
  echo "  --brotli         also generate Brotli-compressed assets"
  exit 0
fi

if [ -z "$TARGET" ]; then
  echo ""
  echo "请选择目标架构:"
  echo "  1) 当前平台 (不交叉编译)"
  echo "  2) linux/amd64    (Linux x64)"
  echo "  3) linux/arm64    (Linux ARM64)"
  echo "  4) linux/arm/v7   (Linux ARMv7)"
  echo "  5) darwin/amd64   (macOS x64)"
  echo "  6) darwin/arm64   (macOS Apple Silicon)"
  echo "  7) windows/amd64  (Windows x64)"
  echo ""
  read -p "输入序号 [1]: " choice
  choice=${choice:-1}
  case "$choice" in
    1) TARGET=native ;;
    2) TARGET=linux/amd64 ;;
    3) TARGET=linux/arm64 ;;
    4) TARGET=linux/arm/v7 ;;
    5) TARGET=darwin/amd64 ;;
    6) TARGET=darwin/arm64 ;;
    7) TARGET=windows/amd64 ;;
    *) echo "无效选择"; exit 1 ;;
  esac
fi

case "$TARGET" in
  native)
    OUTPUT=sing-panel
    ;;
  linux/amd64)
    GOOS=linux GOARCH=amd64 OUTPUT=sing-panel-linux-amd64
    ;;
  linux/arm64)
    GOOS=linux GOARCH=arm64 OUTPUT=sing-panel-linux-arm64
    ;;
  linux/arm/v7)
    GOOS=linux GOARCH=arm GOARM=7 OUTPUT=sing-panel-linux-armv7
    ;;
  darwin/amd64)
    GOOS=darwin GOARCH=amd64 OUTPUT=sing-panel-darwin-amd64
    ;;
  darwin/arm64)
    GOOS=darwin GOARCH=arm64 OUTPUT=sing-panel-darwin-arm64
    ;;
  windows/amd64)
    GOOS=windows GOARCH=amd64 OUTPUT=sing-panel-windows-amd64.exe
    ;;
  *)
    echo "无效目标架构: $TARGET" >&2
    echo "运行 \"$0 --help\" 查看支持的目标架构" >&2
    exit 1
    ;;
esac

# Create build directory
BUILD_DIR="$SCRIPT_DIR/build"
mkdir -p "$BUILD_DIR"

# Build frontend when its inputs changed. The marker lives under build/, which
# is already ignored, so repeated deployments can reuse the generated assets.
FRONTEND_CACHE="$BUILD_DIR/frontend.sha256"
FRONTEND_HASH="$(
  git -C "$SCRIPT_DIR" ls-files -co --exclude-standard -- frontend |
    while IFS= read -r file; do
      sha256sum "$SCRIPT_DIR/$file"
    done |
    sha256sum |
    cut -d ' ' -f1
 )"
FRONTEND_HASH="$(printf '%s:%s' "$FRONTEND_HASH" "$ENABLE_BROTLI" | sha256sum | cut -d ' ' -f1)"
if [ -f "$FRONTEND_CACHE" ] && [ "$(<"$FRONTEND_CACHE")" = "$FRONTEND_HASH" ] && [ -f "$SCRIPT_DIR/frontend/dist/index.html" ]; then
  echo "Frontend unchanged; reusing cached build."
else
  echo "Building frontend..."
  cd "$SCRIPT_DIR/frontend"
  npm install --prefer-offline --no-audit --no-fund
  if [ "$ENABLE_BROTLI" -eq 1 ]; then
    SING_PANEL_BROTLI=1 npm run build
  else
    npm run build
  fi
  cd "$SCRIPT_DIR"
  printf '%s\n' "$FRONTEND_HASH" > "$FRONTEND_CACHE"
fi

# Copy frontend dist to backend for embedding
echo "Preparing embedded frontend..."
rm -rf "$SCRIPT_DIR/backend/frontend"
mkdir -p "$SCRIPT_DIR/backend/frontend"
cp -r "$SCRIPT_DIR/frontend/dist" "$SCRIPT_DIR/backend/frontend/dist"

# Build backend
echo ""
echo "Building backend for ${GOOS:-$(go env GOOS)}/${GOARCH:-$(go env GOARCH)}..."
cd "$SCRIPT_DIR/backend"
go mod tidy
BUILD_TIME="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
LDFLAGS="-X sing-panel/services.BuildTime=$BUILD_TIME"
if [ -n "$GOOS" ]; then
  CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH ${GOARM:+GOARM=$GOARM} go build -tags "with_utls with_quic with_gvisor with_clash_api" -ldflags "$LDFLAGS" -o "$BUILD_DIR/$OUTPUT" .
else
  go build -tags "with_utls with_quic with_gvisor with_clash_api" -ldflags "$LDFLAGS" -o "$BUILD_DIR/$OUTPUT" .
fi
cd "$SCRIPT_DIR"

echo ""
echo "=== Build Complete ==="
echo "Binary: $BUILD_DIR/$OUTPUT"
echo "Run with: $BUILD_DIR/$OUTPUT"
