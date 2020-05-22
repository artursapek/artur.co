package index

import (
	"html/template"
	"log"
	"net/http"
	"strings"

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
	if strings.Contains(r.Referer(), "cryptowat.ch") {
		http.Redirect(w, r, "https://twitter.com/borrowcheck", 301)
		return
	}

	indexTemplate.Execute(w, struct {
		BlogEntries blog.Entries
	}{
		BlogEntries: blog.All[0:3],
	})
}
