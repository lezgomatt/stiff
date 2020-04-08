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

type FileMap map[string]FileDetails

type FileDetails struct {
	MimeType  string
	ETag      string
	HasBrotli bool
	HasGZip   bool
}

func BuildFileMap(config *ServerConfig, dir string, rm RouteMap, mm MimeMap) (FileMap, error) {
	fileMap := make(map[string]FileDetails)
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
			if fd, ok := fileMap[p]; ok {
				fd.HasBrotli = true
				fileMap[p] = fd
			}
		} else if strings.HasSuffix(path, ".gz") {
			p = strings.TrimSuffix(p, ".gz")
			if fd, ok := fileMap[p]; ok {
				fd.HasGZip = true
				fileMap[p] = fd
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

			fileMap[p] = FileDetails{ETag: eTag, MimeType: mType}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return fileMap, nil
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
