package main

import (
	"log"
	"net/http"
	"time"

	"github.com/NYTimes/gziphandler"

	cache "github.com/victorspringer/http-cache"

	"github.com/victorspringer/http-cache/adapter/memory"

	"github.com/gorilla/mux"
	spa "github.com/tmaxmax/serverpropertiesapi"
)

func main() {
	memcached, err := memory.NewAdapter(memory.AdapterWithAlgorithm(memory.LRU), memory.AdapterWithCapacity(10000000))
	if err != nil {
		log.Fatalln(err)
	}
	cacheClient, err := cache.NewClient(cache.ClientWithAdapter(memcached), cache.ClientWithTTL(24*time.Hour))
	if err != nil {
		log.Fatalln(err)
	}

	r := mux.NewRouter()
	api := r.PathPrefix("/serverproperties/v1").Subrouter()

	api.Handle("/", cacheClient.Middleware(gziphandler.GzipHandler(http.HandlerFunc(spa.GetAllProperties)))).Methods(http.MethodGet)
	api.Handle("/{key}", cacheClient.Middleware(gziphandler.GzipHandler(http.HandlerFunc(spa.GetProperty)))).Methods(http.MethodGet)
	api.Handle("/meta/", cacheClient.Middleware(gziphandler.GzipHandler(http.HandlerFunc(spa.GetMetadata)))).Methods(http.MethodGet)
	api.Handle("/", gziphandler.GzipHandler(http.HandlerFunc(spa.MethodNotAllowedHandler(http.MethodGet))))
	api.Handle("/{key}", gziphandler.GzipHandler(http.HandlerFunc(spa.MethodNotAllowedHandler(http.MethodGet))))
	api.Handle("/meta/", gziphandler.GzipHandler(http.HandlerFunc(spa.MethodNotAllowedHandler(http.MethodGet))))

	log.Fatalln(http.ListenAndServe(":80", r))
}
