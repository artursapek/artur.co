package photos

import (
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

type Photo struct {
	Src, Caption string
}

func (p Photo) RawPath() string {
	return filepath.Join(config.Config.RawRoot, p.Src)
}

func (p Photo) ResizedPath() string {
	return filepath.Join(config.Config.ResizedRoot, p.Src)
}

func (p Photo) RawURL() string {
	return config.Config.RawURLPrefix + p.Src
}

func (p Photo) ResizedURL() string {
	return config.Config.ResizedURLPrefix + p.Src
}

// On-the-fly photo resizing that memoizes on disk
func PhotoHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var (
		p = Photo{Src: params.ByName("path")}
	)

	if _, statErr := os.Stat(p.ResizedPath()); statErr != nil {
		// Not resized before, resize on the fly and cache it

		original, openErr := imaging.Open(p.RawPath())
		if openErr != nil {
			http.Error(w, "photo not found", 404)
			return
		}

		resized := imaging.Fit(original, maxDimension, maxDimension, imaging.Lanczos)

		// Ensure directory structure exists
		dir := filepath.Dir(p.ResizedPath())
		mkdirErr := os.MkdirAll(dir, 0700)
		if mkdirErr != nil {
			log.Fatal(mkdirErr)
		}

		saveErr := imaging.Save(resized, p.ResizedPath())
		if saveErr != nil {
			log.Fatal(saveErr)
		}
	}

	http.ServeFile(w, r, p.ResizedPath())
}
