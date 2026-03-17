package helpers

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
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

	// Create a filesystem.FS from the file's directory
	dir := filepath.Dir(filePath)
	baseName := filepath.Base(filePath)
	var fsys fs.FS = os.DirFS(dir)

	// Upload using the SDK's UploadFromFS method which handles CAR wrapping automatically
	uploadResult, err := client.Upload().UploadFromFS(ctx, fsys, baseName, nil)
	if err != nil {
		return "", err
	}

	return uploadResult.CID, nil
}
