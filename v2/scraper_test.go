package v2

import (
	"encoding/json"
	"testing"
)

func TestServerProperties(t *testing.T) {
	filters := []filters{
		{
			contains:  nil,
			exactName: "difficulty",
			sortMap:   nil,
			typesMap:  nil,
			upcoming:  "",
		},
		{
			contains:  nil,
			exactName: "",
			sortMap:   nil,
			typesMap:  nil,
			upcoming:  "",
		},
		{
			contains: map[string]bool{
				"allow": true,
			},
			exactName: "",
			sortMap:   nil,
			typesMap:  nil,
			upcoming:  "",
		},
		{
			contains: map[string]bool{
				"max":   false,
				"allow": false,
			},
			exactName: "",
			sortMap: map[string]bool{
				"name": true,
			},
			typesMap: nil,
			upcoming: "",
		},
		{
			contains:  nil,
			exactName: "",
			sortMap: map[string]bool{
				"name": false,
			},
			typesMap: map[string]bool{
				minecraftStringTypename: true,
			},
		},
	}
	for i, f := range filters {
		t.Logf("\nTest %d...\n", i)
		validateTypes(&f)
		prop, err := ServerProperties(&f)
		if err != nil {
			t.Errorf("\nTest %d failed, error: %v\n", i, err)
		}
		if f.exactName != "" && (len(prop) > 1 || prop[0].Name != f.exactName) {
			t.Errorf("exactName filtering failed")
		}
		for j, p := range prop {
			// If any of these fields are empty, it means that gathering the information from the website failed.
			// This could be caused by a redesign, for example.
			if p.Name == "" || p.Type == "" || p.Values.Default == "" || p.Description == "" {
				t.Errorf("\tFailed on property %d, one or more fields are empty. Check if website format hasn't changed!\nProperty instance: %+v\n", j, p)
			}
			// This is to assure that the limits have the default value if they are not mentioned in the documentation.
			if p.Values.Min == propertyDefaultLimitValue && p.Values.Max != p.Values.Min {
				t.Errorf("\tFailed on property %d, invalid Min and Max.\np.Values.Min = %d, p.Values.Max = %d\n", j, p.Values.Min, p.Values.Max)
			}
			// If the feature isn't going to be added in an upcoming version, but the UpcomingVersion string isn't empty,
			// or vice-versa, it obviously means that gathering the information failed.
			if (p.Upcoming && p.UpcomingVersion == "") || (!p.Upcoming && p.UpcomingVersion != "") {
				t.Errorf("\tFailed on property %d, Upcoming information mismatch.\np.Upcoming = %t, p.UpcomingVersion = %s", j, p.Upcoming, p.UpcomingVersion)
			}
		}
		data, err := json.MarshalIndent(prop, "", "  ")
		if err != nil {
			t.Errorf("Test %d failed, error: %v\n", i, err)
		}
		t.Logf("Test %d data:\n%s\n", i, string(data))
	}
}
