package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/NYTimes/gziphandler"

	cache "github.com/victorspringer/http-cache"

	"github.com/victorspringer/http-cache/adapter/memory"

	"github.com/gorilla/mux"
	spa "github.com/tmaxmax/serverpropertiesapi"
)

func main() {
	memcached, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000000),
	)
	if err != nil {
		log.Fatalln(err)
	}
	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(24*time.Hour),
		cache.ClientWithRefreshKey("opn"),
	)
	if err != nil {
		log.Fatalln(err)
	}

	// main router
	r := mux.NewRouter()
	// server.properties API sub router
	sprop := r.PathPrefix("/v1/serverproperties").Subrouter()

	sprop.Handle("", cacheClient.Middleware(gziphandler.GzipHandler(http.HandlerFunc(spa.GetAllProperties)))).Methods(http.MethodGet)
	sprop.Handle("/{key}", cacheClient.Middleware(gziphandler.GzipHandler(http.HandlerFunc(spa.GetProperty)))).Methods(http.MethodGet)
	sprop.Handle("/meta/", cacheClient.Middleware(gziphandler.GzipHandler(http.HandlerFunc(spa.GetMetadata)))).Methods(http.MethodGet)

	sprop.Handle("", gziphandler.GzipHandler(http.HandlerFunc(spa.MethodNotAllowedHandler(http.MethodGet))))
	sprop.Handle("/{key}", gziphandler.GzipHandler(http.HandlerFunc(spa.MethodNotAllowedHandler(http.MethodGet))))
	sprop.Handle("/meta/", gziphandler.GzipHandler(http.HandlerFunc(spa.MethodNotAllowedHandler(http.MethodGet))))

	log.Fatalln(http.ListenAndServeTLS(":443", os.Getenv("CERTFILE"), os.Getenv("KEYFILE"), r))
}
