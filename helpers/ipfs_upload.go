package helpers

import (
	"context"
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
