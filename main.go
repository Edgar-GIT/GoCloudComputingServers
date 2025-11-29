package main

import (
	"flag"
	"log"
	"path/filepath"

	"GoCloudComputingServers/server"
)

func main() {
	// Define flags for configuration
	port := flag.String("port", "8080", "Server port")
	webDir := flag.String("web", "./web", "Web files directory")
	dataDir := flag.String("data", "./data", "Data directory")
	flag.Parse()

	// Convert to absolute paths
	webPath, err := filepath.Abs(*webDir)
	if err != nil {
		log.Fatal("Error getting absolute path of web directory:", err)
	}

	dataPath, err := filepath.Abs(*dataDir)
	if err != nil {
		log.Fatal("Error getting absolute path of data directory:", err)
	}

	log.Println("=== File Manager Server ===")
	log.Printf("Port: %s", *port)
	log.Printf("Web Directory: %s", webPath)
	log.Printf("Data Directory: %s", dataPath)
	log.Println("==========================")

	// Start server
	if err := server.StartServer(*port, webPath, dataPath); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
