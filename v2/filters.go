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

type filters struct {
	contains  map[string]bool
	exactName string
	language  language.Tag
	sortMap   map[string]bool
	typesMap  map[string]bool
	upcoming  string
}

func (f *filters) valid(types map[string]struct{}) bool {
	for key := range f.typesMap {
		if _, ok := types[key]; !ok {
			return false
		}
	}
	for key := range f.sortMap {
		if key != filterName && key != filterType && key != filterUpcoming {
			return false
		}
	}
	if f.upcoming != "" && f.upcoming != "true" && f.upcoming != "false" {
		return false
	}
	return true
}

func (f *filters) filter(p []Property) []Property {
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
	if f.typesMap == nil && f.contains == nil && f.upcoming == "" {
		return p
	}
	for i := range p {
		if f.typesMap != nil {
			if accept, ok := f.typesMap[p[i].Type]; !accept || !ok {
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

func (f *filters) sort(p []Property) {
	if order, ok := f.sortMap[filterName]; ok {
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
	if order, ok := f.sortMap[filterType]; ok {
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
	if order, ok := f.sortMap[filterUpcoming]; ok {
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
