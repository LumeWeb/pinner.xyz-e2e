package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

// IPFSCommonSteps holds shared step definitions for IPFS operations
type IPFSCommonSteps struct{}

// NewIPFSCommonSteps creates a new IPFSCommonSteps instance
func NewIPFSCommonSteps() *IPFSCommonSteps {
	return &IPFSCommonSteps{}
}

// InitializeScenario registers all shared IPFS step definitions with godog
func (s *IPFSCommonSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Shared wait/verification steps for IPFS operations
	ctx.Step(`^the IPFS pin reaches pinned status$`, s.theIPFSPinReachesPinnedStatus)
	ctx.Step(`^the operation completes$`, s.theOperationCompletes)
	
	// Plural versions for multi CID operations (e.g., concurrent uploads)
	ctx.Step(`^all IPFS pins reach pinned status$`, s.allIPFSPinsReachPinnedStatus)
	ctx.Step(`^all operations complete$`, s.allOperationsComplete)
	
	// Content verification steps
	ctx.Step(`^the IPFS content is retrievable$`, s.theIPFSContentIsRetrievable)
	ctx.Step(`^the IPFS content matches original$`, s.theIPFSContentMatchesOriginal)
	ctx.Step(`^the IPFS file size matches original$`, s.theIPFSFileSizeMatchesOriginal)
	ctx.Step(`^the directory structure is preserved$`, s.theDirectoryStructureIsPreserved)
	
	// Plural versions for multi CID operations
	ctx.Step(`^all (\d+) IPFS files are retrievable$`, s.allIPFSFilesAreRetrievable)
}

// theIPFSPinReachesPinnedStatus waits for the IPFS pin status to reach StatusPinned
func (s *IPFSCommonSteps) theIPFSPinReachesPinnedStatus(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "IPFS pinning status")
	if err != nil {
		return ctx, err
	}

	if err := helpers.WaitForPinCreation(ctx, cidStr); err != nil {
		return ctx, fmt.Errorf("IPFS pin did not reach pinned status: %w", err)
	}

	return ctx, nil
}

// theDirectoryStructureIsPreserved verifies the uploaded directory structure matches the expected structure
func (s *IPFSCommonSteps) theDirectoryStructureIsPreserved(ctx context.Context) (context.Context, error) {
	// Get CID from context
	cidStr, err := helpers.RequireCID(ctx, "directory structure")
	if err != nil {
		return ctx, err
	}

	// Get expected directory entries from context
	expectedEntries, ok := helpers.GetDirectoryEntries(ctx)
	if !ok {
		return ctx, fmt.Errorf("no directory entries found in context")
	}

	// Use helper to verify directory structure
	if err := helpers.VerifyDirectoryStructure(ctx, cidStr, expectedEntries); err != nil {
		return ctx, err
	}

	return ctx, nil
}

// allIPFSFilesAreRetrievable verifies all uploaded files can be downloaded
// Used for concurrent uploads where multiple files need verification
func (s *IPFSCommonSteps) allIPFSFilesAreRetrievable(ctx context.Context, count int) (context.Context, error) {
	cids, ok := helpers.GetCIDs(ctx)
	if !ok || len(cids) != count {
		return ctx, fmt.Errorf("expected %d CIDs in context, got %d", count, len(cids))
	}

	client, err := helpers.GetIPFSClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get IPFS client: %w", err)
	}

	// Verify each file is retrievable
	for i, cidStr := range cids {
		parsedCID, err := helpers.ParseCID(cidStr)
		if err != nil {
			return ctx, fmt.Errorf("failed to parse CID %s: %w", cidStr, err)
		}

		has, err := client.Download().Has(ctx, parsedCID)
		if err != nil {
			return ctx, fmt.Errorf("failed to check if file %d exists: %w", i, err)
		}

		if !has {
			return ctx, fmt.Errorf("file %d with CID %s is not retrievable", i, cidStr)
		}
	}

	return ctx, nil
}

// theOperationCompletes waits for the account operation to reach StatusCompleted
// Note: Operations are global (use account service - not service-specific)
func (s *IPFSCommonSteps) theOperationCompletes(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "operation")
	if err != nil {
		return ctx, err
	}

	if err := helpers.WaitForOperation(ctx, cidStr); err != nil {
		return ctx, fmt.Errorf("operation did not complete: %w", err)
	}

	return ctx, nil
}

// allIPFSPinsReachPinnedStatus waits for all IPFS pins in the context to reach StatusPinned
func (s *IPFSCommonSteps) allIPFSPinsReachPinnedStatus(ctx context.Context) (context.Context, error) {
	cids, ok := helpers.GetCIDs(ctx)
	if !ok {
		// Fall back to single CID for backward compatibility
		cidStr, err := helpers.RequireCID(ctx, "IPFS pinning status")
		if err != nil {
			return ctx, fmt.Errorf("no CIDs found in context and single CID missing: %w", err)
		}
		
		if err := helpers.WaitForPinCreation(ctx, cidStr); err != nil {
			return ctx, fmt.Errorf("IPFS pin did not reach pinned status: %w", err)
		}
		return ctx, nil
	}

	for i, cid := range cids {
		if err := helpers.WaitForPinCreation(ctx, cid); err != nil {
			return ctx, fmt.Errorf("IPFS pin %d (%s) did not reach pinned status: %w", i, cid, err)
		}
	}

	return ctx, nil
}

// allOperationsComplete waits for all operations in the context to reach StatusCompleted
func (s *IPFSCommonSteps) allOperationsComplete(ctx context.Context) (context.Context, error) {
	cids, ok := helpers.GetCIDs(ctx)
	if !ok {
		// Fall back to single CID for backward compatibility
		cidStr, err := helpers.RequireCID(ctx, "operation")
		if err != nil {
			return ctx, fmt.Errorf("no CIDs found in context and single CID missing: %w", err)
		}
		
		if err := helpers.WaitForOperation(ctx, cidStr); err != nil {
			return ctx, fmt.Errorf("operation did not complete: %w", err)
		}
		return ctx, nil
	}

	for i, cid := range cids {
		if err := helpers.WaitForOperation(ctx, cid); err != nil {
			return ctx, fmt.Errorf("operation %d failed to complete for CID %s: %w", i, cid, err)
		}
	}

	return ctx, nil
}

// theIPFSContentIsRetrievable verifies IPFS content can be downloaded
// This is a lightweight check that verifies the CID is accessible without downloading full content
func (s *IPFSCommonSteps) theIPFSContentIsRetrievable(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "content retrieval")
	if err != nil {
		return ctx, err
	}

	parsedCID, err := helpers.ParseCID(cidStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse CID: %w", err)
	}

	client, err := helpers.GetIPFSClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get IPFS client: %w", err)
	}

	// Check if content exists
	has, err := client.Download().Has(ctx, parsedCID)
	if err != nil {
		return ctx, fmt.Errorf("failed to check if content exists: %w", err)
	}

	if !has {
		return ctx, fmt.Errorf("content with CID %s is not retrievable", cidStr)
	}

	return ctx, nil
}

// theIPFSContentMatchesOriginal verifies IPFS content matches the original byte-for-byte
// Used for small files where content is stored in context
func (s *IPFSCommonSteps) theIPFSContentMatchesOriginal(ctx context.Context) (context.Context, error) {
	// Get CID from context
	cidStr, err := helpers.RequireCID(ctx, "content verification")
	if err != nil {
		return ctx, err
	}

	// Get original content from context
	originalContent, ok := helpers.GetKnownContent(ctx)
	if !ok {
		return ctx, fmt.Errorf("no original content found in context")
	}

	// Use helper to download and verify content
	if err := helpers.DownloadAndVerifyContent(ctx, cidStr, originalContent, "downloaded"); err != nil {
		return ctx, err
	}

	return ctx, nil
}

// theIPFSFileSizeMatchesOriginal verifies IPFS file size matches the original
// Used for large files where full content verification is impractical
func (s *IPFSCommonSteps) theIPFSFileSizeMatchesOriginal(ctx context.Context) (context.Context, error) {
	// Get CID from context
	cidStr, err := helpers.RequireCID(ctx, "size verification")
	if err != nil {
		return ctx, err
	}

	// Get original size from context
	expectedSize, ok := helpers.GetFileSize(ctx)
	if !ok {
		return ctx, fmt.Errorf("no file size found in context")
	}

	// Parse the CID
	parsedCID, err := helpers.ParseCID(cidStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse CID: %w", err)
	}

	// Get IPFS client
	client, err := helpers.GetIPFSClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get IPFS client: %w", err)
	}

	// Get file size from IPFS as int64 and convert to int
	downloadedSize64, err := client.Download().FileSize(ctx, parsedCID)
	if err != nil {
		return ctx, fmt.Errorf("failed to get file size from IPFS: %w", err)
	}

	downloadedSize := int(downloadedSize64)

	// Verify size matches
	if downloadedSize != expectedSize {
		return ctx, fmt.Errorf("downloaded file size %d bytes does not match expected %d bytes", downloadedSize, expectedSize)
	}

	return ctx, nil
}
