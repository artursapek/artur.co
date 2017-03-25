package index

import (
	"net/http"
	"path/filepath"

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
		path := filepath.Join(root, p.ByName("path"))
		http.ServeFile(rw, req, path)
	}

}
