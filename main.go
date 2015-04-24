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

	router.GET("/photos/photo/*path", photos.PhotoHandler)
	router.GET("/photos/album/:slug", photos.AlbumHandler)
	router.GET("/assets/*path", index.AssetsHandler)
	router.GET("/", index.IndexHandler)

	log.Fatal(http.ListenAndServe(":8080", router))
}
