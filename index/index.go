package index

import (
	"html/template"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var (
	indexTemplate *template.Template
)

func init() {
	var parseErr error
	indexTemplate, parseErr = template.ParseFiles("templates/index.html")
	if parseErr != nil {
		log.Fatal(parseErr)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	indexTemplate.Execute(w, nil)
}
