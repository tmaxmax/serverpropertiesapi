package v2

import (
	"sort"
	"strings"
)

const (
	filterName     = "name"
	filterType     = "type"
	filterUpcoming = "upcoming"
)

type filters struct {
	contains  map[string]bool
	exactName string
	sortMap   map[string]bool
	typesMap  map[string]bool
	upcoming  string
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
			for substr, shouldHave := range f.contains {
				if strings.Contains(p[i].Name, substr) == !shouldHave {
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
