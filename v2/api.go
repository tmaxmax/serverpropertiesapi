package v2

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

type rfc7807 struct {
	errType string
	Title   string `json:"title"`
	Status  int    `json:"status"`
	Detail  string `json:"detail"`
}

var (
	apiNoError               = rfc7807{}
	apiErrorPropertyNotFound = rfc7807{
		errType: "",
		Title:   "Property not found",
		Status:  http.StatusNotFound,
		Detail:  "Could not find property with given name",
	}
	apiErrorUnacceptedType = rfc7807{
		errType: "",
		Title:   "Unaccepted type",
		Status:  http.StatusNotAcceptable,
		Detail:  "API does not support the MIME-types requested",
	}
	apiErrorMethodNotAllowed = rfc7807{
		errType: "",
		Title:   "Method not allowed",
		Status:  http.StatusMethodNotAllowed,
		Detail:  "Endpoint does not support request method",
	}
	apiErrorEndpointInvalidOrMissing = rfc7807{
		errType: "",
		Title:   "Endpoint invalid or missing",
		Status:  http.StatusNotFound,
		Detail:  "Endpoint requested does not exist or URL is invalid",
	}
	apiErrorServerError = rfc7807{
		errType: "",
		Title:   "Server error",
		Status:  http.StatusInternalServerError,
		Detail:  "Something wrong occurred on the server",
	}
	acceptedTypes = map[string]bool{
		"*/*":                      false,
		"application/json":         false,
		"application/problem+json": false,
	}
)

// makeCompleteArray transforms a slice that contains elements which are made of comma separated strings
// into a slice that contains all the elements on a different position
//
// Example: ["a", "b,c", "d,e,f"] -> ["a", "b", "c", "d", "e", "f"]
func makeCompleteArray(a []string) []string {
	if len(a) == 0 {
		return []string{}
	}
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

func obtainMap(keys []string, neg uint8) map[string]bool {
	if len(keys) == 0 {
		return nil
	}
	ret := make(map[string]bool)
	for _, k := range keys {
		allow := k[0] != neg
		if !allow {
			k = k[1:]
		}
		ret[k] = allow
	}
	return ret
}

// checkRequest returns the first API error encountered.
// Returns apiNoError if no error is encountered.
func checkRequest(r *http.Request) rfc7807 {
	accept, i := regexp.MustCompile(`,[ ]?|;[ ]?[qv]=([^,;]+)`).Split(r.Header.Get("Accept"), -1), 0
	for ; i < len(accept); i++ {
		if _, ok := acceptedTypes[accept[i]]; ok {
			break
		}
	}
	if len(accept) != 0 && i == len(accept) {
		return apiErrorUnacceptedType
	}
	return apiNoError
}

func writeError(w http.ResponseWriter, e rfc7807) bool {
	if e == apiNoError {
		return false
	}
	data, _ := json.MarshalIndent(e, "", "  ")
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(e.Status)
	w.Write(data)
	return true
}

// MethodNotAllowedHandler returns a handler for requests that use an unaccepted HTTP method.
func MethodNotAllowedHandler(methods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if writeError(w, checkRequest(r)) {
			return
		}
		w.Header().Set("Allowed", strings.Join(methods, ","))
		writeError(w, apiErrorMethodNotAllowed)
	})
}

// NotFoundHandler returns a handler for requests to invalid endpoints.
func NotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if writeError(w, checkRequest(r)) {
			return
		}
		writeError(w, apiErrorEndpointInvalidOrMissing)
	})
}
