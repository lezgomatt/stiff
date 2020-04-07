package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

const DefaultPort = "1717"
const PublicDir = "public"
const ConfigPath = "stiff.json"

func main() {
	cJson, err := ioutil.ReadFile(ConfigPath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	var config *ServerConfig
	if cJson != nil {
		config = new(ServerConfig)
		err := json.Unmarshal(cJson, config)
		if err != nil {
			log.Fatalf("stiff.json: %s", err.Error())
		}
	}

	startTime := time.Now()
	fileServer, err := NewFileServer(config, PublicDir)
	if err != nil {
		log.Fatal(err)
	}
	elapsed := time.Now().Sub(startTime)
	log.Printf("filemap generated in %d ms", elapsed.Milliseconds())

	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	s := http.Server{Handler: fileServer}

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("listening on port %s...", port)
	log.Fatal(s.Serve(ln))
}
