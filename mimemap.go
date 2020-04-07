package main

import "mime"

type MimeMap map[string]string

func NewMimeMapWithDefaults() MimeMap {
	return MimeMap{
		".css":  "text/css; charset=utf-8",
		".htm":  "text/html; charset=utf-8",
		".html": "text/html; charset=utf-8",
		".js":   "text/javascript; charset=utf-8",
		".mjs":  "text/javascript; charset=utf-8",
		".txt":  "text/plain; charset=utf-8",

		".gif":  "image/gif",
		".jpeg": "image/jpeg",
		".jpg":  "image/jpeg",
		".png":  "image/png",
		".svg":  "image/svg+xml",
		".webp": "image/webp",

		".woff":  "font/woff",
		".woff2": "font/woff2",

		".json": "application/json",
		".pdf":  "application/pdf",
		".xml":  "application/xml",
		".zip":  "application/zip",
	}
}

func (mm MimeMap) FindType(ext string) string {
	if mType, found := mm[ext]; found {
		return mType
	}

	return mime.TypeByExtension(ext)
}
