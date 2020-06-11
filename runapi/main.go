package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	spa "github.com/tmaxmax/serverpropertiesapi"
)

func main() {
	r := mux.NewRouter()
	api := r.PathPrefix("/serverproperties/v1").Subrouter()

	api.HandleFunc("/", spa.GetAllProperties).Methods(http.MethodGet)
	api.HandleFunc("/{key}", spa.GetProperty).Methods(http.MethodGet)
	api.HandleFunc("/meta/", spa.GetMetadata).Methods(http.MethodGet)
	api.HandleFunc("/", spa.MethodNotAllowedHandler(http.MethodGet))
	api.HandleFunc("/{key}", spa.MethodNotAllowedHandler(http.MethodGet))
	api.HandleFunc("/meta/", spa.MethodNotAllowedHandler(http.MethodGet))

	log.Fatalln(http.ListenAndServe(":80", r))
}
