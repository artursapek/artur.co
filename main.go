package main

import (
	"log"
	"net/http"

	"github.com/artursapek/artur.co/index"
	"github.com/artursapek/artur.co/photos"
	"github.com/julienschmidt/httprouter"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	router := httprouter.New()

	router.GET("/albums", photos.AlbumsIndexHandler)
	router.GET("/albums/:slug", photos.AlbumHandler)
	// Old URL used to be http://artur.co/photos
	router.GET("/photos", photos.PhotosIndexHandler)
	router.GET("/photos/month/:year/:month", photos.PhotosYearMonthHandler)
	router.GET("/photos/permalink/*path", photos.PhotoPermalinkHandler)
	router.POST("/photos/*path", photos.PhotoModifyHandler)
	router.GET("/videos/permalink/*path", photos.VideoPermalinkHandler)

	router.GET("/assets/photos/*path", photos.OnTheFlyPhotoResizeHandler(photos.ExpandDimension))
	router.GET("/assets/thumbs/*path", photos.OnTheFlyPhotoResizeHandler(photos.ThumbDimension))

	router.GET("/assets/styles/*path", index.AssetsHandler)
	router.GET("/assets/data/*path", index.AssetsHandler)
	router.GET("/assets/scripts/*path", index.AssetsHandler)

	router.GET("/", index.IndexHandler)

	log.Fatal(http.ListenAndServe(":8081", router))
}
