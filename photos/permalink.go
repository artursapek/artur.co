package photos

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/artursapek/artur.co/config"
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

	PrevSiblings, NextSiblings []ContentItem

	NextLink string
	PrevLink string
}

func PhotoPermalinkHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if err := albumsAuthWall(w, r); err != nil {
		return
	}

	if params.ByName("path") == "" || params.ByName("path") == "/" {
		http.Redirect(w, r, "/albums", 302)
		return
	}

	permalinkHandler("photo", w, r, params)
}

func VideoPermalinkHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if err := albumsAuthWall(w, r); err != nil {
		return
	}
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

	var (
		nextLink, prevLink    string
		base                      = filepath.Join(filepath.Dir(item.RawPath()), "*")
		siblingPaths, globErr     = filepath.Glob(base)
		currentIndex          int = -1

		siblings ItemsByTimestamp

		prevSibStart, prevSibEnd int
		nextSibStart, nextSibEnd int
	)

	if globErr == nil {
		for _, fn := range siblingPaths {
			sib := ContentItem{
				Type: t,
				Src:  strings.Replace(fn, filepath.Join(config.Config.RawRoot, t+"s")+"/", "", 1),
			}

			if sib.Timestamp().Unix() > 0 {
				siblings = append(siblings, sib)
			}
		}

		sort.Sort(siblings)

		for i, sib := range siblings {
			if sib.RawPath() == item.RawPath() {
				currentIndex = i
				break
			}
		}

		if currentIndex > -1 {
			if currentIndex > 0 {
				prevLink = siblings[currentIndex-1].Permalink()
			}
			if currentIndex < len(siblings)-1 {
				nextLink = siblings[currentIndex+1].Permalink()
			}

			prevSibStart = currentIndex - 10
			prevSibEnd = currentIndex
			if prevSibEnd < 0 {
				prevSibEnd = 0
			}
			if prevSibStart < 0 {
				prevSibStart = 0
			}

			nextSibStart = currentIndex + 1
			nextSibEnd = currentIndex + 10
			if nextSibStart >= len(siblings) {
				nextSibStart = len(siblings) - 1
			}
			if nextSibEnd >= len(siblings) {
				nextSibEnd = len(siblings) - 1
			}
		}

	}

	renderErr := permalinkTemplate.Execute(w, Permalink{
		ContentItem: item,
		Time:        item.Timestamp(),
		Location:    item.Location(),

		NextLink: nextLink,
		PrevLink: prevLink,

		PrevSiblings: siblings[prevSibStart:prevSibEnd],
		NextSiblings: siblings[nextSibStart:nextSibEnd],
	})
	if renderErr != nil {
		log.Fatal(renderErr)
	}
}
