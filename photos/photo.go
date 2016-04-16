package photos

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/artursapek/artur.co/config"
	"github.com/disintegration/imaging"
	"github.com/julienschmidt/httprouter"
	"github.com/rwcarlsen/goexif/exif"
)

const (
	ExpandDimension = 800
	ThumbDimension  = 100
)

type ContentItem struct {
	Type, Src string
	Caption   template.HTML
}

func (item ContentItem) RawPath() string {
	return filepath.Join(config.Config.RawRoot, item.Type+"s", item.Src)
}

func (item ContentItem) SrcAsHTML() template.HTML {
	return template.HTML(item.Src)
}

func (item ContentItem) ResizedPath() string {
	if item.Type == "video" || item.Type == "audio" {
		return item.RawPath()
	} else {
		return filepath.Join(config.Config.ResizedRoot, item.Type+"s", "expand", item.Src)
	}
}

func (item ContentItem) ThumbPath() string {
	if item.Type == "video" || item.Type == "audio" {
		return item.RawPath()
	} else {
		return filepath.Join(config.Config.ThumbRoot, item.Type+"s", "thumb", item.Src)
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
	case "video":
		out, _ := exec.Command("ffmpeg", "-i", item.RawPath()).CombinedOutput()
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "creation_time") {
				timeStr := strings.Join(strings.Split(line, ":")[1:], ":")
				timeStr = strings.Trim(timeStr, " ")
				t, terr := time.Parse("2006-01-02 15:04:05", timeStr)
				if terr != nil {
					log.Println("Video time parse error: " + terr.Error())
				}
				return t
			}
		}

	case "photo":
		f, ferr := os.Open(item.RawPath())
		defer f.Close()
		if ferr != nil {
			log.Println(ferr)
			return time.Time{}
		} else {
			ex, exerr := exif.Decode(f)
			if exerr != nil {
				log.Println(exerr)
				return time.Time{}
			} else {
				datetime, _ := ex.DateTime()
				return datetime
			}
		}
	}
	return time.Time{}
}

func (item ContentItem) Location() Location {
	switch item.Type {
	case "video":
		out, _ := exec.Command("ffmpeg", "-i", item.RawPath()).CombinedOutput()
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "location ") {
				// example: "+40.7224-073.9375+016.678/"
				locationStr := strings.Split(line, ":")[1]
				locationStr = strings.Trim(locationStr, " /")
				locationParts := regexp.MustCompile("[-+][0-9]+\\.[0-9]+").FindAllString(locationStr, -1)
				lat, _ := strconv.ParseFloat(locationParts[0], 64)
				lng, _ := strconv.ParseFloat(locationParts[1], 64)
				return Location{lat, lng}
			}
		}

	case "photo":
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

func (item ContentItem) ThumbURL() string {
	switch item.Type {
	case "photo":
		return config.Config.ResizedURLPrefix + "assets/thumbs/" + item.Src
	default:
		return ""
	}
}

func (item ContentItem) Resize(maxDimension int, path string) error {
	original, openErr := imaging.Open(item.RawPath())
	if openErr != nil {
		return openErr
	}

	resized := imaging.Fit(original, maxDimension, maxDimension, imaging.Lanczos)

	// Ensure directory structure exists
	dir := filepath.Dir(path)
	mkdirErr := os.MkdirAll(dir, 0700)
	if mkdirErr != nil {
		return mkdirErr
	}

	saveErr := imaging.Save(resized, path)
	if saveErr != nil {
		return saveErr
	}

	return nil
}

func OnTheFlyPhotoResizeHandler(maxDimension int) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		var (
			item = ContentItem{Src: params.ByName("path"), Type: "photo"}
		)

		var path string
		if maxDimension == ExpandDimension {
			path = item.ResizedPath()
		} else if maxDimension == ThumbDimension {
			path = item.ThumbPath()
		} else {
			http.Error(w, "Not found", 404)
			return
		}

		if _, statErr := os.Stat(path); statErr != nil {
			// Not resized before, resize on the fly and cache it
			err := item.Resize(maxDimension, path)
			if err != nil {
				http.Error(w, err.Error(), 500)
			}
		}

		http.ServeFile(w, r, path)
	}
}

// On-the-fly photo resizing that memoizes on disk

func PhotosRedirectHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	http.Redirect(w, r, "/albums", 302)
}
