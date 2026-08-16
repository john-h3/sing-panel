#!/bin/bash

set -e

echo "=== Building Sing Box Panel ==="

# Create build directory
BUILD_DIR="build"
mkdir -p "$BUILD_DIR"

# Build frontend
echo "Building frontend..."
cd frontend
npm install
npm run build
cd ..

# Copy frontend dist to backend for embedding
echo "Preparing embedded frontend..."
rm -rf backend/frontend
mkdir -p backend/frontend
cp -r frontend/dist backend/frontend/dist

# Select target architecture
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

case $choice in
  1) GOOS=""      GOARCH=""           OUTPUT=sing-panel ;;
  2) GOOS=linux   GOARCH=amd64        OUTPUT=sing-panel-linux-amd64 ;;
  3) GOOS=linux   GOARCH=arm64        OUTPUT=sing-panel-linux-arm64 ;;
  4) GOOS=linux   GOARCH=arm   GOARM=7 OUTPUT=sing-panel-linux-armv7 ;;
  5) GOOS=darwin  GOARCH=amd64        OUTPUT=sing-panel-darwin-amd64 ;;
  6) GOOS=darwin  GOARCH=arm64        OUTPUT=sing-panel-darwin-arm64 ;;
  7) GOOS=windows GOARCH=amd64        OUTPUT=sing-panel-windows-amd64.exe ;;
  *) echo "无效选择"; exit 1 ;;
esac

# Build backend
echo ""
echo "Building backend for ${GOOS:-$(go env GOOS)}/${GOARCH:-$(go env GOARCH)}..."
cd backend
go mod tidy
if [ -n "$GOOS" ]; then
  CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH ${GOARM:+GOARM=$GOARM} go build -tags "with_utls with_quic with_gvisor with_clash_api" -o ../$BUILD_DIR/$OUTPUT .
else
  go build -tags "with_utls with_quic with_gvisor with_clash_api" -o ../$BUILD_DIR/$OUTPUT .
fi
cd ..

echo ""
echo "=== Build Complete ==="
echo "Binary: $BUILD_DIR/$OUTPUT"
echo "Run with: ./$BUILD_DIR/$OUTPUT"
