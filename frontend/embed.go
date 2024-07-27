package frontend

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed dist/*
var distFolder embed.FS

func StaticFilesPath() (fs.FS, fs.FS) {
	dist, err := fs.Sub(distFolder, "dist")
	if err != nil {
		fmt.Println("'dist' directory is not found in vite build files")
		panic(err)
	}

	assets, err := fs.Sub(distFolder, "dist/assets")
	if err != nil {
		fmt.Println("'dist/assets' directory is not found in vite build files")
		panic(err)
	}

	return dist, assets
}
