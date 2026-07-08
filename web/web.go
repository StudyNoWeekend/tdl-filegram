package web

import "embed"

// FS 嵌入前端构建产物（web/dist）
//
//go:embed all:dist
var FS embed.FS
