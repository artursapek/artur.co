package photos

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"launchpad.net/goyaml"

	"github.com/artursapek/artur.co/config"
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
	dateOutputFormat = "January 2, 2006"
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
	albumTemplate,
	albumIndexTemplate,
	photosIndexTemplate,
	photosMonthTemplate *template.Template
)

func init() {
	var showParseErr, indexParseErr error
	albumTemplate, showParseErr = template.ParseFiles("templates/albums/album.html")
	if indexParseErr != nil {
		log.Fatal(showParseErr)
	}
	albumIndexTemplate, indexParseErr = template.ParseFiles("templates/albums/index.html")
	if indexParseErr != nil {
		log.Fatal(indexParseErr)
	}

	photosIndexTemplate, indexParseErr = template.ParseFiles("templates/photos/index.html")
	if indexParseErr != nil {
		log.Fatal(indexParseErr)
	}

	photosMonthTemplate, indexParseErr = template.ParseFiles("templates/photos/month.html")
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
			next = albums[thisIndex-1]
		}
		if thisIndex < len(albums)-1 {
			prev = albums[thisIndex+1]
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

type RawDeviceStats struct {
	NumFiles            int32
	Blocks, Used, Usage int32
}

var dfOutputMatcher = regexp.MustCompile("([^\\ ]+)\\ +([0-9]+)\\ +([0-9]+)\\ +([0-9]+)\\ +([0-9]+)")

func getRawDeviceStats() RawDeviceStats {
	findOutput, findErr := exec.Command("find", config.Config.RawRoot, "-type", "f").Output()
	if findErr != nil {
		return RawDeviceStats{}
	}
	numFiles := len(strings.Split(string(findOutput), "\n"))

	dfOutput, err := exec.Command("df").Output()
	if err != nil {
		return RawDeviceStats{}
	} else {
		lines := strings.Split(string(dfOutput), "\n")
		for _, line := range lines {
			if strings.Contains(line, config.Config.RawDeviceName) {
				parts := dfOutputMatcher.FindStringSubmatch(line)
				fmt.Println(line)
				fmt.Println(parts)
				fmt.Println(len(parts))
				blocks, _ := strconv.ParseInt(parts[2], 10, 32)
				used, _ := strconv.ParseInt(parts[3], 10, 32)
				capacity, _ := strconv.ParseInt(parts[5], 10, 32)
				fmt.Println(parts)
				return RawDeviceStats{
					NumFiles: int32(numFiles),
					Blocks:   int32(blocks),
					Used:     int32(used),
					Usage:    int32(capacity),
				}
			}
		}
	}
	return RawDeviceStats{}
}

type albumIndexData struct {
	Albums
	RawDeviceStats
}

func AlbumsIndexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	albums := allAlbums()
	stats := getRawDeviceStats()
	fmt.Println(stats)
	renderErr := albumIndexTemplate.Execute(w, albumIndexData{albums, stats})
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
