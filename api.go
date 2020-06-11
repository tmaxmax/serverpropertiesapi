package serverpropertiesapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Error struct {
	httpCode int
	Error    string `json:"error"`
	Retry    bool   `json:"retry"`
}

var internalServerError = Error{
	httpCode: http.StatusInternalServerError,
	Error:    "500 Internal Server Error, failed to get information",
	Retry:    false,
}

// checkRequest checks the request header if the data necessary to make an API request is available
// It returns an Error slice, which contains all the problems found in the request.
func checkRequest(r *http.Request) []Error {
	var ret []Error

	// Check if the client requests data in a supported format. If no format is provided,
	// or any format is accepted, application/json is implied.
	accept := r.Header.Get("Accept")
	if accept != "" && accept != "*/*" && accept != "application/json" {
		ret = append(ret, Error{
			httpCode: http.StatusNotImplemented,
			Error:    "501 Not Implemented, API does not support " + accept + " format",
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
func GetAllProperties(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	errors := checkRequest(r)
	if len(errors) != 0 {
		writeErrors(errors, w)
		return
	}

	filters := r.URL.Query()
	opt := Options{
		Contains:  filters.Get("contains"),
		exactName: "",
		Type:      filters.Get("type"),
		Upcoming:  filters.Get("upcoming"),
	}
	p, err := ServerProperties(opt)
	if err != nil {
		writeErrors([]Error{internalServerError}, w)
		return
	}
	if len(p) == 0 {
		writeErrors([]Error{{
			httpCode: http.StatusNotFound,
			Error:    "404 Not Found, there are no properties satisfying your filters",
			Retry:    false,
		}}, w)
		return
	}

	data, _ := json.Marshal(struct {
		Options    Options    `json:"options"`
		Properties []Property `json:"properties"`
	}{Options: opt, Properties: p})
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func GetProperty(w http.ResponseWriter, r *http.Request) {
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
	p, err := ServerProperty(name)
	if err != nil {
		if err.Error() == "not found" {
			writeErrors([]Error{{
				httpCode: http.StatusNotFound,
				Error:    "404 Not Found, key \"" + name + "\" doesn't exist",
				Retry:    false,
			}}, w)
			return
		}
		writeErrors([]Error{internalServerError}, w)
		return
	}

	// If this point is reached, there are no errors.
	// Send the Property instance to the client.
	data, _ := json.Marshal(p)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func GetMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	errors := checkRequest(r)
	if len(errors) != 0 {
		writeErrors(errors, w)
		return
	}

	data, _ := json.Marshal(struct {
		Meta map[string]interface{} `json:"meta"`
	}{Meta: map[string]interface{}{
		"minecraftBooleanTypename":  minecraftBooleanTypename,
		"minecraftIntegerTypename":  minecraftIntegerTypename,
		"minecraftStringTypename":   minecraftStringType,
		"propertyDefaultLimitValue": propertyDefaultLimitValue,
	}})
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func MethodNotAllowedHandler(allowed ...string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		allowedMethods := strings.Join(allowed, ", ")
		errors := checkRequest(r)
		errors = append(errors, Error{
			httpCode: http.StatusMethodNotAllowed,
			Error:    "405 Status Method Not Allowed, can't use " + r.Method,
			Retry:    false,
		})
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Allowed", allowedMethods)
		writeErrors(errors, w)
	}
}
