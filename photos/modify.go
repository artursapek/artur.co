package photos

import (
	"net/http"
	"os"
	"os/exec"

	"github.com/julienschmidt/httprouter"
)

func PhotoModifyHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var (
		item = ContentItem{Src: params.ByName("path"), Type: "photo"}
	)

	if _, statErr := os.Stat(item.ResizedPath()); statErr != nil {
		http.Error(w, "Photo doesn't exist", 404)
	} else {
		redirect := r.PostFormValue("redirect")
		if redirect == "" {
			redirect = r.Referer()
		}

		if rotate := r.PostFormValue("rotate"); rotate != "" {
			rotateErr := exec.Command("convert", item.RawPath(), "-rotate", rotate, item.RawPath()).Run()
			if rotateErr != nil {
				http.Error(w, "Error while rotating: "+rotateErr.Error(), 500)
				return
			}
			removeErr := exec.Command("rm", item.ResizedPath()).Run()
			if removeErr != nil {
				http.Error(w, "Error while busting cache: "+removeErr.Error(), 500)
				return
			}

			// Prime the cache
			item.Resize(ExpandDimension)
			item.Resize(ThumbDimension)

			http.Redirect(w, r, redirect, 301)
		} else {
			http.Error(w, "No actions to take", 400)
		}
	}
}
