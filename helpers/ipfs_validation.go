package helpers

import (
	"context"
	"fmt"

	cid "github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

// CIDNotFoundError occurs when CID not found in context
type CIDNotFoundError struct {
	Context string
}

func (e *CIDNotFoundError) Error() string {
	return fmt.Sprintf("no CID found in context (%s)", e.Context)
}

// PinnedNotFoundError occurs when content not pinned
type PinnedNotFoundError struct {
	Context string
	CID     string
}

func (e *PinnedNotFoundError) Error() string {
	return fmt.Sprintf("%s content not pinned for CID %s", e.Context, e.CID)
}

// CIDTooShortError occurs when CID is too short to be valid
type CIDTooShortError struct {
	Context   string
	CID       string
	Length    int
	MinLength int
}

func (e *CIDTooShortError) Error() string {
	return fmt.Sprintf("CID too short for %s: %s (len=%d, min=%d)", e.Context, e.CID, e.Length, e.MinLength)
}

// VerifyCIDPinned checks if CID exists in context and is pinned
func VerifyCIDPinned(ctx context.Context, contextDesc string) error {
	cid, ok := GetCID(ctx)
	if !ok || cid == "" {
		return &CIDNotFoundError{Context: contextDesc}
	}

	pinned, err := IPFSIsPinned(ctx, cid)
	if err != nil {
		return fmt.Errorf("failed to verify %s: %w", contextDesc, err)
	}

	if !pinned {
		return &PinnedNotFoundError{Context: contextDesc, CID: cid}
	}

	return nil
}

// VerifyCIDLength checks CID minimum length
func VerifyCIDLength(ctx context.Context, minLen int, contextDesc string) error {
	cid, ok := GetCID(ctx)
	if !ok || cid == "" {
		return &CIDNotFoundError{Context: contextDesc}
	}

	if len(cid) < minLen {
		return &CIDTooShortError{Context: contextDesc, CID: cid, Length: len(cid), MinLength: minLen}
	}

	return nil
}

// RequireCID retrieves CID from context or returns an error
func RequireCID(ctx context.Context, contextDesc string) (string, error) {
	cid, ok := GetCID(ctx)
	if !ok || cid == "" {
		return "", &CIDNotFoundError{Context: contextDesc}
	}
	return cid, nil
}

// IPFSUploadAndPin uploads content to IPFS and pins it in a single operation.
// This is a common pattern that eliminates duplicate code.
// The filename parameter is optional; when non-empty, it will be stored in the context.
func IPFSUploadAndPin(ctx context.Context, content []byte, filename string) (string, context.Context, error) {
	cid, err := KuboAdd(ctx, content)
	if err != nil {
		return "", ctx, fmt.Errorf("failed to upload file %s: %w", filename, err)
	}

	_, ctx, err = IPFSPinAdd(ctx, cid)
	if err != nil {
		return cid, ctx, fmt.Errorf("failed to pin content %s: %w", filename, err)
	}

	ctx = SetCID(ctx, cid)
	if filename != "" {
		ctx = SetFilename(ctx, filename)
	}

	return cid, ctx, nil
}

// ParseCID parses a CID string into a cid.Cid object
func ParseCID(cidStr string) (cid.Cid, error) {
	c, err := cid.Decode(cidStr)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to parse CID %s: %w", cidStr, err)
	}
	return c, nil
}

// ComputeCID computes a CID from content using v1 format with flexible encoding
// This is for verification purposes during content integrity tests
func ComputeCID(content []byte, codec uint64, mhType uint64, mhLength int) (string, error) {
	// Create a CID using specified parameters with v1 format
	c, err := cid.Prefix{
		Version:  1,
		Codec:    codec,
		MhType:   mhType,
		MhLength: mhLength,
	}.Sum(content)
	if err != nil {
		return "", fmt.Errorf("failed to compute CID from content: %w", err)
	}
	return c.String(), nil
}

// ComputeCIDWithDefaults computes a CID using sensible defaults for UnixFS files
// Uses CIDv1 with SHA2_256 hash and DagProtobuf codec
func ComputeCIDWithDefaults(content []byte) (string, error) {
	return ComputeCID(content, cid.Raw, multihash.SHA2_256, -1)
}

// VerifyDirectoryStructure verifies uploaded directory structure matches expected entries
// This helper validates directory contents including entry names, types (file/directory), and sizes.
// Returns error if verification fails.
func VerifyDirectoryStructure(ctx context.Context, cidStr string, expectedEntries []DirectoryEntry) error {
	parsedCID, err := ParseCID(cidStr)
	if err != nil {
		return fmt.Errorf("failed to parse CID: %w", err)
	}

	client, err := GetIPFSClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get IPFS client: %w", err)
	}

	entries, err := client.Download().ListDirectory(ctx, parsedCID)
	if err != nil {
		return fmt.Errorf("failed to list directory: %w", err)
	}

	if len(entries) != len(expectedEntries) {
		return fmt.Errorf("directory has %d entries, expected %d", len(entries), len(expectedEntries))
	}

	expectedMap := make(map[string]DirectoryEntry)
	for _, entry := range expectedEntries {
		expectedMap[entry.Name] = entry
	}

	for _, entry := range entries {
		name := entry.Name()
		expected, exists := expectedMap[name]
		if !exists {
			return fmt.Errorf("unexpected entry found: %s", name)
		}

		node := entry.Node()
		if node == nil {
			return fmt.Errorf("could not get node for entry %s", name)
		}

		isDir := node.Mode().IsDir()
		if isDir != expected.IsDir {
			return fmt.Errorf("entry %s: type mismatch, got dir=%v, expected dir=%v", name, isDir, expected.IsDir)
		}

		if !isDir {
			size, err := node.Size()
			if err != nil {
				return fmt.Errorf("failed to get size for entry %s: %w", name, err)
			}

			expectedSize := int64(expected.Size)
			if size != expectedSize {
				return fmt.Errorf("entry %s: size mismatch, got %d bytes, expected %d bytes", name, size, expectedSize)
			}
		}
	}

	return nil
}
