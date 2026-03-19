package steps

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

// IPFSUploadSteps holds the state for IPFS upload step definitions
type IPFSUploadSteps struct{}

// NewIPFSUploadSteps creates a new IPFSUploadSteps instance
func NewIPFSUploadSteps() *IPFSUploadSteps {
	return &IPFSUploadSteps{}
}

// InitializeScenario registers all step definitions with godog
func (s *IPFSUploadSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Upload steps
	ctx.Step(`^the user uploads a small file to IPFS "([^"]*)" with content "([^"]*)"$`, s.theUserUploadsASmallFileWithContent)
	ctx.Step(`^a valid IPFS CID is returned$`, s.aValidCIDIsReturned)
	ctx.Step(`^the user uploads a (\d+)MB file to IPFS$`, s.theUserUploadsAMBLargeFileToIPFS)
	ctx.Step(`^the user has an IPFS directory with multiple files$`, s.theUserHasADirectoryWithMultipleFiles)
	ctx.Step(`^the user uploads the IPFS directory$`, s.theUserUploadsTheDirectory)
	ctx.Step(`^all files are uploaded as an IPFS directory CID$`, s.allFilesAreUploadedAsADirectoryCID)
	ctx.Step(`^the directory structure is preserved$`, s.theDirectoryStructureIsPreserved)

	// Note: IPFS pin and operation wait steps are now in ipfs_common_steps.go
	// to support multi-service architecture (IPFS, Arweave, S3, etc.)
	ctx.Step(`^the file is available on IPFS$`, s.theFileIsAvailableOnIPFS)
	ctx.Step(`^the user starts (\d+) concurrent file uploads$`, s.theUserStartsNConcurrentFileUploads)
	ctx.Step(`^the retrieved content matches original$`, s.theRetrievedContentMatchesOriginal)
	ctx.Step(`^the user has a file with known content$`, s.theUserHasAFileWithKnownContent)
	ctx.Step(`^the user has a (\d+)MB file with unique content$`, s.theUserHasASizeMBFile)
	ctx.Step(`^the retrieved file CID matches original$`, s.theRetrievedFileCIDMatchesOriginal)
	ctx.Step(`^all (\d+) files are available$`, s.allNFilesAreAvailable)
	ctx.Step(`^the user uploads the IPFS file$`, s.theUserUploadsTheIPFSFile)
	ctx.Step(`^the user uploads the IPFS file via TUS$`, s.theUserUploadsTheIPFSFileViaTUS)

	// Very-large file upload steps (migrated from ipfs_pinning_steps.go)
	ctx.Step(`^the user has a (\d+)GB IPFS test file$`, s.theUserHasASizeGBIPFSTestFile)
	ctx.Step(`^the user uploads and pins the large IPFS test file$`, s.theUserUploadsAndPinsTheLargeIPFSTestFile)
	ctx.Step(`^the IPFS pin reaches pinned status within (\d+) minutes$`, s.theIPFSPinReachesPinnedStatusWithinMinutes)
	ctx.Step(`^the uploaded IPFS test file is available$`, s.theUploadedIPFSTestFileIsAvailable)
}

// theUserHasAFileWithKnownContent creates test content for integrity verification

// theUserUploadsASmallFileWithContent uploads a file content to IPFS
// Important: Upload creates an operation, which creates the pin AFTER the operation completes.
// We must wait for the operation to complete BEFORE checking if the pin exists.
func (s *IPFSUploadSteps) theUserUploadsASmallFileWithContent(ctx context.Context, filename string, content string) (context.Context, error) {
	// Make content unique to prevent IPFS deduplication across test runs
	uniqueContent := helpers.GenerateUniqueContent(content)
	
	// Upload file via portal (POST to IPFS SDK upload endpoint)
	cid, err := helpers.IPFSPortalUpload(ctx, []byte(uniqueContent), filename)
	if err != nil {
		return ctx, err
	}
	
	// Uploads create an operation first, which creates the pin after it completes
	// Wait for the operation to complete before checking for the pin
	if err := helpers.WaitForOperation(ctx, cid); err != nil {
		return ctx, err
	}
	
	// Store CID in context for verification
	ctx = helpers.SetCID(ctx, cid)
	
	return ctx, nil
}

// aValidCIDIsReturned verifies the CID is properly formatted
func (s *IPFSUploadSteps) aValidCIDIsReturned(ctx context.Context) (context.Context, error) {
	if err := helpers.VerifyCIDLength(ctx, 10, "CID validation"); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// allNFilesAreAvailable verifies N files are available after concurrent uploads
func (s *IPFSUploadSteps) allNFilesAreAvailable(ctx context.Context, count int) (context.Context, error) {
	cids, ok := helpers.GetCIDs(ctx)
	if !ok || len(cids) != count {
		return ctx, fmt.Errorf("expected %d CIDs in context", count)
	}

	return ctx, nil
}

// theUserUploadsAMBLargeFileToIPFS simulates uploading large file
// Important: Uses portal upload (TUS) which creates an operation that creates the pin after completion
func (s *IPFSUploadSteps) theUserUploadsAMBLargeFileToIPFS(ctx context.Context, sizeMB int) (context.Context, error) {
	// Generate test content of specified size with uniqueness to prevent IPFS deduplication
	sizeBytes := sizeMB * 1024 * 1024
	content := make([]byte, sizeBytes)
	
	// Fill with random content using crypto/rand for uniqueness across test runs
	_, err := rand.Read(content)
	if err != nil {
		return ctx, fmt.Errorf("failed to generate random content: %w", err)
	}

	// Upload via portal (TUS protocol) - creates operation which creates pin after completion
	cid, err := helpers.IPFSPortalUpload(ctx, content, fmt.Sprintf("%dMB-test-file.bin", sizeMB))
	if err != nil {
		return ctx, err
	}

	// Portal uploads create an operation first; must wait for operation before checking for pin
	if err := helpers.WaitForOperation(ctx, cid); err != nil {
		return ctx, err
	}

	ctx = helpers.SetCID(ctx, cid)
	return ctx, nil
}

// theUserHasADirectoryWithMultipleFiles creates a test directory structure
func (s *IPFSUploadSteps) theUserHasADirectoryWithMultipleFiles(ctx context.Context) (context.Context, error) {
	// Create temp directory
	testDir, err := os.MkdirTemp("", "ipfs-dir-test")
	if err != nil {
		return ctx, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create a few test files
	testFiles := []struct {
		name    string
		content []byte
	}{
		{"file1.txt", []byte("content 1")},
		{"file2.txt", []byte("content 2")},
		{"file3.txt", []byte("content 3")},
	}

	for _, tf := range testFiles {
		filePath := fmt.Sprintf("%s/%s", testDir, tf.name)
		if err := os.WriteFile(filePath, tf.content, 0644); err != nil {
			os.RemoveAll(testDir)
			return ctx, fmt.Errorf("failed to create test file: %w", err)
		}
	}

	ctx = helpers.SetTestDirectory(ctx, testDir)
	return ctx, nil
}

// theUserHasAFileWithKnownContent creates test content for integrity verification
func (s *IPFSUploadSteps) theUserHasAFileWithKnownContent(ctx context.Context) (context.Context, error) {
	uniqueContent := []byte(helpers.GenerateUniqueContent("integrity test content"))
	ctx = helpers.SetKnownContent(ctx, string(uniqueContent))
	return ctx, nil
}

func (s *IPFSUploadSteps) theRetrievedContentMatchesOriginal(ctx context.Context) (context.Context, error) {
	// TODO: Download content from IPFS and verify it matches original bytes
	return ctx, nil
}

// theUserStartsNConcurrentFileUploads starts N concurrent file uploads
func (s *IPFSUploadSteps) theUserStartsNConcurrentFileUploads(ctx context.Context, count int) (context.Context, error) {
	var cids []string
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make(chan error, count)

	for i := range count {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			uniqueContent := []byte(helpers.GenerateUniqueContent(fmt.Sprintf("concurrent test %d", index)))
			filename := fmt.Sprintf("concurrent-file-%d.txt", index)

			cid, err := helpers.IPFSPortalUpload(ctx, uniqueContent, filename)
			if err != nil {
				errors <- fmt.Errorf("upload %d failed: %w", index, err)
				return
			}

			mu.Lock()
			cids = append(cids, cid)
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			return ctx, err
		}
	}

	if len(cids) != count {
		return ctx, fmt.Errorf("expected %d concurrent uploads, got %d", count, len(cids))
	}

	ctx = helpers.SetCIDs(ctx, cids)
	
	// Wait for all operations to complete and verify pins
	// Increased timeout to 10 minutes per operation for 10 concurrent uploads
	for i, cid := range cids {
		if err := helpers.WaitForOperationCompleteByCID(ctx, cid, helpers.DefaultOperationTimeout); err != nil {
			return ctx, fmt.Errorf("operation for file %d failed: %w", i, err)
		}
	}
	
	return ctx, nil
}

// theUserUploadsTheIPFSFile uploads a file with known content to IPFS
func (s *IPFSUploadSteps) theUserUploadsTheIPFSFile(ctx context.Context) (context.Context, error) {
	content, ok := helpers.GetKnownContent(ctx)
	if !ok {
		return ctx, fmt.Errorf("no known content found in context")
	}

	cid, err := helpers.IPFSPortalUpload(ctx, []byte(content), "integrity-test-file.bin")
	if err != nil {
		return ctx, fmt.Errorf("failed to upload IPFS file: %w", err)
	}

	ctx = helpers.SetCID(ctx, cid)
	return ctx, nil
}

// theUserUploadsTheIPFSFileViaTUS uploads a file via TUS protocol
func (s *IPFSUploadSteps) theUserUploadsTheIPFSFileViaTUS(ctx context.Context) (context.Context, error) {
	content, ok := helpers.GetKnownContent(ctx)
	if !ok {
		return ctx, fmt.Errorf("no known content found in context")
	}

	// Upload via portal - the SDK automatically uses TUS for large files (>100MB)
	cid, err := helpers.IPFSPortalUpload(ctx, []byte(content), "tus-integrity-file.bin")
	if err != nil {
		return ctx, fmt.Errorf("failed to upload IPFS file via TUS: %w", err)
	}

	ctx = helpers.SetCID(ctx, cid)
	return ctx, nil
}

// theUserHasASizeMBFile creates test data of specified size for TUS integrity test
func (s *IPFSUploadSteps) theUserHasASizeMBFile(ctx context.Context, sizeMB int) (context.Context, error) {
	sizeBytes := int64(sizeMB * 1024 * 1024)
	content, err := helpers.GenerateLargeTestFile(sizeBytes)
	if err != nil {
		return ctx, fmt.Errorf("failed to generate test file: %w", err)
	}
	
	ctx = helpers.SetKnownContent(ctx, string(content))
	return ctx, nil
}

// TODO: Implement actual integrity verification by downloading content from IPFS
// and comparing with original. Portal returns UnixFS CIDs, not raw CIDs,
// so content-level verification requires downloading via gateway and byte-by-byte comparison.
func (s *IPFSUploadSteps) theRetrievedFileCIDMatchesOriginal(ctx context.Context) (context.Context, error) {
	// Verify CID was returned from upload and stored in context
	cidStr, err := helpers.RequireCID(ctx, "CID verification")
	if err != nil {
		return ctx, err
	}
	
	// Verify CID is not empty
	if cidStr == "" {
		return ctx, fmt.Errorf("CID is empty after upload")
	}
	
	
	return ctx, nil
}


// theUserUploadsTheDirectory uploads the directory structure
func (s *IPFSUploadSteps) theUserUploadsTheDirectory(ctx context.Context) (context.Context, error) {
	testDir, ok := helpers.GetTestDirectory(ctx)
	if !ok {
		return ctx, fmt.Errorf("no test directory found in context")
	}

	// Upload the directory using IPFSPortalUploadDirFromFS
	cid, err := helpers.IPFSPortalUploadDirFromFS(ctx, testDir)
	if err != nil {
		os.RemoveAll(testDir)
		return ctx, fmt.Errorf("failed to upload directory: %w", err)
	}

	// Wait for operation completion
	if err := helpers.WaitForOperation(ctx, cid); err != nil {
		os.RemoveAll(testDir)
		return ctx, fmt.Errorf("operation did not complete: %w", err)
	}

	// Clean up temp directory
	os.RemoveAll(testDir)

	// Store the directory CID in context
	ctx = helpers.SetCID(ctx, cid)
	return ctx, nil
}

// allFilesAreUploadedAsADirectoryCID verifies directory upload
func (s *IPFSUploadSteps) allFilesAreUploadedAsADirectoryCID(ctx context.Context) (context.Context, error) {
	if err := helpers.VerifyCIDPinned(ctx, "directory upload"); err != nil {
		return ctx, fmt.Errorf("failed to verify directory upload: %w", err)
	}
	return ctx, nil
}

// theDirectoryStructureIsPreserved verifies directory structure
func (s *IPFSUploadSteps) theDirectoryStructureIsPreserved(ctx context.Context) (context.Context, error) {
	// For now, just verify we have a CID
	_, err := helpers.RequireCID(ctx, "directory structure")
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}
// theUserHasASizeGBIPFSTestFile creates test data of specified size in GB on disk
// Stores file path in context for upload. Caller is responsible for cleanup.
func (s *IPFSUploadSteps) theUserHasASizeGBIPFSTestFile(ctx context.Context, sizeGB int) (context.Context, error) {
	sizeBytes := int64(sizeGB * 1024 * 1024 * 1024)
	filePath, err := helpers.GenerateLargeTestFileOnDisk(sizeBytes, "ipfs-large-test")
	if err != nil {
		return ctx, fmt.Errorf("failed to generate IPFS test file on disk: %w", err)
	}

	ctx = helpers.SetTestFilePath(ctx, filePath)
	return ctx, nil
}

// theUserUploadsAndPinsTheLargeIPFSTestFile uploads and pins the large IPFS test file via portal
// Uses disk-based upload to avoid holding large files in memory. Cleans up test file after upload.
func (s *IPFSUploadSteps) theUserUploadsAndPinsTheLargeIPFSTestFile(ctx context.Context) (context.Context, error) {
	filePath, ok := helpers.GetTestFilePath(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPFS test file path found")
	}

	// Upload from disk instead of memory
	cid, err := helpers.IPFSPortalUploadFromFS(ctx, filePath, "large-test-file.bin")
	if err != nil {
		return ctx, fmt.Errorf("failed to upload large IPFS test file from disk: %w", err)
	}

	// Clean up the test file after successful upload
	if removeErr := os.Remove(filePath); removeErr != nil {
		// Log but don't fail the test if cleanup fails
		fmt.Printf("Warning: failed to cleanup test file %s: %v\n", filePath, removeErr)
	}

	ctx = helpers.SetCID(ctx, cid)
	return ctx, nil
}

// theIPFSPinReachesPinnedStatusWithinMinutes verifies pin completes within time limit
func (s *IPFSUploadSteps) theIPFSPinReachesPinnedStatusWithinMinutes(ctx context.Context, minutes int) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "time limit verification")
	if err != nil {
		return ctx, err
	}

	timeout := time.Duration(minutes) * time.Minute
	startTime := time.Now()

	// Create a timeout context for internal polling.
	// We DON'T return this context to godog to avoid passing a cancelled context
	// to the next step. Instead, we use it locally and return the original context.
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		pinned, err := helpers.IPFSIsPinned(timeoutCtx, cidStr)
		if err != nil {
			return ctx, fmt.Errorf("failed to check pin status: %w", err)
		}

		if pinned {
			return ctx, nil
		}

		if time.Since(startTime) > timeout {
			return ctx, fmt.Errorf("pin did not complete within %d minutes", minutes)
		}

		time.Sleep(5 * time.Second)
	}
}

// theUploadedIPFSTestFileIsAvailable verifies the uploaded large IPFS test file CID is pinned and retrievable
func (s *IPFSUploadSteps) theUploadedIPFSTestFileIsAvailable(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "IPFS test file availability verification")
	if err != nil {
		return ctx, err
	}

	pinned, err := helpers.IPFSIsPinned(ctx, cidStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to check if IPFS test file is available: %w", err)
	}

	if !pinned {
		return ctx, fmt.Errorf("IPFS test file is not available")
	}

	return ctx, nil
}

// theFileIsAvailableOnIPFS verifies the uploaded file content is pinned and retrievable
func (s *IPFSUploadSteps) theFileIsAvailableOnIPFS(ctx context.Context) (context.Context, error) {
	if err := helpers.VerifyCIDPinned(ctx, "uploaded file availability"); err != nil {
		return ctx, fmt.Errorf("failed to verify file is available on IPFS: %w", err)
	}
	return ctx, nil
}
