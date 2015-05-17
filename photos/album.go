package photos

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sort"
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

	unix int64
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
		if parseErr != nil {
			return a, errors.New("Failed to parse album: " + parseErr.Error())
		} else {
			date, dateParseErr := time.Parse(dateInputFormat, a.Date)

			if dateParseErr != nil {
				return a, errors.New("Failed to parse date: " + a.Date + "\n" + dateParseErr.Error())
			}

			a.Date = date.Format(dateOutputFormat)
			a.unix = date.Unix()
			a.Slug = slug

			return a, nil
		}
	}
}

func allAlbums() Albums {
	var albums Albums

	albumFiles, globErr := filepath.Glob(filepath.Join("content", "photos", "albums", "*.yml"))

	if globErr != nil {
		log.Fatal(globErr)
	}

	for _, fp := range albumFiles {
		slug := strings.Split(filepath.Base(fp), ".")[0]
		album, loadErr := loadAlbum(slug)
		if loadErr != nil {
			log.Fatal(loadErr)
		} else {
			albums = append(albums, album)
		}
	}

	sort.Sort(albums)

	return albums
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

type albumHandlerContext struct {
	Current, Prev, Next Album
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
		albums := allAlbums()
		var thisIndex int
		for i, album := range albums {
			if album.Slug == a.Slug {
				thisIndex = i
				break
			}
		}
		var prev, next Album

		if thisIndex > 0 {
			prev = albums[thisIndex-1]
		}
		if thisIndex < len(albums)-1 {
			next = albums[thisIndex+1]
		}

		context := albumHandlerContext{
			Current: a,
			Prev:    prev,
			Next:    next,
		}

		renderErr := albumTemplate.Execute(w, context)
		if renderErr != nil {
			log.Fatal(renderErr)
		}
	}
}

func AlbumsIndexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	albums := allAlbums()
	renderErr := albumIndexTemplate.Execute(w, albums)
	if renderErr != nil {
		log.Fatal(renderErr)
	}
}

type Albums []Album

func (albums Albums) Len() int {
	return len(albums)
}

func (albums Albums) Less(i, j int) bool {
	return albums[i].unix > albums[j].unix
}

func (albums Albums) Swap(i, j int) {
	albums[i], albums[j] = albums[j], albums[i]
}
