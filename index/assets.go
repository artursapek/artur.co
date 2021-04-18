package index

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func AssetsHandler(rw http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	http.ServeFile(rw, req, req.URL.Path[1:])
}

func RawHandler(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	path := filepath.Join("/mnt/raw", p.ByName("path"))
	http.ServeFile(rw, req, path)
}

func GetSingleAssetHandler(path string) httprouter.Handle {
	return func(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
		http.ServeFile(rw, req, path)
	}
}

func GetAssetHandler(root string) httprouter.Handle {
	return func(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
		// Cache images
		name := p.ByName("path")
		if strings.Contains(name, ".png") || strings.Contains(name, ".jpg") {
			rw.Header().Set("Cache-Control", "max-age=86400")
		}
		path := filepath.Join(root, name)
		http.ServeFile(rw, req, path)
	}

}
