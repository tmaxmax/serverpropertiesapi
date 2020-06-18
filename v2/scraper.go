package v2

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
	"golang.org/x/text/language"
)

const (
	mathAPI                   = "http://api.mathjs.org/v4/"
	minecraftBooleanTypename  = "boolean"
	minecraftIntegerTypename  = "integer"
	minecraftStringTypename   = "string"
	propertyDefaultLimitValue = -int(^uint32(0)>>1) - 1
	regexpLimits              = `(\d+[-–](?:\d+|\(\d+([+\-*/^]\d+[ ]?[+\-*/^][ ]?\d+)?\)))`
	regexpLimitsNotCalculated = `\d+[+\-*/^]\d+ ?[+\-*/^]? ?\d+?`
	wikiLink                  = "https://minecraft.gamepedia.com/Server.properties"
	wikiFormatLink            = "https://minecraft-%s.gamepedia.com/Server.properties"
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

func findTable(l language.Tag) (*colly.HTMLElement, error) {
	var ret *colly.HTMLElement
	link := wikiLink
	if l != language.English {
		base, _ := l.Base()
		link = fmt.Sprintf(wikiFormatLink, base.String())
	}

	c, found := colly.NewCollector(), false
	c.IgnoreRobotsTxt = false
	c.OnHTML(`.wikitable`, func(table *colly.HTMLElement) {
		rowCount := 0
		table.ForEach(`tr:first-child th`, func(_ int, _ *colly.HTMLElement) {
			rowCount++
		})
		if rowCount != 4 || found {
			return
		}
		found = true
		ret = table
	})
	err := c.Visit(link)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func getLanguages() ([]language.Tag, error) {
	tags := []language.Tag{language.English}

	c := colly.NewCollector()
	c.IgnoreRobotsTxt = false
	c.OnHTML(`#p-lang .interlanguage-link a`, func(l *colly.HTMLElement) {
		tags = append(tags, language.MustParse(l.Attr("lang")))
	})
	err := c.Visit(wikiLink)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// ServerProperties is the main function that scrapes the official documentation based on the passed Filters.
func ServerProperties(f *Filters) ([]Property, error) {
	langs, err := getLanguages()
	if err != nil {
		return nil, err
	}

	matcher := language.NewMatcher(langs)
	l, _, _ := matcher.Match(f.language)

	table, err := findTable(language.English)
	if err != nil {
		return nil, err
	}

	var (
		ret      = make([]Property, 0, 50)
		indexMap = make(map[string]int)
		index    int
		limitErr = error(nil)
	)

	table.ForEach(`tr:not(:first-child)`, func(_ int, row *colly.HTMLElement) {
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
							limitErr = fmt.Errorf("upper limit expression failed to evaluate, error: %v", err)
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
		indexMap[p.Name] = index
		index++
	})
	if limitErr != nil {
		return nil, limitErr
	}

	if f != nil && f.language != language.English {
		langTable, err := findTable(l)

		if err != nil {
			return nil, err
		}

		langTable.ForEach(`tr:not(:first-child)`, func(_ int, row *colly.HTMLElement) {
			name := row.ChildText(`th > span > b`)
			descOn3 := false
			row.ForEach(`td`, func(i int, col *colly.HTMLElement) {
				switch i {
				case 0:
					if name == "" {
						name = col.ChildText(`b`)
						descOn3 = true
					}
				case 2:
					if idx, ok := indexMap[name]; !descOn3 && ok {
						ret[idx].Description = strings.TrimSpace(col.Text)
					}
				case 3:
					if idx, ok := indexMap[name]; descOn3 && ok {
						ret[idx].Description = strings.TrimSpace(col.Text)
					}
				}
			})
		})
	}

	if f != nil {
		ret = f.Filter(ret)
		f.Sort(ret)
	}

	return ret, nil
}
