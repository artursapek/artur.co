package photos

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/artursapek/artur.co/config"
	"github.com/disintegration/imaging"
	"github.com/julienschmidt/httprouter"
	"github.com/rwcarlsen/goexif/exif"
)

const (
	maxDimension = 800
)

type ContentItem struct {
	Type, Src string
	Caption   template.HTML
}

func (item ContentItem) RawPath() string {
	return filepath.Join(config.Config.RawRoot, item.Src)
}

func (item ContentItem) ResizedPath() string {
	if item.Type == "video" {
		return item.RawPath()
	} else {
		return filepath.Join(config.Config.ResizedRoot, item.Src)
	}
}

func (item ContentItem) BaseFilename() string {
	return filepath.Base(item.ResizedPath())
}

func (item ContentItem) RawURL() string {
	return config.Config.RawURLPrefix + item.Type + "s/" + item.Src
}

func (item ContentItem) Timestamp() time.Time {
	switch item.Type {
	case "video", "photo":
		f, ferr := os.Open(item.RawPath())
		defer f.Close()
		if ferr != nil {
			log.Println(ferr)
			return time.Now()
		} else {
			ex, exerr := exif.Decode(f)
			if exerr != nil {
				log.Println(exerr)
				return time.Now()
			} else {
				datetime, _ := ex.DateTime()
				return datetime
			}
		}
	}
	return time.Now()
}

func (item ContentItem) Location() Location {
	switch item.Type {
	case "video", "photo":
		f, ferr := os.Open(item.RawPath())
		defer f.Close()
		if ferr != nil {
			log.Println(ferr)
			return Location{}
		} else {
			ex, exerr := exif.Decode(f)
			if exerr != nil {
				return Location{}
			} else {
				lat, lng, err := ex.LatLong()
				if err == nil {
					return Location{lat, lng}
				} else {
					return Location{}
				}
			}
		}
	}
	return Location{}
}

func (item ContentItem) ResizedURL() string {
	switch item.Type {
	case "video":
		return item.RawURL()
	case "photo":
		return config.Config.ResizedURLPrefix + "assets/photos/" + item.Src
	default:
		return ""
	}
}

func (item ContentItem) Resize() error {
	original, openErr := imaging.Open(item.RawPath())
	if openErr != nil {
		return openErr
	}

	resized := imaging.Fit(original, maxDimension, maxDimension, imaging.Lanczos)

	// Ensure directory structure exists
	dir := filepath.Dir(item.ResizedPath())
	mkdirErr := os.MkdirAll(dir, 0700)
	if mkdirErr != nil {
		return mkdirErr
	}

	saveErr := imaging.Save(resized, item.ResizedPath())
	if saveErr != nil {
		return saveErr
	}

	return nil
}

// On-the-fly photo resizing that memoizes on disk
func PhotoHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var (
		item = ContentItem{Src: params.ByName("path"), Type: "photo"}
	)

	if _, statErr := os.Stat(item.ResizedPath()); statErr != nil {
		// Not resized before, resize on the fly and cache it
		err := item.Resize()
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	}

	http.ServeFile(w, r, item.ResizedPath())
}

func PhotosRedirectHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	http.Redirect(w, r, "/albums", 302)
}
