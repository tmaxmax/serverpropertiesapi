package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/tmaxmax/serverpropertiesapi"
)

func main() {
	prop, err := serverpropertiesapi.ServerProperties()
	if err != nil {
		log.Fatalln("Couldn't get properties data, API cannot start. Error:", err)
	}

	r := mux.NewRouter()
	api := r.PathPrefix("/serverproperties/v1").Subrouter()

	api.HandleFunc("/", serverpropertiesapi.GetAllProperties(prop)).Methods(http.MethodGet)
	api.HandleFunc("/{key}", serverpropertiesapi.GetProperty(prop)).Methods(http.MethodGet)
	api.HandleFunc("/meta/", serverpropertiesapi.GetMetadata).Methods(http.MethodGet)

	api.HandleFunc("/", serverpropertiesapi.MethodNotAllowedHandler(http.MethodGet))
	api.HandleFunc("/{key}", serverpropertiesapi.MethodNotAllowedHandler(http.MethodGet))
	api.HandleFunc("/meta/", serverpropertiesapi.MethodNotAllowedHandler(http.MethodGet))

	log.Fatalln(http.ListenAndServe(":80", r))
}
