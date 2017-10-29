package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/artursapek/artur.co/blog"
	"github.com/artursapek/artur.co/config"
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
	router.GET("/photos/permalink/*path", photos.PhotoPermalinkHandler)
	router.POST("/photos/*path", photos.PhotoModifyHandler)
	router.GET("/videos/permalink/*path", photos.VideoPermalinkHandler)

	router.GET("/assets/photos/*path", photos.OnTheFlyPhotoResizeHandler(photos.ExpandDimension))
	router.GET("/assets/photos2x/*path", photos.OnTheFlyPhotoResizeHandler(photos.ExpandDimension*2))

	router.GET("/assets/styles/*path", index.AssetsHandler)
	router.GET("/assets/data/*path", index.AssetsHandler)
	router.GET("/assets/scripts/*path", index.AssetsHandler)

	router.GET("/raw/*path", index.RawHandler)

	router.GET("/blog", blog.IndexHandler)
	router.GET("/blog/:id", blog.EntryHandler)

	router.GET("/recipes/bread.html", index.GetSingleAssetHandler("static/bread/bread.html"))
	router.GET("/static/*path", index.GetAssetHandler("static/"))
	router.GET("/artur.asc", index.GetSingleAssetHandler("static/artur.asc"))

	router.GET("/", index.IndexHandler)

	c := &tls.Config{}
	c.Certificates = []tls.Certificate{}

	cert, certErr := tls.LoadX509KeyPair(config.Config.TLSCertFile, config.Config.TLSKeyFile)
	if certErr != nil {
		log.Printf("Omitting %s; error loading: %s", config.Config.TLSCertFile, certErr.Error())
	}

	c.Certificates = append(c.Certificates, cert)

	s := &http.Server{
		Addr:         ":443",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig:    c,
	}

	ss := &http.Server{
		Addr: ":80",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			fmt.Println(time.Now().Format(time.RFC1123Z), req.URL, req.Referer())
			if req.URL.Host == "artur.co" {
				http.Redirect(w, req, "https://artur.co"+req.URL.Path, 302)
			} else {
				router.ServeHTTP(w, req)
			}
		}),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig:    c,
	}

	go ss.ListenAndServe()

	log.Fatal(s.ListenAndServeTLS("", ""))
}
