package photos

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/artursapek/artur.co/config"
	"github.com/disintegration/imaging"
	"github.com/julienschmidt/httprouter"
	"github.com/rwcarlsen/goexif/exif"
)

const (
	ExpandDimension = 800
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

func (item ContentItem) ResizedPath(maxDimension int) string {
	if item.Type == "video" || item.Type == "audio" {
		return item.RawPath()
	} else {
		return filepath.Join(config.Config.ResizedRoot, "photos", fmt.Sprintf("%d", maxDimension), item.Src)
	}
}

func (item ContentItem) BaseFilename() string {
	return filepath.Base(item.Src)
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

func (item ContentItem) Resized2xURL() string {
	switch item.Type {
	case "video":
		return item.RawURL()
	case "photo":
		return config.Config.ResizedURLPrefix + "assets/photos2x/" + item.Src
	default:
		return ""
	}
}

func (item ContentItem) Year() string {
	if item.Src == "" {
		return ""
	}
	return strings.Split(item.Src, "/")[0]
}

func (item ContentItem) Month() string {
	if item.Src == "" {
		return ""
	}
	return strings.Split(item.Src, "/")[1]
}

func (item ContentItem) Base() string {
	return filepath.Base(item.Src)
}

var ResizeMutex = new(sync.Mutex)

func (item ContentItem) Resize(maxDimension int, path string, filter imaging.ResampleFilter) error {
	ResizeMutex.Lock()
	defer ResizeMutex.Unlock()
	original, openErr := imaging.Open(item.RawPath())
	if openErr != nil {
		return openErr
	}

	// Get orientation
	f, ferr := os.Open(item.RawPath())
	if ferr != nil {
		log.Println(ferr)
		return ferr
	}
	ex, exErr := exif.Decode(f)
	f.Close()
	if exErr != nil {
		log.Println(exErr)
		return exErr
	}

	// Rotate if necessary
	var orientation string

	orientationTag, oerr := ex.Get(exif.Orientation)
	if oerr != nil {
		fmt.Println("Couldnt detect orientation for " + path)
		orientation = "1"
	} else {
		orientation = orientationTag.String()
	}

	switch orientation {
	case "3":
		original = imaging.Rotate180(original)
	case "6":
		original = imaging.Rotate270(original)
	case "8":
		original = imaging.Rotate90(original)
	}

	resized := imaging.Fit(original, maxDimension, maxDimension, filter)

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

// On-the-fly photo resizing that memoizes on disk
func OnTheFlyPhotoResizeHandler(maxDimension int) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		var (
			item   = ContentItem{Src: params.ByName("path"), Type: "photo"}
			path   = item.ResizedPath(maxDimension)
			filter = imaging.Lanczos
		)

		if _, statErr := os.Stat(path); statErr != nil {
			// Not resized before, resize on the fly and cache it
			err := item.Resize(maxDimension, path, filter)
			if err != nil {
				http.Error(w, err.Error(), 500)
			}
		}

		http.ServeFile(w, r, path)
	}
}
