package server

/*
import (
	"net/http"
	"os"
	"path"
	"strings"
)

const INDEX = "index.html"

type localFileSystem struct {
	//http.FileSystem
	root    string
	indexes bool
}

func LocalFile(root string, indexes bool) *localFileSystem {
	return &localFileSystem{
		//FileSystem: gin.Dir(root, indexes),
		root:    root,
		indexes: indexes,
	}
}

func Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		name := path.Join(l.root, p)
		stats, err := os.Stat(name)
		if err != nil {
			return false
		}
		if stats.IsDir() {
			if !l.indexes {
				index := path.Join(name, INDEX)
				_, err := os.Stat(index)
				if err != nil {
					return false
				}
			}
		}
		return true
	}
	return false
}

// Static returns a middleware handler that serves static files in the given directory.
func Serve(urlPrefix string, fs http.FileSystem) http.HandlerFunc {
	fileserver := http.FileServer(fs)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(res http.ResponseWriter, req *http.Request) {
		if fs.Exists(urlPrefix, req.URL.Path) {
			fileserver.ServeHTTP(res, req)
		}
	}
}
*/
