package photos

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
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

type Permalink struct {
	ContentItem
	time.Time
	Location
}

func PhotoPermalinkHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	item := ContentItem{
		Type: "photo",
		Src:  params.ByName("path")[1:],
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
