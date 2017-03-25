package index

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

func AssetsHandler(rw http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	http.ServeFile(rw, req, req.URL.Path[1:])
}

func RawHandler(rw http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	p := filepath.Join("/mnt/raw", req.URL.Path[4:])
	log.Println(p)
	http.ServeFile(rw, req, p)
}
