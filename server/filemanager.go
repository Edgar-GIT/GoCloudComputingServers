package server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileItem represents a file or folder
type FileItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"` // "file" or "folder"
	Size     string `json:"size,omitempty"`
	Modified string `json:"modified"`
	Path     string `json:"path,omitempty"`
}

// FileManager manages file operations
type FileManager struct {
	baseDir string
}

// NewFileManager creates a new file manager
func NewFileManager(baseDir string) *FileManager {
	return &FileManager{
		baseDir: baseDir,
	}
}

// GetUserDir returns a user's directory
func (fm *FileManager) GetUserDir(username string) string {
	return filepath.Join(fm.baseDir, username)
}

// EnsureUserDir creates the user directory if it doesn't exist
func (fm *FileManager) EnsureUserDir(username string) error {
	userDir := fm.GetUserDir(username)
	return os.MkdirAll(userDir, 0755)
}

// ListFiles lists files in a folder (relative to user directory)
func (fm *FileManager) ListFiles(username, path string) ([]FileItem, error) {
	// Get user base directory
	userDir := fm.GetUserDir(username)

	// Normalize path
	if path == "" || path == "root" || path == "/" {
		path = userDir
	} else {
		// Check if path is relative or absolute
		if !filepath.IsAbs(path) {
			path = filepath.Join(userDir, path)
		}
		// Check if path is within userDir (security)
		absUserDir, _ := filepath.Abs(userDir)
		absPath, _ := filepath.Abs(path)
		if !strings.HasPrefix(absPath, absUserDir) {
			return nil, errors.New("invalid path")
		}
	}

	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("not a directory")
	}

	// Read directory
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var items []FileItem
	for _, entry := range entries {
		item := FileItem{
			ID:   entry.Name(),
			Name: entry.Name(),
		}

		entryPath := filepath.Join(path, entry.Name())
		info, err := os.Stat(entryPath)
		if err != nil {
			continue
		}

		if info.IsDir() {
			item.Type = "folder"
		} else {
			item.Type = "file"
			item.Size = formatSize(info.Size())
		}

		item.Modified = info.ModTime().Format("2006-01-02")
		items = append(items, item)
	}

	// Sort: folders first, then files, both by name
	sort.Slice(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type == "folder"
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	return items, nil
}

// CreateFolder creates a new folder (relative to user directory)
func (fm *FileManager) CreateFolder(username, parentPath, folderName string) error {
	// Get user base directory
	userDir := fm.GetUserDir(username)

	// Normalize path
	if parentPath == "" || parentPath == "root" || parentPath == "/" {
		parentPath = userDir
	} else {
		if !filepath.IsAbs(parentPath) {
			parentPath = filepath.Join(userDir, parentPath)
		}
		// Check security
		absUserDir, _ := filepath.Abs(userDir)
		absPath, _ := filepath.Abs(parentPath)
		if !strings.HasPrefix(absPath, absUserDir) {
			return errors.New("invalid path")
		}
	}

	// Validate folder name
	if folderName == "" || strings.ContainsAny(folderName, "/\\") {
		return errors.New("invalid folder name")
	}

	newPath := filepath.Join(parentPath, folderName)
	return os.MkdirAll(newPath, 0755)
}

// DeleteItems deletes files or folders (relative to user directory)
func (fm *FileManager) DeleteItems(username, path string, names []string) error {
	// Get user base directory
	userDir := fm.GetUserDir(username)

	// Normalize path
	if path == "" || path == "root" || path == "/" {
		path = userDir
	} else {
		if !filepath.IsAbs(path) {
			path = filepath.Join(userDir, path)
		}
		// Check security
		absUserDir, _ := filepath.Abs(userDir)
		absPath, _ := filepath.Abs(path)
		if !strings.HasPrefix(absPath, absUserDir) {
			return errors.New("invalid path")
		}
	}

	for _, name := range names {
		itemPath := filepath.Join(path, name)

		// Check security again
		absUserDir, _ := filepath.Abs(userDir)
		absItemPath, _ := filepath.Abs(itemPath)
		if !strings.HasPrefix(absItemPath, absUserDir) {
			continue
		}

		info, err := os.Stat(itemPath)
		if err != nil {
			continue
		}

		if info.IsDir() {
			os.RemoveAll(itemPath)
		} else {
			os.Remove(itemPath)
		}
	}

	return nil
}

// RenameItem renames a file or folder (relative to user directory)
func (fm *FileManager) RenameItem(username, path, oldName, newName string) error {
	// Get user base directory
	userDir := fm.GetUserDir(username)

	// Normalize path
	if path == "" || path == "root" || path == "/" {
		path = userDir
	} else {
		if !filepath.IsAbs(path) {
			path = filepath.Join(userDir, path)
		}
		// Check security
		absUserDir, _ := filepath.Abs(userDir)
		absPath, _ := filepath.Abs(path)
		if !strings.HasPrefix(absPath, absUserDir) {
			return errors.New("invalid path")
		}
	}

	// Validate new name
	if newName == "" || strings.ContainsAny(newName, "/\\") {
		return errors.New("invalid name")
	}

	oldPath := filepath.Join(path, oldName)
	newPath := filepath.Join(path, newName)

	// Check security
	absUserDir, _ := filepath.Abs(userDir)
	absOldPath, _ := filepath.Abs(oldPath)
	absNewPath, _ := filepath.Abs(newPath)
	if !strings.HasPrefix(absOldPath, absUserDir) || !strings.HasPrefix(absNewPath, absUserDir) {
		return errors.New("invalid path")
	}

	return os.Rename(oldPath, newPath)
}

// GetFileInfo gets information about a file (relative to user directory)
func (fm *FileManager) GetFileInfo(username, path, name string) (*FileItem, error) {
	// Get user base directory
	userDir := fm.GetUserDir(username)

	// Normalize path
	if path == "" || path == "root" || path == "/" {
		path = userDir
	} else {
		if !filepath.IsAbs(path) {
			path = filepath.Join(userDir, path)
		}
		// Check security
		absUserDir, _ := filepath.Abs(userDir)
		absPath, _ := filepath.Abs(path)
		if !strings.HasPrefix(absPath, absUserDir) {
			return nil, errors.New("invalid path")
		}
	}

	filePath := filepath.Join(path, name)

	// Check security
	absUserDir, _ := filepath.Abs(userDir)
	absFilePath, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absFilePath, absUserDir) {
		return nil, errors.New("invalid path")
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	item := &FileItem{
		ID:       name,
		Name:     name,
		Modified: info.ModTime().Format("2006-01-02"),
	}

	if info.IsDir() {
		item.Type = "folder"
	} else {
		item.Type = "file"
		item.Size = formatSize(info.Size())
	}

	return item, nil
}

// formatSize formats file size
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), []string{"KB", "MB", "GB", "TB"}[exp])
}
