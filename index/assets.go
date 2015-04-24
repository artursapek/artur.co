package index

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func AssetsHandler(rw http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	http.ServeFile(rw, req, req.URL.Path[1:])
}
