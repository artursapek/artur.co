package photos

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"launchpad.net/goyaml"

	"github.com/julienschmidt/httprouter"
)

var (
	albumNotFoundError = errors.New("Album not found")
)

type Album struct {
	Title   string
	Date    string
	Content []ContentItem
}

const (
	dateInputFormat  = "1/2/06"
	dateOutputFormat = "Jan 2, 2006"
)

func loadAlbum(slug string) (a Album, err error) {
	fp := filepath.Join("content", "photos", "albums", slug+".yml")

	albumContent, getContentErr := ioutil.ReadFile(fp)
	if getContentErr != nil {
		return a, albumNotFoundError
	} else {
		parseErr := goyaml.Unmarshal(albumContent, &a)
		fmt.Println(a, parseErr)
		if parseErr != nil {
			return a, errors.New("Failed to parse album: " + parseErr.Error())
		} else {
			date, dateParseErr := time.Parse(dateInputFormat, a.Date)

			if dateParseErr != nil {
				return a, errors.New("Failed to parse date: " + a.Date + "\n" + dateParseErr.Error())
			}

			a.Date = date.Format(dateOutputFormat)

			return a, nil
		}
	}
}

var (
	albumTemplate *template.Template
)

func init() {
	var tParseErr error
	albumTemplate, tParseErr = template.ParseFiles("templates/photos/album.html")
	if tParseErr != nil {
		log.Fatal(tParseErr)
	}
}

func AlbumHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	a, getErr := loadAlbum(p.ByName("slug"))
	if getErr != nil {
		switch getErr {
		case albumNotFoundError:
			http.Error(w, "album not found", 404)
		default:
			http.Error(w, "internal error", 500)
		}
	} else {
		renderErr := albumTemplate.Execute(w, a)
		if renderErr != nil {
			log.Fatal(renderErr)
		}
	}
}
