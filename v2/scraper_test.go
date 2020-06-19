package v2

import (
	"encoding/json"
	"testing"
)

func TestServerProperties(t *testing.T) {
	var prop, err = ServerProperties(&filters{
		contains: map[string]bool{
			"max":  true,
			"time": false,
		},
		exactName: "",
		sortMap: map[string]bool{
			"name": true,
		},
		typesMap: map[string]bool{
			"integer": true,
		},
		upcoming: "",
	})
	if err != nil {
		t.Fatal(err)
	}
	text, _ := json.MarshalIndent(prop, "", "  ")
	t.Log(string(text))
}
