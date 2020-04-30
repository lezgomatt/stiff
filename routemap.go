package main

import (
	"fmt"
	"sort"
	"strings"
)

type RouteMap []RouteRule

func NewRouteMap(sc *ServerConfig) (RouteMap, error) {
	var rm RouteMap

	if sc == nil {
		return rm, nil
	}

	routes := make([]string, 0, len(sc.Routes))
	for r := range sc.Routes {
		if !strings.HasPrefix(r, "/") {
			return nil, fmt.Errorf("stiff.json: invalid route %q, missing leading slash", r)
		}

		routes = append(routes, r)
	}

	// sort in reverse so that the longer matches will go first
	sort.Sort(sort.Reverse(sort.StringSlice(routes)))

	for _, route := range routes {
		rc := sc.Routes[route]
		rule := RouteRule{
			pattern:  route,
			matchDir: strings.HasSuffix(route, "/"),
			config:   NewRouteConfig(sc, &rc),
		}

		rm = append(rm, rule)
	}

	if _, found := sc.Routes["/"]; !found {
		rm = append(rm, RouteRule{pattern: "/", matchDir: true, config: NewRouteConfig(sc, nil)})
	}

	return rm, nil
}

func (rm RouteMap) GetConfig(route string) RouteConfig {
	for _, r := range rm {
		if r.matches(route) {
			return r.config
		}
	}

	return defaultConfig
}

type RouteRule struct {
	pattern  string
	matchDir bool
	config   RouteConfig
}

func (r RouteRule) matches(route string) bool {
	if r.matchDir {
		return strings.HasPrefix(route, r.pattern)
	} else {
		return route == r.pattern
	}
}
