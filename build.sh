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
echo "  1) linux/amd64    (Linux x64)"
echo "  2) linux/arm64    (Linux ARM64)"
echo "  3) linux/arm/v7   (Linux ARMv7)"
echo "  4) darwin/amd64   (macOS x64)"
echo "  5) darwin/arm64   (macOS Apple Silicon)"
echo "  6) windows/amd64  (Windows x64)"
echo "  7) 当前平台 (不交叉编译)"
echo ""
read -p "输入序号 [1]: " choice
choice=${choice:-1}

case $choice in
  1) GOOS=linux   GOARCH=amd64        OUTPUT=sing-panel-linux-amd64 ;;
  2) GOOS=linux   GOARCH=arm64        OUTPUT=sing-panel-linux-arm64 ;;
  3) GOOS=linux   GOARCH=arm   GOARM=7 OUTPUT=sing-panel-linux-armv7 ;;
  4) GOOS=darwin  GOARCH=amd64        OUTPUT=sing-panel-darwin-amd64 ;;
  5) GOOS=darwin  GOARCH=arm64        OUTPUT=sing-panel-darwin-arm64 ;;
  6) GOOS=windows GOARCH=amd64        OUTPUT=sing-panel-windows-amd64.exe ;;
  7) GOOS=""      GOARCH=""           OUTPUT=sing-panel ;;
  *) echo "无效选择"; exit 1 ;;
esac

# Build backend
echo ""
echo "Building backend for ${GOOS:-$(go env GOOS)}/${GOARCH:-$(go env GOARCH)}..."
cd backend
go mod tidy
if [ -n "$GOOS" ]; then
  CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH ${GOARM:+GOARM=$GOARM} go build -tags "with_utls with_quic" -o ../$BUILD_DIR/$OUTPUT .
else
  go build -tags "with_utls with_quic" -o ../$BUILD_DIR/$OUTPUT .
fi
cd ..

echo ""
echo "=== Build Complete ==="
echo "Binary: $BUILD_DIR/$OUTPUT"
echo "Run with: ./$BUILD_DIR/$OUTPUT"
