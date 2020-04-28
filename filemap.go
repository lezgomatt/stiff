package main

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const HashLength = 16

type FileMap struct {
	files      map[string]FileDetails
	errorPages map[string]FileDetails
}

type FileDetails struct {
	MimeType  string
	ETag      string
	HasBrotli bool
	HasGzip   bool
}

func BuildFileMap(config *ServerConfig, dir string, rm RouteMap, mm MimeMap) (FileMap, error) {
	files := make(map[string]FileDetails)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		p := strings.TrimPrefix(path, PublicDir)
		if strings.HasSuffix(p, ".br") {
			p = strings.TrimSuffix(p, ".br")
			if fd, found := files[p]; found {
				fd.HasBrotli = true
				files[p] = fd
			}
		} else if strings.HasSuffix(path, ".gz") {
			p = strings.TrimSuffix(p, ".gz")
			if fd, found := files[p]; found {
				fd.HasGzip = true
				files[p] = fd
			}
		} else {
			mType := mm.FindType(filepath.Ext(p))

			var eTag string
			if *rm.GetConfig(strings.TrimSuffix(p, ".html")).ETag {
				eTag, err = computeETag(path, mType)
				if err != nil {
					return err
				}
			}

			files[p] = FileDetails{ETag: eTag, MimeType: mType}
		}

		return nil
	})

	if err != nil {
		return FileMap{}, err
	}

	errorPages := make(map[string]FileDetails)
	errorPaths := []string{"/404.html", "/500.html"}

	for _, p := range errorPaths {
		if fd, found := files[p]; found {
			errorPages[p] = fd
			delete(files, p)
		}
	}

	return FileMap{files: files, errorPages: errorPages}, nil
}

func computeETag(path string, mType string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha512.New512_256()
	io.WriteString(h, mType)
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	checksum := base64.URLEncoding.EncodeToString(h.Sum(nil))

	// use weak etags to allow the same hash for compressed versions
	return fmt.Sprintf(`W/"%s"`, checksum[:HashLength]), nil
}
