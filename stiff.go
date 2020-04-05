package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const DefaultPort = "1717"
const PublicDir = "public"

func main() {
	fileServer, err := NewFileServer(PublicDir)
	if err != nil {
		log.Fatalln(err)
	}

	s := http.Server{
		Addr:    ":" + DefaultPort,
		Handler: fileServer,
	}

	log.Printf("Listening on port %s...\n", DefaultPort)
	log.Fatal(s.ListenAndServe())
}

type FileDetails struct{}

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

		p := strings.TrimPrefix(path, PublicDir)
		fileMap[p] = FileDetails{}

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
			s.Serve500(w, r)
		}
	}()

	var p string
	url := path.Clean(r.URL.Path)

	if url == "/" {
		p = filepath.FromSlash("/index.html")
	} else {
		p = filepath.FromSlash(url)
	}

	if _, found := s.fileMap[p]; !found {
		s.Serve404(w, r)
		return
	}

	targetPath := filepath.Join(PublicDir, p)
	filename := filepath.Base(targetPath)

	file, err := os.Open(targetPath)
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	fileInfo, err := file.Stat()
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	if fileInfo.IsDir() {
		s.Serve404(w, r)
		return
	}

	http.ServeContent(w, r, filename, time.Time{}, file)
}

func (s *FileServer) ServeError(w http.ResponseWriter, r *http.Request, err error) {
	if os.IsNotExist(err) {
		s.Serve404(w, r)
	} else {
		log.Println(err)
		s.Serve500(w, r)
	}
}

func (s *FileServer) Serve404(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404 page not found", http.StatusNotFound)
}

func (s *FileServer) Serve500(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
}
