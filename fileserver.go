package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var indexPath = filepath.FromSlash("/index.html")

type FileServer struct {
	routeMap RouteMap
	fileMap  FileMap
}

func NewFileServer(config *ServerConfig, dir string) (*FileServer, error) {
	rm, err := NewRouteMap(config)
	if err != nil {
		return nil, err
	}

	mm := NewMimeMapWithDefaults()
	if config != nil {
		for ext, mType := range config.MimeTypes {
			if !strings.HasPrefix(ext, ".") {
				return nil, fmt.Errorf(`stiff.json: invalid extension %q, missing dot`, ext)
			}

			mm[ext] = mType
		}
	}

	fm, err := BuildFileMap(config, dir, rm, mm)
	if err != nil {
		return nil, err
	}

	return &FileServer{routeMap: rm, fileMap: fm}, nil
}

func (s *FileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Print(err)
			s.send500(w, r)
		}
	}()

	url := path.Clean("/" + r.URL.Path)
	rc := s.routeMap.GetConfig(url)

	for k, v := range rc.Headers {
		w.Header().Set(k, v)
	}

	if url != r.URL.Path || strings.HasSuffix(url, ".html") {
		http.Redirect(w, r, strings.TrimSuffix(url, ".html"), http.StatusMovedPermanently)
		return
	}

	var p string
	if rc.Serve != "" {
		p = rc.Serve
	} else if url == "/" {
		p = indexPath
	} else {
		p = filepath.FromSlash(url)
	}

	var fileDetails FileDetails

	if fd, found := s.fileMap.files[p]; found {
		fileDetails = fd
	} else if fd, found := s.fileMap.files[p+".html"]; found {
		p += ".html"
		fileDetails = fd
	} else {
		s.send404(w, r)
		return
	}

	if p == indexPath && url != "/" {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	targetPath := filepath.Join(PublicDir, p)
	rangeReq := r.Header.Get("Range")
	if fileDetails.HasBrotli || fileDetails.HasGZip {
		w.Header().Add("Vary", "Accept-Encoding")

		ae := parseAcceptEncoding(r.Header.Get("Accept-Encoding"))
		if ae.Brotli && rangeReq == "" && fileDetails.HasBrotli {
			w.Header().Set("Content-Encoding", "br")
			targetPath += ".br"
		} else if ae.GZip && rangeReq == "" && fileDetails.HasGZip {
			w.Header().Set("Content-Encoding", "gzip")
			targetPath += ".gz"
		}
	}

	file, err := os.Open(targetPath)
	if err != nil {
		s.handleError(w, r, err)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	if fileInfo.IsDir() {
		s.send404(w, r)
		return
	}

	w.Header().Set("Content-Type", fileDetails.MimeType)

	if rangeReq == "" {
		// http.ServeContent skips Content-Length when Content-Encoding is set
		w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	}

	if fileDetails.ETag != "" {
		w.Header().Set("ETag", fileDetails.ETag)
	}

	var modTime time.Time
	if *rc.LastMod {
		modTime = fileInfo.ModTime()
	} else {
		modTime = time.Time{}
	}

	http.ServeContent(w, r, p, modTime, file)
}

func (s *FileServer) handleError(w http.ResponseWriter, r *http.Request, err error) {
	if os.IsNotExist(err) {
		s.send404(w, r)
	} else {
		log.Print(err)
		s.send500(w, r)
	}
}

func (s *FileServer) send404(w http.ResponseWriter, r *http.Request) {
	w.Header().Del("Content-Encoding")
	w.Header().Del("Cache-Control")
	w.Header().Del("ETag")
	w.Header().Del("Last-Modified")

	err := s.sendErrorPage(w, r, "/404.html", http.StatusNotFound)
	if err != nil {
		http.Error(w, "404 page not found", http.StatusNotFound)
	}
}

func (s *FileServer) send500(w http.ResponseWriter, r *http.Request) {
	w.Header().Del("Content-Encoding")
	w.Header().Del("Cache-Control")
	w.Header().Del("ETag")
	w.Header().Del("Last-Modified")

	err := s.sendErrorPage(w, r, "/500.html", http.StatusInternalServerError)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *FileServer) sendErrorPage(w http.ResponseWriter, r *http.Request, errorPath string, statusCode int) error {
	p := filepath.FromSlash(errorPath)
	fd, found := s.fileMap.errorPages[p]
	if !found {
		return fmt.Errorf("fileserver: file not found %q", errorPath)
	}

	file, err := os.Open(filepath.Join(PublicDir, p))
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("fileserver: expected HTML file %q, found a directory instead", errorPath)
	}

	size := fileInfo.Size()

	w.Header().Set("Content-Type", fd.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.WriteHeader(statusCode)

	io.CopyN(w, file, size)

	return nil
}
