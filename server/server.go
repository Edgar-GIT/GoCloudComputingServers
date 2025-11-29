package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// StartServer starts the HTTP server
func StartServer(port string, webDir string, dataDir string) error {
	// Check if web directory exists
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		return err
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	// Create files directory if it doesn't exist
	filesDir := filepath.Join(dataDir, "files")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return err
	}

	// Serve static files
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	// API routes
	apiHandler := NewAPIHandler(dataDir)
	http.HandleFunc("/api/login", apiHandler.HandleLogin)
	http.HandleFunc("/api/register", apiHandler.HandleRegister)
	http.HandleFunc("/api/logout", apiHandler.HandleLogout)
	http.HandleFunc("/api/files", apiHandler.HandleFiles)
	http.HandleFunc("/api/files/upload", apiHandler.HandleUpload)
	http.HandleFunc("/api/files/folder", apiHandler.HandleCreateFolder)
	http.HandleFunc("/api/files/download", apiHandler.HandleDownload)
	http.HandleFunc("/api/files/rename", apiHandler.HandleRename)

	log.Printf("Server started on port %s", port)
	log.Printf("Web interface available at http://localhost:%s", port)
	log.Printf("Data directory: %s", dataDir)

	return http.ListenAndServe(":"+port, nil)
}
