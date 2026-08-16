#!/bin/bash
#
# Sing Box Panel 启动脚本（开发环境用）
#
# 使用 build/ 下的编译产物启动，自动选择当前平台对应的二进制。
# 数据目录默认使用项目根目录下的 data/，可通过 --data-dir 指定。
#
# 用法:
#   ./start.sh                     # 默认监听 :8080，数据目录 ./data
#   ./start.sh --listen :3000      # 指定端口
#   ./start.sh --data-dir /etc/sing-panel   # 指定数据目录
#   ./start.sh install             # 安装为系统服务 (透传给二进制)

set -e

# 脚本所在目录（项目根目录）
BASE_DIR="$(cd "$(dirname "$(readlink -f "$0" 2>/dev/null || echo "$0")")" && pwd)"

# install/uninstall/help 直接透传给二进制，不解析参数
case "${1:-run}" in
  install|uninstall|help|-h|--help)
    exec "$BASE_DIR/build/sing-panel" "$@"
    ;;
esac

DATA_DIR="$BASE_DIR/data"
LISTEN=":8080"
EXTRA_ARGS=()

# 解析参数 (run 模式)
while [ $# -gt 0 ]; do
  case "$1" in
    --data-dir)
      DATA_DIR="${2:-$DATA_DIR}"
      shift 2
      ;;
    --data-dir=*)
      DATA_DIR="${1#*=}"
      shift
      ;;
    --listen)
      LISTEN="${2:-$LISTEN}"
      EXTRA_ARGS+=("--listen" "$LISTEN")
      shift 2
      ;;
    --listen=*)
      EXTRA_ARGS+=("$1")
      shift
      ;;
    *)
      EXTRA_ARGS+=("$1")
      shift
      ;;
  esac
done

# 选择编译产物
pick_binary() {
  local cand
  if [ -x "$BASE_DIR/build/sing-panel" ]; then
    echo "$BASE_DIR/build/sing-panel"
    return
  fi
  local arch_map="x86_64:amd64 aarch64:arm64 arm64:arm64 armv7l:armv7 armv6l:armv6"
  local machine
  machine="$(uname -m 2>/dev/null || echo unknown)"
  for pair in $arch_map; do
    local native="${pair%%:*}" goarch="${pair##*:}"
    if [ "$machine" = "$native" ]; then
      cand="$BASE_DIR/build/sing-panel-linux-$goarch"
      if [ -x "$cand" ]; then
        echo "$cand"
        return
      fi
    fi
  done
  echo ""
}

BINARY="$(pick_binary)"
if [ -z "$BINARY" ]; then
  echo "错误: 未找到编译产物 (build/sing-panel)。请先运行 ./build.sh 构建。" >&2
  exit 1
fi

mkdir -p "$DATA_DIR"
echo "==> 二进制: $BINARY"
echo "==> 数据目录: $DATA_DIR"
echo "==> 监听: $LISTEN"
echo ""

exec "$BINARY" run --data-dir "$DATA_DIR" "${EXTRA_ARGS[@]}"
