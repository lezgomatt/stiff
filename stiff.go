package main

import (
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const DefaultPort = "1717"
const PublicDir = "public"

var indexPath = filepath.FromSlash("/index.html")

func main() {
	fileServer, err := NewFileServer(PublicDir)
	if err != nil {
		log.Fatalln(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	s := http.Server{Handler: fileServer}

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on port %s...\n", port)
	log.Fatal(s.Serve(ln))
}

type FileDetails struct {
	ETag string
}

type FileServer struct {
	fileMap map[string]FileDetails
}

func NewFileServer(dir string) (*FileServer, error) {
	fileMap := make(map[string]FileDetails)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		eTag, err := computeETag(path)
		if err != nil {
			return err
		}

		p := strings.TrimPrefix(path, PublicDir)
		fileMap[p] = FileDetails{ETag: eTag}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &FileServer{fileMap: fileMap}, nil
}

func (s *FileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			s.send500(w, r)
		}
	}()

	if strings.HasSuffix(r.URL.Path, "/") && r.URL.Path != "/" {
		http.Redirect(w, r, strings.TrimRight(r.URL.Path, "/"), http.StatusMovedPermanently)
		return
	}

	if strings.HasSuffix(r.URL.Path, ".html") {
		http.Redirect(w, r, strings.TrimSuffix(r.URL.Path, ".html"), http.StatusMovedPermanently)
		return
	}

	var p string
	url := path.Clean(r.URL.Path)

	if url == "/" {
		p = indexPath
	} else {
		p = filepath.FromSlash(url)
	}

	var fileDetails FileDetails

	if fd, found := s.fileMap[p]; !found {
		fd, found := s.fileMap[p+".html"]
		if !found {
			s.send404(w, r)
			return
		}

		p = p + ".html"
		fileDetails = fd
	} else {
		fileDetails = fd
	}

	if p == indexPath && url != "/" {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	targetPath := filepath.Join(PublicDir, p)
	filename := filepath.Base(targetPath)

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

	if fileDetails.ETag != "" {
		w.Header().Set("ETag", fileDetails.ETag)
	}

	http.ServeContent(w, r, filename, time.Time{}, file)
}

func (s *FileServer) handleError(w http.ResponseWriter, r *http.Request, err error) {
	if os.IsNotExist(err) {
		s.send404(w, r)
	} else {
		log.Println(err)
		s.send500(w, r)
	}
}

func (s *FileServer) send404(w http.ResponseWriter, r *http.Request) {
	err := s.sendHTML(w, r, "/404.html", http.StatusNotFound)
	if err != nil {
		http.Error(w, "404 page not found", http.StatusNotFound)
	}
}

func (s *FileServer) send500(w http.ResponseWriter, r *http.Request) {
	err := s.sendHTML(w, r, "/500.html", http.StatusInternalServerError)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *FileServer) sendHTML(w http.ResponseWriter, r *http.Request, htmlPath string, statusCode int) error {
	p := filepath.FromSlash(htmlPath)
	if _, found := s.fileMap[p]; !found {
		return errors.New("fileserver: file not found")
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
		return errors.New("fileserver: expected an HTML file, but found a directory instead")
	}

	size := fileInfo.Size()

	w.Header().Set("Content-Type", mime.TypeByExtension(".html"))
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.WriteHeader(statusCode)

	io.CopyN(w, file, size)

	return nil
}

func computeETag(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha512.New512_256()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	checksum := base64.URLEncoding.EncodeToString(h.Sum(nil))

	return fmt.Sprintf(`W/"%s"`, checksum[:20]), nil
}
