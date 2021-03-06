package photos

import (
	"log"
	"net/http"
	"os"

	"github.com/disintegration/imaging"
	"github.com/julienschmidt/httprouter"
)

func PhotoModifyHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var (
		item = ContentItem{Src: params.ByName("path"), Type: "photo"}
	)

	_, statErr1 := os.Stat(item.ResizedPath(ExpandDimension))
	_, statErr2 := os.Stat(item.ResizedPath(ExpandDimension * 2))

	if statErr1 != nil && statErr2 != nil {
		http.Error(w, "Photo doesn't exist", 404)
	} else {
		redirect := r.PostFormValue("redirect")
		if redirect == "" {
			redirect = r.Referer()
		}

		dimens := []int{
			ExpandDimension, ExpandDimension * 2,
		}

		for _, dimen := range dimens {

			img, openErr := imaging.Open(item.ResizedPath(dimen))

			if openErr != nil {
				log.Println(openErr)
				continue
			}

			if rotate := r.PostFormValue("rotate"); rotate != "" {
				switch rotate {
				case "90":
					img = imaging.Rotate90(img)
				case "180":
					img = imaging.Rotate180(img)
				case "270":
					img = imaging.Rotate270(img)
				default:
					http.Error(w, "Invalid value "+rotate, 400)
				}

				saveErr := imaging.Save(img, item.ResizedPath(dimen))

				if saveErr != nil {
					http.Error(w, saveErr.Error(), 500)
				}
			} else {
				http.Error(w, "No actions to take", 400)
			}
		}

		http.Redirect(w, r, redirect, 301)
	}
}
