package helpers

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"go.lumeweb.com/ipfs-content/car"
	ipfs "go.lumeweb.com/ipfs-sdk"
	ipfs_sdk_fs "go.lumeweb.com/ipfs-sdk/fs"
)

// containsErrorMessage checks if an error or any wrapped error contains the specified message.
// This is useful for detecting error types when errors are wrapped multiple times.
func containsErrorMessage(err error, msg string) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), msg)
}

// IsQuotaEnforcementError checks if an error is caused by quota enforcement.
// Returns true if the error matches ipfs.ErrRateLimitExceeded or if the error message
// indicates a partial download failure (e.g., "failed to fetch all nodes") or
// quota exceeded errors from the portal.
func IsQuotaEnforcementError(err error) bool {
	if errors.Is(err, ipfs.ErrRateLimitExceeded) {
		return true
	}

	// Check for partial download failures that occur when quota enforcement
	// blocks gateway access, preventing all blocks from being fetched
	// We check both direct errors and wrapped errors that contain this message
	if err != nil && containsErrorMessage(err, "failed to fetch all nodes") {
		return true
	}

	// Check for portal quota exceeded errors
	if err != nil {
		errorMsg := err.Error()
		// Check for quota exceeded messages (case-insensitive substring matching)
		if strings.Contains(strings.ToLower(errorMsg), "quota exceeded") {
			return true
		}
		// Check for HTTP 429 rate limit status code (case-insensitive)
		if strings.Contains(strings.ToLower(errorMsg), "http 429") {
			return true
		}
	}

	return false
}

// IPFSPortalUpload uploads content to the portal via the IPFS SDK upload endpoint.
// This creates an account operation and returns the CID and DAG size.
// The caller must wait for operation completion using WaitForOperation.
// This is for testing user uploads to IPFS (POST to portal), NOT for pinning existing CIDs.
// The content is wrapped in CAR format automatically by the SDK.
func IPFSPortalUpload(ctx context.Context, content []byte, filename string) (string, int64, error) {
	// Get IPFS SDK client
	client, err := GetIPFSClient(ctx)
	if err != nil {
		return "", 0, err
	}

	// Upload using the SDK's UploadBytes method which handles CAR wrapping automatically
	uploadResult, err := client.Upload().UploadBytes(ctx, content, filename, nil)
	if err != nil {
		return "", 0, err
	}

	dagSize := uploadResult.DAGSize
	return uploadResult.CID, dagSize, nil
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
		// Enhance error message for quota enforcement errors
		if IsQuotaEnforcementError(err) {
			return fmt.Errorf("failed to read downloaded content: %w "+
				"(quota enforcement causing partial download failures)", err)
		}
		return fmt.Errorf("failed to read downloaded content: %w", err)
	}

	if !bytes.Equal(downloadedContent, []byte(originalContent)) {
		return fmt.Errorf("%s content does not match original: got %d bytes, expected %d bytes",
			contextDesc, len(downloadedContent), len(originalContent))
	}

	return nil
}

// CalculateDAGSizeFromFileContent calculates the DAG size for file content without uploading it.
// This is useful for setting exact quota limits before performing an upload.
// The content is prepared using the SDK's CAR preparation logic to ensure the DAG size
// matches what would be calculated during actual upload.
//
// This function performs the same CAR preparation that UploadBytes uses internally:
// - Parses the content as a single file
// - Wraps it in a CAR format
// - Calculates the total DAG size (all blocks)
//
// Returns the DAG size in bytes that would be used for quota tracking during upload.
func CalculateDAGSizeFromFileContent(ctx context.Context, content []byte) (int64, error) {
	// Create an in-memory filesystem containing the file
	filename := "file.bin"
	fsys := ipfs_sdk_fs.NewBytesFS(content, filename)

	// Prepare CAR without uploading (matches SDK's PrepareCAR logic)
	_, summary, err := car.PrepareCAR(ctx, fsys, false)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare CAR: %w", err)
	}

	// Return the DAG size
	return int64(summary.TotalSize), nil
}

// GenerateAndUploadTestFileContent generates test content of specified size, uploads it via portal,
// and waits for the operation to complete. This is a common pattern for testing large file uploads.
// Returns the CID and DAG size of the uploaded content.
func GenerateAndUploadTestFileContent(ctx context.Context, sizeMB int, filename string) (string, int64, error) {
	size := int64(sizeMB * 1024 * 1024)
	content := make([]byte, size)

	// Fill with random content using crypto/rand for uniqueness
	_, err := rand.Read(content)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate random content: %w", err)
	}

	// Upload via portal (TUS protocol for large files)
	cid, dagSize, err := IPFSPortalUpload(ctx, content, filename)
	if err != nil {
		return "", 0, err
	}

	// Wait for operation completion (upload creates operation which creates pin)
	if err := WaitForOperation(ctx, cid); err != nil {
		return "", 0, err
	}

	return cid, dagSize, nil
}
