// Package webui 内嵌前端构建产物（web/dist 构建后复制到 server/webui/dist）。
package webui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

func FS() fs.FS {
	sub, _ := fs.Sub(dist, "dist")
	return sub
}
