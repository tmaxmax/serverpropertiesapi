package serverpropertiesapi

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

const (
	mathAPI                   = "http://api.mathjs.org/v4/"
	querySelector             = `[data-description="Server properties"]:not([data-description="Server properties"] ~ [data-description="Server properties"]) tr:not(:first-child)`
	regexpLimits              = `(\d+[-–](?:\d+|\(\d+([+\-*/^]\d+[ ]?[+\-*/^][ ]?\d+)?\)))`
	regexpLimitsNotCalculated = `\d+[+\-*/^]\d+ ?[+\-*/^]? ?\d+?`
	minecraftBooleanTypename  = "boolean"
	minecraftIntegerTypename  = "integer"
	minecraftStringType       = "string"
	// propertyDefaultLimitValue is the value that the Min and Max fields of the
	// PropertyValues struct are assigned by default. It is the minimum signed
	// 32-bit possible integer value.
	propertyDefaultLimitValue = -int(^uint32(0)>>1) - 1
)

// The PropertyValues struct represents the legal values that a
// server.properties key may have. Based on the extracted data
// from the Wiki, it contains a potential minimum and maximum value
// it can have, and the possible values it may have.
//
//
// By default, Min and Max are propertyDefaultLimitValue, and Possible is
// an empty slice. Min and Max must be both propertyDefaultLimitValue
// or have any other value assigned!
// If Min and Max are propertyDefaultLimitValue, but Possible isn't empty,
// then the key must have one of the values stored in Possible.
//
// If Min and Max are assigned different values than propertyDefaultLimitValue,
// and the Possible slice is empty, then the key must have any value
// in the interval [Min, Max].
//
// If Min and Max are assigned different values than propertyDefaultLimitValue,
// and the Possible slice isn't empty, then the key must have either
// the values stored in the Possible slice, or values in the interval [Min, Max].
// The values stored in Possible may have a special significance in this case,
// which is documented in the key's description.
type PropertyValues struct {
	Min      int      `json:"min"`
	Max      int      `json:"max"`
	Possible []string `json:"possible"`
}

// The Property struct represents a server.properties key, along with
// its documentation.
type Property struct {
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	Default         string         `json:"defaultValue"`
	Values          PropertyValues `json:"values"`
	Description     string         `json:"description"`
	Upcoming        bool           `json:"upcoming"`
	UpcomingVersion string         `json:"upcomingVersion"`
}

// The Options struct contains parameters used when scraping the wiki,
// to only get relevant properties.
type Options struct {
	Contains  string `json:"contains"`
	exactName string
	Type      string `json:"type"`
	Upcoming  string `json:"upcoming"`
}

// evaluateMath computes the result of a string expression, using the
// mathjs API: https://api.mathjs.org/
//
// On any error, the return value is always 0.
func evaluateMath(e string) (int, error) {
	res, err := http.Get(mathAPI + "?expr=" + url.QueryEscape(e))
	if err != nil {
		return 0, err
	}
	body, _ := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	// Float precision is set to 64, as the input numbers are 64-bit integers.
	calc, _, err := big.ParseFloat(string(body), 10, 64, big.ToNearestEven)
	if err != nil {
		return 0, nil
	}
	calcConv, _ := calc.Int64()
	return int(calcConv), nil
}

// ServerProperty scrapes the official Minecraft Wiki and extracts
// the Java Server Property with the key name identical to the exactName
// string parameter. Returns an empty instance and a "not found" error.
func ServerProperty(exactName string) (Property, error) {
	found, err := ServerProperties(Options{
		exactName: exactName,
	})
	if err != nil {
		return Property{}, err
	}
	if len(found) == 0 {
		return Property{}, fmt.Errorf("not found")
	}
	return found[0], nil
}

// ServerProperties scrapes the official Minecraft Wiki and extracts
// the Java Server Properties, returning a slice with the keys and their
// respective documentation.
func ServerProperties(o Options) ([]Property, error) {
	prop := make([]Property, 0, 50)
	limitErr := error(nil)

	c := colly.NewCollector()
	c.IgnoreRobotsTxt = false

	// Loop through each table row and append a new Property instance, based
	// on the information on the row.
	c.OnHTML(querySelector, func(row *colly.HTMLElement) {
		// Create a default Property instance, initialized with the values
		// mentioned in the documentation.
		p := Property{
			Values: PropertyValues{
				Min:      propertyDefaultLimitValue,
				Max:      propertyDefaultLimitValue,
				Possible: []string{},
			},
		}
		valid := true
		// Loop through each row column and gather the information.
		// i is the column number, counting starts from 0. As of 6/10/2020,
		// the key name is on the first column (i = 0), the value type on the second,
		// the default value on the third and the description and possible values
		// on the fourth.
		row.ForEach(`td`, func(i int, col *colly.HTMLElement) {
			// Don't execute if the property is already invalid
			if !valid {
				return
			}

			switch i {
			case 0:
				p.Name = col.ChildText(`b`)
				// If the name doesn't contain the specified string, or if it isn't equal to the name requested,
				// mark as invalid
				if o.Contains != "" && !strings.Contains(p.Name, o.Contains) || o.exactName != "" && p.Name != o.exactName {
					valid = false
					return
				}
				if strings.Contains(col.ChildText(`sup > i > span`), "upcoming") {
					// If upcoming features aren't requested, mark as invalid
					if o.Upcoming == "false" {
						valid = false
						return
					}
					p.Upcoming = true
					p.UpcomingVersion = col.ChildText(`sup > i > a`)
				} else if o.Upcoming == "true" {
					valid = false
					return
				}
			case 1:
				rawType := strings.TrimSpace(col.Text)
				if rawType == minecraftBooleanTypename {
					p.Type = rawType
					p.Values.Min = 0
					p.Values.Max = 1
				} else if strings.Contains(rawType, minecraftIntegerTypename) {
					p.Type = minecraftIntegerTypename

					if o.Type != "" && p.Type != o.Type {
						// If the type requested isn't minecraftIntegerTypename, mark as invalid
						valid = false
						return
					}

					// The three regexps are used to get the limits string, split into two separate values and extract
					// the upper limit into an evaluable expression, when it isn't a number.
					// The regexps are always correct, as they are hardcoded. They must be verified to work on all values
					// before modifying them.
					// NOTE: The second, longer dash is included because of an inconsistency on the website.
					r, rs, ru := regexp.MustCompile(regexpLimits), regexp.MustCompile(`([-–])`), regexp.MustCompile(regexpLimitsNotCalculated)

					// l is the limits string, which is found in the rawType.
					l := r.FindString(rawType)
					if l == "" {
						// nothing to do here when the limits string is empty
						return
					}

					// Get the two values.
					values := rs.Split(l, 2)
					// Evaluate the upper limit expression, if necessary
					if e := ru.FindString(values[1]); e != "" {
						result, err := evaluateMath(e)
						if err != nil {
							// This point can be reached only if the mathjs API is down.
							// Exiting the function assures that Min and Max are both -1 or
							// both assigned another value
							limitErr = fmt.Errorf("upper limit expression failed to evaluate, error: %v", err)
							return
						}
						p.Values.Max = result
					} else {
						p.Values.Max, _ = strconv.Atoi(values[1])
					}
					p.Values.Min, _ = strconv.Atoi(values[0])
				} else if rawType == minecraftStringType {
					p.Type = minecraftStringType
				}
				// If the property isn't of specified type,
				if o.Type != "" && p.Type != o.Type {
					valid = false
					return
				}
			case 2:
				p.Default = strings.TrimSpace(col.Text)
			case 3:
				p.Description = strings.TrimSpace(col.Text)
				// Loop through each significant value, appending them to the Possible
				// field of the Values field of the p instance. If property is boolean,
				// this is unnecessary.
				if p.Type == "boolean" {
					return
				}
				col.ForEach(`dl dd`, func(_ int, val *colly.HTMLElement) {
					t := val.ChildText(`b:first-child`)
					// Don't append empty values, as they are coded in the markdown of the
					// website, and also don't append Note notices, as they are inside the same
					// tag structure in the HTML
					if t == "" || t == "Note:" {
						return
					}
					p.Values.Possible = append(p.Values.Possible, t)
				})
			}
		})
		// Finally, append the property instance to the return slice, if it is valid
		if valid {
			prop = append(prop, p)
		}
	})

	// Visit the website, to gather the properties
	err := c.Visit("https://minecraft.gamepedia.com/Server.properties")
	if err != nil {
		// If there is an error on visiting the website, then no data can be obtained.
		return nil, err
	}
	if limitErr != nil {
		// If it failed to parse the limits, return the respective error.
		return nil, limitErr
	}

	// If this point is reached, the the information has been gathered successfully.
	return prop, nil
}
