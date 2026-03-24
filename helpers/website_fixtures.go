package helpers

import (
	"context"
	"os"
	"path/filepath"
)

// GetSampleWebsiteFixturePath returns the absolute path to the sample website fixture directory
func GetSampleWebsiteFixturePath() string {
	basePath, err := os.Getwd()
	if err != nil {
		return "./internal/fixtures/sample-website"
	}
	return filepath.Join(basePath, "internal", "fixtures", "sample-website")
}

// GetSampleWebsiteIndexPath returns the absolute path to the sample website's index.html
func GetSampleWebsiteIndexPath() string {
	fixturePath := GetSampleWebsiteFixturePath()
	return filepath.Join(fixturePath, "index.html")
}

// ReadSampleWebsiteFixture reads the entire sample website fixture directory
// Returns a map of file paths to file contents
// This is useful for operations that need to upload a whole website to IPFS
func ReadSampleWebsiteFixture(ctx context.Context) (map[string][]byte, error) {
	fixturePath := GetSampleWebsiteFixturePath()
	files := make(map[string][]byte)

	// Read index.html
	indexPath := filepath.Join(fixturePath, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		content, err := os.ReadFile(indexPath)
		if err != nil {
			return nil, err
		}
		files["index.html"] = content
	}

	// Read CSS file
	cssPath := filepath.Join(fixturePath, "css", "styles.css")
	if _, err := os.Stat(cssPath); err == nil {
		content, err := os.ReadFile(cssPath)
		if err != nil {
			return nil, err
		}
		files["css/styles.css"] = content
	}

	// Read JavaScript file
	jsPath := filepath.Join(fixturePath, "js", "main.js")
	if _, err := os.Stat(jsPath); err == nil {
		content, err := os.ReadFile(jsPath)
		if err != nil {
			return nil, err
		}
		files["js/main.js"] = content
	}

	return files, nil
}
