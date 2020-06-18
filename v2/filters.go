package v2

import (
	"sort"
	"strings"

	"golang.org/x/text/language"
)

const (
	filterName     = "name"
	filterType     = "type"
	filterUpcoming = "upcoming"
)

type Filters struct {
	contains  map[string]bool
	exactName string
	language  language.Tag
	sort      map[string]bool
	types     map[string]bool
	upcoming  string
}

func (f *Filters) Valid(types map[string]struct{}) bool {
	for key, _ := range f.types {
		if _, ok := types[key]; !ok {
			return false
		}
	}
	for key, _ := range f.sort {
		if key != filterName && key != filterType && key != filterUpcoming {
			return false
		}
	}
	if f.upcoming != "" && f.upcoming != "true" && f.upcoming != "false" {
		return false
	}
	return true
}

func (f *Filters) Filter(p []Property) []Property {
	ret := make([]Property, 0)
	if f.exactName != "" {
		for i := range p {
			if p[i].Name == f.exactName {
				ret = append(ret, p[i])
				break
			}
		}
		return ret
	}
	if f.types == nil && f.contains == nil && f.upcoming == "" {
		return p
	}
	for i := range p {
		if f.types != nil {
			if accept, ok := f.types[p[i].Type]; !accept || !ok {
				continue
			}
		}
		if f.contains != nil {
			doesntMatchFilter := false
			for substr, negate := range f.contains {
				if hasSubstr := strings.Contains(p[i].Name, substr); hasSubstr == negate {
					doesntMatchFilter = true
					break
				}
			}
			if doesntMatchFilter {
				continue
			}
		}
		if f.upcoming != "" {
			if (p[i].Upcoming && f.upcoming == "false") || (!p[i].Upcoming && f.upcoming == "true") {
				continue
			}
		}
		ret = append(ret, p[i])
	}
	return ret
}

func (f *Filters) Sort(p []Property) {
	if order, ok := f.sort[filterName]; ok {
		if order {
			sort.Slice(p, func(i, j int) bool {
				return p[i].Name < p[j].Name
			})
		} else {
			sort.Slice(p, func(i, j int) bool {
				return p[i].Name > p[j].Name
			})
		}
	}
	if order, ok := f.sort[filterType]; ok {
		if order {
			sort.Slice(p, func(i, j int) bool {
				return p[i].Type < p[j].Type
			})
		} else {
			sort.Slice(p, func(i, j int) bool {
				return p[i].Type > p[j].Type
			})
		}
	}
	if order, ok := f.sort[filterUpcoming]; ok {
		if order {
			sort.Slice(p, func(i, j int) bool {
				return p[i].Upcoming == false && p[j].Upcoming == true
			})
		} else {
			sort.Slice(p, func(i, j int) bool {
				return p[i].Upcoming == true && p[j].Upcoming == false
			})
		}
	}
}
