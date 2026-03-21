package helpers

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"io/fs"
	"os"
)

// IPFSPortalUpload uploads content to the portal via the IPFS SDK upload endpoint.
// This creates an account operation and returns the CID.
// The caller must wait for operation completion using WaitForOperation.
// This is for testing user uploads to IPFS (POST to portal), NOT for pinning existing CIDs.
// The content is wrapped in CAR format automatically by the SDK.
func IPFSPortalUpload(ctx context.Context, content []byte, filename string) (string, error) {
	// Get IPFS SDK client
	client, err := GetIPFSClient(ctx)
	if err != nil {
		return "", err
	}

	// Upload using the SDK's UploadBytes method which handles CAR wrapping automatically
	uploadResult, err := client.Upload().UploadBytes(ctx, content, filename, nil)
	if err != nil {
		return "", err
	}

	return uploadResult.CID, nil
}

// IPFSPortalUploadFromFS uploads a file from disk to the portal via the IPFS SDK upload endpoint.
// This creates an account operation and returns the CID.
// The caller must wait for operation completion using WaitForOperation.
// This is for testing large file uploads without holding data in memory.
// ThefilePath is the path to the file on disk to upload.
func IPFSPortalUploadFromFS(ctx context.Context, filePath string, filename string) (string, error) {
	// Get IPFS SDK client
	client, err := GetIPFSClient(ctx)
	if err != nil {
		return "", err
	}

	// Open the file for upload
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Upload using the SDK's UploadFile method which wraps single files in filesystem
	// UploadFile automatically wraps the file in a SingleFileFS and uploads via UploadFromFS
	uploadResult, err := client.Upload().UploadFile(ctx, file, filename, nil)
	if err != nil {
		return "", err
	}

	return uploadResult.CID, nil
}

// IPFSPortalUploadDirFromFS uploads a directory from disk to the portal via the IPFS SDK upload endpoint.
// This creates an account operation and returns the directory CID.
// The caller must wait for operation completion using WaitForOperation.
// This is for testing directory uploads without holding data in memory.
// dirPath is the path to the directory to upload.
func IPFSPortalUploadDirFromFS(ctx context.Context, dirPath string) (string, error) {
	// Get IPFS SDK client
	client, err := GetIPFSClient(ctx)
	if err != nil {
		return "", err
	}

	// Create a filesystem from the directory
	var fsys fs.FS = os.DirFS(dirPath)

	// Upload using the SDK's UploadFromFS method which handles directory CAR generation automatically
	// UploadFromFS will detect this is a directory and wrap it propery
	uploadResult, err := client.Upload().UploadFromFS(ctx, fsys, dirPath, nil)
	if err != nil {
		return "", err
	}

	return uploadResult.CID, nil
}

// DownloadAndVerifyContent downloads content from IPFS and verifies it matches original
// This consolidates a common pattern used across multiple step definitions.
// The contextDesc parameter provides context for error messages.
func DownloadAndVerifyContent(ctx context.Context, cidStr, originalContent, contextDesc string) error {
	parsedCID, err := ParseCID(cidStr)
	if err != nil {
		return fmt.Errorf("failed to parse CID: %w", err)
	}

	client, err := GetIPFSClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get IPFS client: %w", err)
	}

	reader, err := client.Download().DownloadFile(ctx, parsedCID)
	if err != nil {
		return fmt.Errorf("failed to download content from IPFS: %w", err)
	}
	defer reader.Close()

	downloadedContent, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read downloaded content: %w", err)
	}

	if !bytes.Equal(downloadedContent, []byte(originalContent)) {
		return fmt.Errorf("%s content does not match original: got %d bytes, expected %d bytes",
			contextDesc, len(downloadedContent), len(originalContent))
	}

	return nil
}

// GenerateAndUploadTestFileContent generates test content of specified size, uploads it via portal,
// and waits for the operation to complete. This is a common pattern for testing large file uploads.
// Returns the CID of the uploaded content.
func GenerateAndUploadTestFileContent(ctx context.Context, sizeMB int, filename string) (string, error) {
	sizeBytes := sizeMB * 1024 * 1024
	content := make([]byte, sizeBytes)

	// Fill with random content using crypto/rand for uniqueness
	_, err := rand.Read(content)
	if err != nil {
		return "", fmt.Errorf("failed to generate random content: %w", err)
	}

	// Upload via portal (TUS protocol for large files)
	cid, err := IPFSPortalUpload(ctx, content, filename)
	if err != nil {
		return "", err
	}

	// Wait for operation completion (upload creates operation which creates pin)
	if err := WaitForOperation(ctx, cid); err != nil {
		return "", err
	}

	return cid, nil
}
