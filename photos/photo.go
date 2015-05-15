package photos

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/artursapek/artur.co/config"
	"github.com/disintegration/imaging"
	"github.com/julienschmidt/httprouter"
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

func (item ContentItem) ResizedURL() string {
	if item.Type == "video" {
		return item.RawURL()
	} else {
		return config.Config.ResizedURLPrefix + item.Type + "s/" + item.Src
	}
}

// On-the-fly photo resizing that memoizes on disk
func PhotoHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var (
		item = ContentItem{Src: params.ByName("path")}
	)

	if _, statErr := os.Stat(item.ResizedPath()); statErr != nil {
		// Not resized before, resize on the fly and cache it

		original, openErr := imaging.Open(item.RawPath())
		if openErr != nil {
			http.Error(w, "photo not found", 404)
			return
		}

		resized := imaging.Fit(original, maxDimension, maxDimension, imaging.Lanczos)

		// Ensure directory structure exists
		dir := filepath.Dir(item.ResizedPath())
		mkdirErr := os.MkdirAll(dir, 0700)
		if mkdirErr != nil {
			log.Fatal(mkdirErr)
		}

		saveErr := imaging.Save(resized, item.ResizedPath())
		if saveErr != nil {
			log.Fatal(saveErr)
		}
	}

	http.ServeFile(w, r, item.ResizedPath())
}

func PhotosRedirectHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	http.Redirect(w, r, "/albums", 302)
}
