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

func rawPhotoPath(path string) string {
	return filepath.Join(config.Config.RawRoot, path)
}

func resizedPhotoPath(path string) string {
	return filepath.Join(config.Config.ResizedRoot, path)
}

// On-the-fly photo resizing that memoizes on disk
func PhotoHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var (
		rawPath     = rawPhotoPath(p.ByName("path"))
		resizedPath = resizedPhotoPath(p.ByName("path"))
	)

	if _, statErr := os.Stat(resizedPath); statErr != nil {
		// Not resized before, resize on the fly and cache it

		original, openErr := imaging.Open(rawPath)
		if openErr != nil {
			log.Fatal(openErr)
		}

		resized := imaging.Fit(original, maxDimension, maxDimension, imaging.BSpline)

		imaging.Save(resized, resizedPath)
	}

	http.ServeFile(w, r, resizedPath)
}
