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
	thumbsTemplate    *template.Template
)

func init() {
	var permalinkTemplateErr error
	permalinkTemplate, permalinkTemplateErr = template.ParseFiles("templates/photos/permalink.html")
	if permalinkTemplateErr != nil {
		panic(permalinkTemplateErr)
	}

	var thumbsTemplateErr error
	thumbsTemplate, thumbsTemplateErr = template.ParseFiles("templates/photos/thumbs.html")
	if thumbsTemplateErr != nil {
		panic(thumbsTemplateErr)
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

func PhotoThumbsHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if err := albumsAuthWall(w, r); err != nil {
		return
	}

	var (
		item = ContentItem{
			Type: "photo",
			Src:  params.ByName("path")[1:],
		}

		prevSiblings, nextSiblings = getSiblings(item, 20)

		items ItemsByTimestamp

		prev, next string
	)

	if len(prevSiblings) > 10 {
		prev = "/photos/thumbs/" + prevSiblings[len(prevSiblings)-10].Src
	} else if len(prevSiblings) > 0 {
		prev = "/photos/thumbs/" + prevSiblings[0].Src
	}
	if len(nextSiblings) > 0 {
		next = "/photos/thumbs/" + nextSiblings[len(nextSiblings)-1].Src
	}

	if len(nextSiblings) > 20 {
		items = nextSiblings[0:20]
	} else {
		items = nextSiblings
	}

	terr := thumbsTemplate.Execute(w, struct {
		Item  ContentItem
		Items ItemsByTimestamp

		Prev, Next string
	}{
		Item:  item,
		Items: items,

		Prev: prev,
		Next: next,
	})

	if terr != nil {
		log.Println(terr, items)
	}
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
		nextLink, prevLink         string
		prevSiblings, nextSiblings = getSiblings(item, 10)
	)

	if len(nextSiblings) > 0 {
		nextLink = nextSiblings[0].Permalink()
	}
	if len(prevSiblings) > 0 {
		prevLink = prevSiblings[len(prevSiblings)-1].Permalink()
	}

	renderErr := permalinkTemplate.Execute(w, Permalink{
		ContentItem: item,
		Time:        item.Timestamp(),
		Location:    item.Location(),

		NextLink: nextLink,
		PrevLink: prevLink,

		PrevSiblings: prevSiblings,
		NextSiblings: nextSiblings,
	})
	if renderErr != nil {
		log.Fatal(renderErr)
	}
}

func getSiblings(item ContentItem, num int) (prev, next ItemsByTimestamp) {
	var (
		base                  = filepath.Join(filepath.Dir(item.RawPath()), "*")
		siblingPaths, globErr = filepath.Glob(base)

		currentIndex int

		// All siblings
		siblings ItemsByTimestamp

		// Trimmed to num on each side

		prevSibStart, prevSibEnd int
		nextSibStart, nextSibEnd int
	)

	if globErr == nil {
		for _, fn := range siblingPaths {
			sib := ContentItem{
				Type: item.Type,
				Src:  strings.Replace(fn, filepath.Join(config.Config.RawRoot, item.Type+"s")+"/", "", 1),
			}

			if sib.Timestamp().Unix() > 0 {
				siblings = append(siblings, sib)
			}
		}

		sort.Sort(siblings)
	}

	for i, sib := range siblings {
		if sib.RawPath() == item.RawPath() {
			currentIndex = i
			break
		}
	}

	if currentIndex > -1 {
		prevSibStart = currentIndex - num
		prevSibEnd = currentIndex
		if prevSibEnd < 0 {
			prevSibEnd = 0
		}
		if prevSibStart < 0 {
			prevSibStart = 0
		}

		nextSibStart = currentIndex + 1
		nextSibEnd = currentIndex + num
		if nextSibStart >= len(siblings) {
			nextSibStart = len(siblings) - 1
		}
		if nextSibEnd >= len(siblings) {
			nextSibEnd = len(siblings) - 1
		}
	}

	return siblings[prevSibStart:prevSibEnd], siblings[nextSibStart:nextSibEnd]
}
