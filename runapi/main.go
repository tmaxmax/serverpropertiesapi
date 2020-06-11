package main

import (
	"log"
	"net/http"
	"time"

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

	api.Handle("/", cacheClient.Middleware(http.HandlerFunc(spa.GetAllProperties))).Methods(http.MethodGet)
	api.Handle("/{key}", cacheClient.Middleware(http.HandlerFunc(spa.GetProperty))).Methods(http.MethodGet)
	api.Handle("/meta/", cacheClient.Middleware(http.HandlerFunc(spa.GetMetadata))).Methods(http.MethodGet)
	api.HandleFunc("/", spa.MethodNotAllowedHandler(http.MethodGet))
	api.HandleFunc("/{key}", spa.MethodNotAllowedHandler(http.MethodGet))
	api.HandleFunc("/meta/", spa.MethodNotAllowedHandler(http.MethodGet))

	log.Fatalln(http.ListenAndServe(":80", r))
}
