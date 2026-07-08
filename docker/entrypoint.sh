#!/bin/sh
set -e

mkdir -p /run/nginx

# 启动后端（固定端口 8743）
APP_PORT=8743 ./tdl-filegram &

# 启动 nginx（固定端口 8744，前台运行）
exec nginx -g 'daemon off;'
