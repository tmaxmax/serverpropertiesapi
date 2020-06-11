package serverpropertiesapi

import (
	"encoding/json"
	"testing"
)

func TestServerProperties(t *testing.T) {
	prop, err := ServerProperties()
	if err != nil {
		t.Errorf("Failed. Error: %v\n", err)
	}

	for i, p := range prop {
		// If any of these fields are empty, it means that gathering the information from the website failed.
		// This could be caused by a redesign, for example.
		if p.Name == "" || p.Type == "" || p.Default == "" || p.Description == "" {
			t.Errorf("\tFailed on property %d, one or more fields are empty. Check if website format hasn't changed!\nProperty instance: %+v\n", i, p)
		}
		// This is to assure that the limits have the default value if they are not mentioned in the documentation.
		if p.Values.Min == PropertyDefaultLimitValue && p.Values.Max != p.Values.Min {
			t.Errorf("\tFailed on property %d, invalid Min and Max.\np.Values.Min = %d, p.Values.Max = %d\n", i, p.Values.Min, p.Values.Max)
		}
		// If the feature isn't going to be added in an upcoming version, but the UpcomingVersion string isn't empty,
		// or vice-versa, it obviously means that gathering the information failed.
		if (p.Upcoming && p.UpcomingVersion == "") || (!p.Upcoming && p.UpcomingVersion != "") {
			t.Errorf("\tFailed on property %d, Upcoming information mismatch.\np.Upcoming = %t, p.UpcomingVersion = %s", i, p.Upcoming, p.UpcomingVersion)
		}
	}

	jsonData, err := json.MarshalIndent(prop, "", "  ")
	if err != nil {
		t.Errorf("Failed, couldn't marshal JSON. Error: %v\n", err)
	}
	t.Logf("Succeeded. Marshaled JSON:\n%s\n", string(jsonData))
}
