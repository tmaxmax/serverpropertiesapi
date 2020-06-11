package serverpropertiesapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type Error struct {
	httpCode int
	Error    string `json:"error"`
	Retry    bool   `json:"retry"`
}

// checkRequest checks the request header if the data necessary to make an API request is available
// It returns an Error slice, which contains all the problems found in the request.
func checkRequest(r *http.Request) []Error {
	var ret []Error

	// Check if the client requests data in a supported format. If no format is provided,
	// then the JSON format is automatically implied.
	accept := r.Header.Get("Accept")
	if accept != "" && accept != "application/json" {
		ret = append(ret, Error{
			httpCode: http.StatusNotImplemented,
			Error:    "501 Not Implemented, API does not support " + accept + " format.",
			Retry:    false,
		})
	}

	return ret
}

// writeErrors writes all the errors found until the point this function is called to the response.
// The response status code will be the httpCode of the first error.
//
// Always check if there are any errors before calling writeErrors!
func writeErrors(errors []Error, w http.ResponseWriter) {
	data, _ := json.Marshal(struct {
		Errors []Error `json:"errors"`
	}{errors})
	w.WriteHeader(errors[0].httpCode)
	w.Write(data)
}

// GetAllProperties is a higher order function that has as parameter the a Property slice,
// which is used to send the information the client requests. It returns a function that
// shall be used as a handler for GET requests of all properties.
func GetAllProperties(p []Property) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		errors := checkRequest(r)
		if len(errors) != 0 {
			writeErrors(errors, w)
			return
		}

		data, _ := json.Marshal(struct {
			Properties []Property `json:"properties"`
		}{p})
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func GetProperty(p []Property) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		errors := checkRequest(r)
		if len(errors) != 0 {
			writeErrors(errors, w)
			return
		}

		// Get the requested key
		pathParams := mux.Vars(r)
		name := pathParams["key"]

		// Find the Property instance with the given name
		var i int
		for i = 0; i < len(p) && p[i].Name != name; i++ {
		}
		if i == len(p) {
			// If it's not found, append a Not Found error to the errors list
			errors = append(errors, Error{
				httpCode: http.StatusNotFound,
				Error:    "404 Not Found, key \"" + name + "\" doesn't exist",
				Retry:    false,
			})
		}

		if len(errors) != 0 {
			writeErrors(errors, w)
			return
		}

		// If this point is reached, there are no errors.
		// Send the Property instance to the client.
		data, _ := json.Marshal(p[i])
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}
