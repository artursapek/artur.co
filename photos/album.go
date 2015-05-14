package photos

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"launchpad.net/goyaml"

	"github.com/julienschmidt/httprouter"
)

var (
	albumNotFoundError = errors.New("Album not found")
)

type Album struct {
	Slug    string
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
			a.Slug = slug

			return a, nil
		}
	}
}

var (
	albumTemplate, albumIndexTemplate *template.Template
)

func init() {
	var showParseErr, indexParseErr error
	albumTemplate, showParseErr = template.ParseFiles("templates/photos/album.html")
	if indexParseErr != nil {
		log.Fatal(showParseErr)
	}
	albumIndexTemplate, indexParseErr = template.ParseFiles("templates/photos/index.html")
	if indexParseErr != nil {
		log.Fatal(indexParseErr)
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

func AlbumsIndexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var albums []Album

	albumFiles, globErr := filepath.Glob(filepath.Join("content", "photos", "albums", "*.yml"))

	for _, fp := range albumFiles {
		slug := strings.Split(filepath.Base(fp), ".")[0]
		album, loadErr := loadAlbum(slug)
		if loadErr != nil {
			log.Fatal(loadErr)
		} else {
			albums = append(albums, album)
		}
	}

	if globErr != nil {
		log.Fatal(globErr)
	}
	renderErr := albumIndexTemplate.Execute(w, albums)
	if renderErr != nil {
		log.Fatal(renderErr)
	}
}
