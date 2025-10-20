package jabulani

import (
	"embed"
	"io/fs"

	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

//go:embed dist/*
var distFolder embed.FS

func StaticFilesPath() (fs.FS, fs.FS) {
	dist, err := fs.Sub(distFolder, "dist")
	assert.Nil(err, "'dist' directory is not found in vite build files", err)

	assets, err := fs.Sub(distFolder, "dist/assets")
	assert.Nil(err, "'dist/assets' directory is not found in vite build files", err)

	return dist, assets
}
