package serverpropertiesapi

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

type Error struct {
	httpCode int
	Error    string `json:"error"`
	Retry    bool   `json:"retry"`
}

var (
	internalServerError = Error{
		httpCode: http.StatusInternalServerError,
		Error:    "500 Internal Server Error, failed to get information",
		Retry:    false,
	}
	metaInformation = map[string]interface{}{
		"minecraftBooleanTypename":  minecraftBooleanTypename,
		"minecraftIntegerTypename":  minecraftIntegerTypename,
		"minecraftStringTypename":   minecraftStringType,
		"propertyDefaultLimitValue": propertyDefaultLimitValue,
	}
)

// makeCompleteArray transforms a slice that contains elements which are made of comma separated strings
// into a slice that contains all the elements on a different position
//
// Example: ["a", "b,c", "d,e,f"] -> ["a", "b", "c", "d", "e", "f"]
func makeCompleteArray(a []string) []string {
	b := make([]string, len(a))
	copy(b, a)
	for i := 0; i < len(b); i++ {
		c := strings.Split(b[i], ",")
		if len(c) > 1 {
			b = append(b[:i], append(c, b[i+1:]...)...)
			i += len(c) - 1
		}
	}
	return b
}

// checkRequest checks the request header if the data necessary to make an API request is available
// It returns an Error slice, which contains all the problems found in the request.
func checkRequest(r *http.Request) []Error {
	var ret []Error

	// Get the accepted formats string
	acceptHeader := r.Header.Get("Accept")

	// Construct the split regex, used to split the header Accept string value
	// into all the requested formats
	split := regexp.MustCompile(`,[ ]?|;[ ]?[qv]=([^,;]+)`)

	// Check if the client requests data in a supported format. If no format is provided,
	// or any format is accepted, application/json is implied.
	accept, i := split.Split(acceptHeader, -1), 0
	for ; i < len(accept); i++ {
		if accept[i] == "*/*" || accept[i] == "application/json" {
			break
		}
	}

	if len(accept) != 0 && i == len(accept) {
		ret = append(ret, Error{
			httpCode: http.StatusNotImplemented,
			Error:    "501 Not Implemented, API does not support " + acceptHeader + " format",
			Retry:    false,
		})
	}

	return ret
}

// writeErrors writes all the errors found until the point this function is called to the response.
// The response status code will be the httpCode of the first error.
//
// Always check if there are any errors before calling writeErrors!
func writeErrors(w http.ResponseWriter, errors ...Error) {
	data, _ := json.MarshalIndent(struct {
		Errors []Error `json:"errors"`
	}{errors}, "", "  ")
	w.WriteHeader(errors[0].httpCode)
	w.Write(data)
}

// GetAllProperties is a higher order function that has as parameter the a Property slice,
// which is used to send the information the client requests. It returns a function that
// shall be used as a handler for GET requests of all properties.
func GetAllProperties(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	errors := checkRequest(r)
	if len(errors) != 0 {
		writeErrors(w, errors...)
		return
	}

	filters := r.URL.Query()
	opt := Options{
		Contains:  makeCompleteArray(filters["contains"]),
		exactName: "",
		Types:     makeCompleteArray(filters["types"]),
		Upcoming:  filters.Get("upcoming"),
	}
	if !opt.Valid() {
		writeErrors(w, Error{
			httpCode: http.StatusBadRequest,
			Error:    "400 Bad Request, options are not valid",
			Retry:    false,
		})
		return
	}
	p, err := ServerProperties(opt)
	if err != nil {
		writeErrors(w, internalServerError)
		return
	}

	data, _ := json.MarshalIndent(struct {
		Options    Options    `json:"options"`
		Properties []Property `json:"properties"`
	}{Options: opt, Properties: p}, "", "  ")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func GetProperty(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	errors := checkRequest(r)
	if len(errors) != 0 {
		writeErrors(w, errors...)
		return
	}

	// Get the requested key
	pathParams := mux.Vars(r)
	name := pathParams["key"]

	// Find the Property instance with the given name
	p, err := ServerProperty(name)
	if err != nil {
		if err.Error() == "not found" {
			writeErrors(w, Error{
				httpCode: http.StatusNotFound,
				Error:    "404 Not Found, key \"" + name + "\" doesn't exist",
				Retry:    false,
			})
			return
		}
		writeErrors(w, internalServerError)
		return
	}

	// If this point is reached, there are no errors.
	// Send the Property instance to the client.
	data, _ := json.MarshalIndent(p, "", "  ")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func GetMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	errors := checkRequest(r)
	if len(errors) != 0 {
		writeErrors(w, errors...)
		return
	}

	data, _ := json.MarshalIndent(struct {
		Meta map[string]interface{} `json:"meta"`
	}{metaInformation}, "", "  ")
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
		writeErrors(w, errors...)
	}
}
