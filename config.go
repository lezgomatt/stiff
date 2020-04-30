package main

import "path"

type ServerConfig struct {
	Headers map[string]string `json:"headers"`
	ETag    *bool             `json:"etag"`
	LastMod *bool             `json:"lastmod"`

	Routes map[string]RouteConfig `json:"routes"`

	MimeTypes map[string]string `json:"mimetypes"`
}

type RouteConfig struct {
	Headers map[string]string `json:"headers"`
	ETag    *bool             `json:"etag"`
	LastMod *bool             `json:"lastmod"`
	Serve   string            `json:"serve"`
}

var yes = true
var no = false
var defaultConfig = RouteConfig{ETag: &yes, LastMod: &no}

func NewRouteConfig(sc *ServerConfig, rc *RouteConfig) RouteConfig {
	nc := defaultConfig
	nc.Headers = make(map[string]string)

	if sc == nil {
		return nc
	}

	for k, v := range sc.Headers {
		if v != "" {
			nc.Headers[k] = v
		}
	}

	if sc.ETag != nil {
		nc.ETag = sc.ETag
	}

	if sc.LastMod != nil {
		nc.LastMod = sc.LastMod
	}

	if rc == nil {
		return nc
	}

	for k, v := range rc.Headers {
		if v == "" {
			delete(nc.Headers, k)
		} else {
			nc.Headers[k] = v
		}
	}

	if rc.ETag != nil {
		nc.ETag = rc.ETag
	}

	if rc.LastMod != nil {
		nc.LastMod = rc.LastMod
	}

	if rc.Serve != "" {
		nc.Serve = path.Clean("/" + rc.Serve)
	}

	return nc
}
