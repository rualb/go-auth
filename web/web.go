package web

import (
	"embed"
	"io/fs"
)

//go:embed go-auth/auth/assets/*
var fsAuthAssetsFS embed.FS

func MustAuthAssetsFS() fs.FS {
	res, err := fs.Sub(fsAuthAssetsFS, "go-auth/auth/assets")
	if err != nil {
		panic(err)
	}
	return res
}

//go:embed  go-auth/index.html
var fsAuthIndexHTML embed.FS

func MustAuthIndexHTML() string {

	data, err := fsAuthIndexHTML.ReadFile("go-auth/index.html")
	if err != nil {
		panic(err)
	}

	return string(data)
}
