package emails

import (
	"embed"
	"io/fs"

	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

//go:embed out/*
var outFolder embed.FS

//go:embed locales/*
var localeFolder embed.FS

func TemplateFS() (fs.FS, fs.FS) {
	sub, err := fs.Sub(outFolder, "out")
	assert.Nil(err, "'out' directory is not found in backend files", err)

	locales, err := fs.Sub(localeFolder, "locales")
	assert.Nil(err, "'locales' directory is not found in backend files", err)

	return sub, locales
}
