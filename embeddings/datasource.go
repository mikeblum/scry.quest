package embeddings

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// FileSystemDataSource reads content from filesystem.
type FileSystemDataSource struct {
	basePath    string
	contentType string
	extensions  []string
}

// NewFileSystemDataSource creates a filesystem data source.
func NewFileSystemDataSource(basePath, contentType string, extensions []string) *FileSystemDataSource {
	return &FileSystemDataSource{
		basePath:    basePath,
		contentType: contentType,
		extensions:  extensions,
	}
}

// Read scans filesystem and returns content items.
func (fs *FileSystemDataSource) Read(ctx context.Context) (<-chan *ContentItem, error) {
	items := make(chan *ContentItem)

	go func() {
		defer close(items)

		err := filepath.Walk(fs.basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			// Check if file has a supported extension
			if !fs.hasValidExtension(path) {
				return nil
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			item, err := fs.createContentItem(path)
			if err != nil {
				return fmt.Errorf("failed to create content item from %s: %w", path, err)
			}

			items <- item
			return nil
		})

		if err != nil {
			// Log error but don't block the channel
			return
		}
	}()

	return items, nil
}

// Close implements DataSource.
func (fs *FileSystemDataSource) Close() error {
	return nil
}

// hasValidExtension checks file extension validity.
func (fs *FileSystemDataSource) hasValidExtension(path string) bool {
	if len(fs.extensions) == 0 {
		return true // Accept all files if no extensions specified
	}

	ext := strings.ToLower(filepath.Ext(path))
	for _, validExt := range fs.extensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

// createContentItem creates a ContentItem from a file path
func (fs *FileSystemDataSource) createContentItem(path string) (*ContentItem, error) {
	content, err := os.ReadFile(filepath.Clean(path)) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	relPath, err := filepath.Rel(fs.basePath, path)
	if err != nil {
		relPath = filepath.Base(path)
	}

	item := &ContentItem{
		ID:      uuid.New(),
		Content: content,
		Type:    fs.contentType,
		Metadata: map[string]interface{}{
			"file_path": path,
			"rel_path":  relPath,
			"file_size": len(content),
			"extension": filepath.Ext(path),
			"base_name": strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		},
	}

	// If it's a JSON file, try to extract metadata
	if strings.ToLower(filepath.Ext(path)) == ".json" {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(content, &jsonData); err == nil {
			item.Metadata["json_data"] = jsonData

			// Extract common fields if they exist
			if name, ok := jsonData["name"].(string); ok {
				item.Metadata["name"] = name
			}
		}
	}

	return item, nil
}

// SliceDataSource creates content items from slice.
type SliceDataSource struct {
	items  []*ContentItem
	index  int
	closed bool
}

// NewSliceDataSource creates a slice data source.
func NewSliceDataSource(items []*ContentItem) *SliceDataSource {
	return &SliceDataSource{
		items: items,
		index: 0,
	}
}

// Read returns channel emitting all items.
func (s *SliceDataSource) Read(ctx context.Context) (<-chan *ContentItem, error) {
	if s.closed {
		return nil, fmt.Errorf("data source is closed")
	}

	items := make(chan *ContentItem)

	go func() {
		defer close(items)

		for i := s.index; i < len(s.items); i++ {
			select {
			case <-ctx.Done():
				return
			case items <- s.items[i]:
			}
		}
	}()

	return items, nil
}

// Close implements DataSource.
func (s *SliceDataSource) Close() error {
	s.closed = true
	return nil
}
