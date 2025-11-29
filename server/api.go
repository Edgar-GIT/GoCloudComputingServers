package server

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// APIHandler manages API endpoints
type APIHandler struct {
	authManager *AuthManager
	fileManager *FileManager
	dataDir     string
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(dataDir string) *APIHandler {
	filesDir := filepath.Join(dataDir, "files")

	// Path to credentials file in admin folder
	adminDir := filepath.Join(filesDir, "admin")
	credsFile := filepath.Join(adminDir, "USER_CREDS.json")

	return &APIHandler{
		authManager: NewAuthManager(credsFile),
		fileManager: NewFileManager(filesDir),
		dataDir:     dataDir,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
}

// HandleLogin processes login requests
func (h *APIHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error processing request", http.StatusBadRequest)
		return
	}

	// Validate credentials
	if !h.authManager.Authenticate(req.Username, req.Password) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// Create user directory if it doesn't exist
	if err := h.fileManager.EnsureUserDir(req.Username); err != nil {
		http.Error(w, "Error creating user directory", http.StatusInternalServerError)
		return
	}

	// Generate token
	token, err := h.authManager.GenerateToken(req.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Success: true,
		Token:   token,
		Message: "Login successful",
	})
}

// HandleRegister processes new user registration
func (h *APIHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error processing request", http.StatusBadRequest)
		return
	}

	// Create new user
	if err := h.authManager.CreateUser(req.Username, req.Password); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Create user directory
	if err := h.fileManager.EnsureUserDir(req.Username); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error creating user directory"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleLogout processes logout requests
func (h *APIHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("Authorization")
	if token != "" {
		// Remove "Bearer " if present
		token = strings.TrimPrefix(token, "Bearer ")
		h.authManager.RevokeToken(token)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// getUsernameFromToken gets the username from the token
func (h *APIHandler) getUsernameFromToken(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	if token != "" {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	t, err := h.authManager.ValidateToken(token)
	if err != nil {
		return "", err
	}

	return t.Username, nil
}

// HandleFiles processes file-related requests
func (h *APIHandler) HandleFiles(w http.ResponseWriter, r *http.Request) {
	// Verify authentication and get username
	username, err := h.getUsernameFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authenticated"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleListFiles(w, r, username)
	case http.MethodDelete:
		h.handleDeleteFiles(w, r, username)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListFiles lists files in a folder
func (h *APIHandler) handleListFiles(w http.ResponseWriter, r *http.Request, username string) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "root"
	}

	items, err := h.fileManager.ListFiles(username, path)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"items":   items,
	})
}

// handleDeleteFiles deletes files
func (h *APIHandler) handleDeleteFiles(w http.ResponseWriter, r *http.Request, username string) {
	var req struct {
		Path  string   `json:"path"`
		Names []string `json:"names"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error processing request", http.StatusBadRequest)
		return
	}

	if len(req.Names) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "No files specified"})
		return
	}

	if err := h.fileManager.DeleteItems(username, req.Path, req.Names); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleUpload processes file uploads
func (h *APIHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify authentication and get username
	username, err := h.getUsernameFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authenticated"})
		return
	}

	// Get destination folder path
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "root"
	}

	// Parse multipart form
	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Error processing form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "No files uploaded"})
		return
	}

	// Get user directory
	userDir := h.fileManager.GetUserDir(username)

	// Normalize path
	if path == "root" || path == "/" {
		path = userDir
	} else {
		path = filepath.Join(userDir, path)
	}

	uploaded := 0
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}

		dstPath := filepath.Join(path, fileHeader.Filename)

		// Check security
		absUserDir, _ := filepath.Abs(userDir)
		absDst, _ := filepath.Abs(dstPath)
		if !strings.HasPrefix(absDst, absUserDir) {
			file.Close()
			continue
		}

		dst, err := os.Create(dstPath)
		if err != nil {
			file.Close()
			continue
		}

		_, err = io.Copy(dst, file)
		file.Close()
		dst.Close()

		if err == nil {
			uploaded++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"uploaded": uploaded,
	})
}

// HandleCreateFolder processes folder creation
func (h *APIHandler) HandleCreateFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify authentication and get username
	username, err := h.getUsernameFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authenticated"})
		return
	}

	var req struct {
		Path       string `json:"path"`
		FolderName string `json:"folderName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error processing request", http.StatusBadRequest)
		return
	}

	if req.FolderName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Folder name not specified"})
		return
	}

	if err := h.fileManager.CreateFolder(username, req.Path, req.FolderName); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleDownload processes file downloads
func (h *APIHandler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify authentication and get username
	username, err := h.getUsernameFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authenticated"})
		return
	}

	path := r.URL.Query().Get("path")
	name := r.URL.Query().Get("name")

	if name == "" {
		http.Error(w, "File name not specified", http.StatusBadRequest)
		return
	}

	// Get user directory
	userDir := h.fileManager.GetUserDir(username)

	// Normalize path
	if path == "" || path == "root" || path == "/" {
		path = userDir
	} else {
		path = filepath.Join(userDir, path)
	}

	filePath := filepath.Join(path, name)

	// Check security
	absUserDir, _ := filepath.Abs(userDir)
	absFilePath, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absFilePath, absUserDir) {
		http.Error(w, "Invalid path", http.StatusForbidden)
		return
	}

	// Check if it's a file
	info, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	if info.IsDir() {
		http.Error(w, "Cannot download a folder", http.StatusBadRequest)
		return
	}

	// Serve the file
	w.Header().Set("Content-Disposition", "attachment; filename="+name)
	http.ServeFile(w, r, filePath)
}

// HandleRename processes file renaming
func (h *APIHandler) HandleRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify authentication and get username
	username, err := h.getUsernameFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authenticated"})
		return
	}

	var req struct {
		Path    string `json:"path"`
		OldName string `json:"oldName"`
		NewName string `json:"newName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error processing request", http.StatusBadRequest)
		return
	}

	if req.OldName == "" || req.NewName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Names not specified"})
		return
	}

	if err := h.fileManager.RenameItem(username, req.Path, req.OldName, req.NewName); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
