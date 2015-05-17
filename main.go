package main

import (
	"log"
	"net/http"

	"github.com/artursapek/artur.co/index"
	"github.com/artursapek/artur.co/photos"
	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()

	router.GET("/albums", photos.AlbumsIndexHandler)
	router.GET("/albums/:slug", photos.AlbumHandler)
	// Old URL used to be http://artur.co/photos
	router.GET("/photos", photos.PhotosRedirectHandler)
	router.GET("/photos/*path", photos.PhotoPermalinkHandler)
	router.GET("/videos/*path", photos.VideoPermalinkHandler)

	router.GET("/assets/photos/*path", photos.PhotoHandler)

	router.GET("/assets/styles/*path", index.AssetsHandler)
	router.GET("/assets/data/*path", index.AssetsHandler)
	router.GET("/assets/scripts/*path", index.AssetsHandler)

	router.GET("/", index.IndexHandler)

	log.Fatal(http.ListenAndServe(":8080", router))
}
