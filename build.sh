#!/bin/bash

set -e

echo "=== Building Sing Box Panel ==="

# Build frontend
echo "Building frontend..."
cd frontend
npm install
npm run build
cd ..

# Build backend
echo "Building backend..."
cd backend
go mod tidy
go build -o ../sing-panel .
cd ..

echo "=== Build Complete ==="
echo "Binary: ./sing-panel"
echo "Run with: ./sing-panel"
