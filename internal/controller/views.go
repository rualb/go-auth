package controller

import (
	"embed"
	"io/fs"
)

//go:embed views/auth/*.html
var viewsFs embed.FS

func ViewsFs() fs.FS {
	return viewsFs
}
