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
	router.GET("/photos", photos.PhotosRedirectHandler)
	router.GET("/photos/*path", photos.PhotoHandler)
	router.GET("/assets/*path", index.AssetsHandler)
	router.GET("/", index.IndexHandler)

	log.Fatal(http.ListenAndServe(":8080", router))
}
