#!/usr/bin/env bash
# 一键构建：前端 -> 内嵌到后端 -> 单个二进制
# 用法：scripts/build.sh [GOOS] [GOARCH]    例：scripts/build.sh linux amd64
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VERSION="$(cat "$ROOT/VERSION")"
GOOS="${1:-$(go env GOOS)}"
GOARCH="${2:-$(go env GOARCH)}"

echo ">> 构建前端"
cd "$ROOT/web"
[ -d node_modules ] || npm install --no-audit --no-fund
npm run build

echo ">> 复制前端产物到 server/webui/dist"
rm -rf "$ROOT/server/webui/dist"
cp -r "$ROOT/web/dist" "$ROOT/server/webui/dist"

echo ">> 编译后端 ($GOOS/$GOARCH, CGO_ENABLED=0)"
cd "$ROOT/server"
mkdir -p "$ROOT/dist"
CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -trimpath -ldflags "-s -w -X main.Version=$VERSION" \
  -o "$ROOT/dist/village-mgmt-$GOOS-$GOARCH" .
cp "$ROOT/server/config.example.yaml" "$ROOT/dist/config.example.yaml"
echo ">> 完成：dist/village-mgmt-$GOOS-$GOARCH"
