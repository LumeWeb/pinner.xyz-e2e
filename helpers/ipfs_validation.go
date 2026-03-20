package helpers

import (
	"context"
	"fmt"
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
	Context  string
	CID      string
	Length   int
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
