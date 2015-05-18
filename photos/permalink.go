package photos

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
)

var (
	permalinkTemplate *template.Template
)

func init() {
	var permalinkTemplateErr error
	permalinkTemplate, permalinkTemplateErr = template.ParseFiles("templates/photos/permalink.html")
	if permalinkTemplateErr != nil {
		panic(permalinkTemplateErr)
	}
}

type Location struct {
	Lat, Lng float64
}

func (loc Location) String() string {
	return fmt.Sprintf("(%.5f, %.5f)", loc.Lat, loc.Lng)
}

func (loc Location) Valid() bool {
	return loc.Lat != 0 && loc.Lng != 0
}

type Permalink struct {
	ContentItem
	time.Time
	Location
}

func PhotoPermalinkHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	permalinkHandler("photo", w, r, params)
}

func VideoPermalinkHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	permalinkHandler("video", w, r, params)
}

func permalinkHandler(t string, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	item := ContentItem{
		Type: t,
		Src:  params.ByName("path")[1:],
	}
	if _, err := os.Stat(item.RawPath()); err != nil {
		http.Error(w, "No such "+t, 404)
		return
	}
	renderErr := permalinkTemplate.Execute(w, Permalink{
		ContentItem: item,
		Time:        item.Timestamp(),
		Location:    item.Location(),
	})
	if renderErr != nil {
		log.Fatal(renderErr)
	}
}
