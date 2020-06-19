package v2

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	apiErrorInvalidTypeQuery = rfc7807{
		errType: "",
		Title:   "Invalid type query",
		Status:  http.StatusBadRequest,
		Detail:  "The type query filter contains mixed allowed and disallowed values",
	}
)

func GetServerProperties() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if writeError(w, checkRequest(r)) {
			return
		}

		filtersQuery := r.URL.Query()
		f := &filters{
			contains:  obtainMap(makeCompleteArray(filtersQuery["contains"]), '!'),
			exactName: "",
			sortMap:   obtainMap(makeCompleteArray(filtersQuery["sort"]), '-'),
			typesMap:  obtainMap(makeCompleteArray(filtersQuery["types"]), '!'),
			upcoming:  filtersQuery.Get("upcoming"),
		}
		if !validateTypes(f) {
			writeError(w, apiErrorInvalidTypeQuery)
			return
		}

		prop, err := ServerProperties(f)
		if err != nil {
			writeError(w, apiErrorServerError)
			return
		}

		data, _ := json.MarshalIndent(prop, "", "  ")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})
}

func GetServerProperty() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if writeError(w, checkRequest(r)) {
			return
		}

		p, err := ServerProperties(&filters{
			exactName: mux.Vars(r)["key"],
		})
		if err != nil {
			writeError(w, apiErrorServerError)
			return
		}
		if len(p) == 0 {
			writeError(w, apiErrorPropertyNotFound)
			return
		}
		data, _ := json.MarshalIndent(p[0], "", "  ")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})
}

func GetMetadata() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if writeError(w, checkRequest(r)) {
			return
		}

		data, err := json.MarshalIndent(metadata, "", "  ")
		if err != nil {
			writeError(w, apiErrorServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})
}
