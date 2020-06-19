package v2

import (
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
	minecraftBooleanTypename  = "boolean"
	minecraftIntegerTypename  = "integer"
	minecraftStringTypename   = "string"
	propertyDefaultLimitValue = -int(^uint32(0)>>1) - 1
	querySelector             = `[data-description="Server properties"]:not([data-description="Server properties"] ~ [data-description="Server properties"]) tr:not(:first-child)`
	regexpLimits              = `(\d+[-–](?:\d+|\(\d+([+\-*/^]\d+[ ]?[+\-*/^][ ]?\d+)?\)))`
	regexpLimitsNotCalculated = `\d+[+\-*/^]\d+ ?[+\-*/^]? ?\d+?`
	wikiLink                  = "https://minecraft.gamepedia.com/Server.properties"
)

// The PropertyValues struct holds all the documented values of a property.
type PropertyValues struct {
	Default  string   `json:"default"`
	Min      int      `json:"min"`
	Max      int      `json:"max"`
	Possible []string `json:"possible"`
}

// The Property struct represents a property with all the information from the official documentation.
type Property struct {
	Name            string         `json:"name"`
	Type            string         `json:"type"`
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

// ServerProperties is the main function that scrapes the official documentation based on the passed filters.
func ServerProperties(f *filters) ([]Property, error) {
	var (
		ret      = make([]Property, 0, 50)
		limitErr = error(nil)
	)

	c := colly.NewCollector()
	c.IgnoreRobotsTxt = false

	c.OnHTML(querySelector, func(row *colly.HTMLElement) {
		p := Property{
			Values: PropertyValues{
				Min:      propertyDefaultLimitValue,
				Max:      propertyDefaultLimitValue,
				Possible: []string{},
			},
		}
		row.ForEach(`td`, func(i int, col *colly.HTMLElement) {
			switch i {
			case 0:
				p.Name = col.ChildText(`b`)
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
					// The regexps are always correct. They must be verified to work on all values before modifying them.
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
							// Exiting the function assures that Min and Max are both propertyDefaultLimitValue or
							// both assigned another value
							limitErr = err
							return
						}
						p.Values.Max = result
					} else {
						p.Values.Max, _ = strconv.Atoi(values[1])
					}
					p.Values.Min, _ = strconv.Atoi(values[0])
				} else if rawType == minecraftStringTypename {
					p.Type = minecraftStringTypename
				}
			case 2:
				p.Values.Default = strings.TrimSpace(col.Text)
			case 3:
				p.Description = strings.TrimSpace(col.Text)
				// Loop through each significant value, appending them to the Possible
				// field of the Values field of the p instance. If property is boolean,
				// this is unnecessary.
				if p.Type == minecraftBooleanTypename {
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
		ret = append(ret, p)
	})
	err := c.Visit(wikiLink)
	if err != nil {
		return nil, err
	}
	if limitErr != nil {
		return nil, limitErr
	}

	if f != nil {
		ret = f.filter(ret)
		f.sort(ret)
	}

	return ret, nil
}
