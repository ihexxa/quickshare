package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed public/*
var embedFS embed.FS

type EmbedStaticFS struct {
	http.FileSystem
}

func NewEmbedStaticFS() (*EmbedStaticFS, error) {
	// the public folder will temporarily be copied to here in building
	publicFS, err := fs.Sub(embedFS, "public")
	if err != nil {
		return nil, err
	}
	httpFS := http.FS(publicFS)

	return &EmbedStaticFS{
		FileSystem: httpFS,
	}, nil
}

func (efs *EmbedStaticFS) Exists(prefix string, path string) bool {
	// prefix should already be considered by http.FileSystem
	_, err := efs.Open(path)
	if err != nil {
		// TODO: need more checking
		// if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return true
}
