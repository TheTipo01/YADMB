package api

import (
	"embed"
	"github.com/gin-contrib/static"
	"io/fs"
	"net/http"
)

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(_ string, path string) bool {
	_, err := e.Open(path)
	return err == nil
}

func EmbedFolder(fsEmbed *embed.FS, targetPath string) static.ServeFileSystem {
	fsys, _ := fs.Sub(fsEmbed, targetPath)
	return embedFileSystem{
		FileSystem: http.FS(fsys),
	}
}
