package index

import (
	"html/template"
	"log"
	"net/http"

	"github.com/artursapek/artur.co/blog"
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
	indexTemplate.Execute(w, struct {
		BlogEntries blog.Entries
	}{
		BlogEntries: blog.All[0:3],
	})
}
