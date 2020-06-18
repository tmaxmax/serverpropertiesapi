package v2

import (
	"encoding/json"
	"testing"

	"golang.org/x/text/language"
)

func TestServerProperties(t *testing.T) {
	var prop, err = ServerProperties(&Filters{
		contains:  nil,
		exactName: "",
		language:  language.LatinAmericanSpanish,
		sort:      nil,
		types:     nil,
		upcoming:  "",
	})
	if err != nil {
		t.Fatal(err)
	}
	text, _ := json.MarshalIndent(prop, "", "  ")
	t.Log(string(text))
}
