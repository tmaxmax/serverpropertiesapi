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
	mathAPI                         = "http://api.mathjs.org/v4/"
	querySelector                   = `[data-description="Server properties"]:not([data-description="Server properties"] ~ [data-description="Server properties"]) tr:not(:first-child)`
	regexpLimits                    = `(\d+[-–](?:\d+|\(\d+([+\-*/^]\d+[ ]?[+\-*/^][ ]?\d+)?\)))`
	regexpLimitsNotCalculated       = `\d+[+\-*/^]\d+ ?[+\-*/^]? ?\d+?`
	minecraftBooleanTypename        = "boolean"
	minecraftIntegerTypename        = "integer"
	minecraftStringType             = "string"
	limitMalformedErrorFormatString = " limit number malformed, failed to convert to integer, error: %v"
	// PropertyDefaultLimitValue is the value that the Min and Max fields of the
	// PropertyValues struct are assigned by default. It is the minimum signed
	// 32-bit possible integer value.
	PropertyDefaultLimitValue = -int(^uint32(0)>>1) - 1
)

// The PropertyValues struct represents the legal values that a
// server.properties key may have. Based on the extracted data
// from the Wiki, it contains a potential minimum and maximum value
// it can have, and the possible values it may have.
//
//
// By default, Min and Max are PropertyDefaultLimitValue, and Possible is
// an empty slice. Min and Max must be both PropertyDefaultLimitValue
// or have any other value assigned!
// If Min and Max are PropertyDefaultLimitValue, but Possible isn't empty,
// then the key must have one of the values stored in Possible.
//
// If Min and Max are assigned different values than PropertyDefaultLimitValue,
// and the Possible slice is empty, then the key must have any value
// in the interval [Min, Max].
//
// If Min and Max are assigned different values than PropertyDefaultLimitValue,
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

// ServerProperties scrapes the official Minecraft Wiki and extracts
// the Java Server Properties, returning a slice with the keys and their
// respective documentation.
func ServerProperties() ([]Property, error) {
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
				Min:      PropertyDefaultLimitValue,
				Max:      PropertyDefaultLimitValue,
				Possible: []string{},
			},
		}
		// Loop through each row column and gather the information.
		// i is the column number, counting starts from 0. As of 6/10/2020,
		// the key name is on the first column (i = 0), the value type on the second,
		// the default value on the third and the description and possible values
		// on the fourth.
		row.ForEach(`td`, func(i int, col *colly.HTMLElement) {
			switch i {
			case 0:
				p.Name = col.ChildText(`b`)
				if strings.Contains(col.ChildText(`sup > i > span`), "upcoming") {
					p.Upcoming = true
					p.UpcomingVersion = col.ChildText(`sup > i > a`)
				}
			case 1:
				rawType := strings.TrimSpace(col.Text)
				if rawType == minecraftBooleanTypename {
					p.Type = rawType
					p.Values.Min = 0
					p.Values.Max = 1
				} else if strings.Contains(rawType, minecraftIntegerTypename) {
					p.Type = minecraftIntegerTypename

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
					// Also initialize the error variable
					values, err := rs.Split(l, 2), error(nil)
					// Evaluate the upper limit expression, if necessary
					if e := ru.FindString(values[1]); e != "" {
						result, err := evaluateMath(e)
						if err != nil {
							// This point can be reached only if the website format changes.
							// Exiting the function assures that Min and Mac are both -1 or
							// both assigned another value
							limitErr = fmt.Errorf("upper limit expression failed to evaluate, error: %v", err)
							return
						}
						p.Values.Max = result
					} else {
						p.Values.Max, err = strconv.Atoi(values[1])
						if err != nil {
							// This point shall not be reached, as the numbers can be malformed
							// only if the regexp is incorrect. The error is handled for the
							// same reason as above.
							p.Values.Max = PropertyDefaultLimitValue
							limitErr = fmt.Errorf("upper"+limitMalformedErrorFormatString, err)
							return
						}
					}
					p.Values.Min, err = strconv.Atoi(values[0])
					if err != nil {
						p.Values.Min = PropertyDefaultLimitValue
						p.Values.Max = PropertyDefaultLimitValue
						limitErr = fmt.Errorf("lower"+limitMalformedErrorFormatString, err)
						return
					}
				} else if rawType == minecraftStringType {
					p.Type = minecraftStringType
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
		// Finally, append the property instance to the return slice
		prop = append(prop, p)
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
