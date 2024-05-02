package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const DefaultPort = "1717"
const PublicDir = "public"
const ConfigPath = "stiff.json"

func main() {
	cJson, err := os.ReadFile(ConfigPath)
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
	elapsed := time.Since(startTime)
	log.Printf("Filemap generated in %d ms", elapsed.Milliseconds())

	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	server := http.Server{
		Addr:    ":" + port,
		Handler: fileServer,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Listening on port %s...", port)
		serverErr <- server.ListenAndServe()
	}()

	interrupt := make(chan os.Signal, 1)
	terminate := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(terminate, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Fatal(err)
	case <-interrupt:
		log.Print("Shutting down gracefully... (got SIGINT)")
	case <-terminate:
		log.Print("Shutting down gracefully... (got SIGTERM)")
	}

	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Failed to shutdown gracefully: %v", err)
	} else {
		log.Print("Shutdown complete")
	}
}
