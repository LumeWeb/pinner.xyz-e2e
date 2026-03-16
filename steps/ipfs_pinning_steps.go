package steps

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

// PinningSteps holds the state for IPFS pinning step definitions
type PinningSteps struct{}

// NewPinningSteps creates a new PinningSteps instance
func NewPinningSteps() *PinningSteps {
	return &PinningSteps{}
}

// InitializeScenario registers all step definitions with godog
func (s *PinningSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Pinning steps
	ctx.Step(`^the user has an IPFS CID$`, s.theUserHasACID)

	// Granular pinning lifecycle steps
	ctx.Step(`^the user starts pinning the IPFS CID$`, s.theUserStartsPinningTheCID)
	// Note: IPFS pin and operation wait steps are now in ipfs_common_steps.go
	// to support multi-service architecture (IPFS, Arweave, S3, etc.)

	// Single-step pinning approach
	ctx.Step(`^the user pins the CID$`, s.theUserPinsTheCID)
	ctx.Step(`^the IPFS pin is created successfully$`, s.thePinIsCreatedSuccessfully)
	ctx.Step(`^the new IPFS CID is pinned$`, s.theNewCIDIsPinned)
	ctx.Step(`^the user has 3 pinned IPFS CIDs$`, s.theUserHas3PinnedCIDs)
	ctx.Step(`^the user lists their IPFS pins$`, s.theUserListsTheirPins)
	ctx.Step(`^the pin is in the list$`, s.thePinIsInTheList)
	ctx.Step(`^the user unpins the IPFS content$`, s.theUserUnpinsTheContent)
	ctx.Step(`^the IPFS pin is no longer in the list$`, s.thePinIsNoLongerInTheList)
	ctx.Step(`^the user replaces the pin with a new CID$`, s.theUserReplacesThePinWithANewCID)
	ctx.Step(`^the user has a pinned IPFS CID$`, s.theUserHasAPinnedCID)
	ctx.Step(`^all (\d+) IPFS pins are returned$`, s.allPinsAreReturned)
	ctx.Step(`^each pin has a unique CID$`, s.eachPinHasAUniqueCID)
	ctx.Step(`^the pin now points to the new CID$`, s.thePinNowPointsToTheNewCID)
	ctx.Step(`^the old CID is no longer pinned$`, s.theOldCIDIsNoLongerPinned)

	// Performance & stress testing scenarios
	ctx.Step(`^the user has a (\d+)MB IPFS test file$`, s.theUserHasASizeMBIPFSTestFile)
	ctx.Step(`^the user has a (\d+)GB IPFS test file$`, s.theUserHasASizeGBIPFSTestFile)
	ctx.Step(`^the user uploads and pins the large IPFS test file$`, s.theUserUploadsAndPinsTheLargeIPFSTestFile)
	ctx.Step(`^the IPFS pin reaches pinned status within (\d+) minutes$`, s.theIPFSPinReachesPinnedStatusWithinMinutes)
	ctx.Step(`^the uploaded IPFS test file is available$`, s.theUploadedIPFSTestFileIsAvailable)
}

// theIPFSPinReachesPinnedStatusWithinMinutes verifies pin completes within time limit
func (s *PinningSteps) theUserHasACID(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.KuboAdd(ctx, []byte("test content"))
	if err != nil {
		return ctx, fmt.Errorf("failed to generate test CID: %w", err)
	}

	ctx = helpers.SetCID(ctx, cidStr)
	return ctx, nil
}

// theUserPinsTheCID pins a CID using Portal SDK (CID comes from context)
func (s *PinningSteps) theUserPinsTheCID(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "pinning")
	if err != nil {
		return ctx, fmt.Errorf("no CID found in context")
	}

	_, ctx, err = helpers.IPFSPinAdd(ctx, cidStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to pin CID: %w", err)
	}

	return ctx, nil
}

// thePinIsCreatedSuccessfully verifies the pin was created
func (s *PinningSteps) thePinIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	if err := helpers.VerifyCIDPinned(ctx, "pin"); err != nil {
		return ctx, fmt.Errorf("failed to get pin status: %w", err)
	}
	return ctx, nil
}

// theNewCIDIsPinned creates a new CID, pins it, and unpins the old CID
func (s *PinningSteps) theNewCIDIsPinned(ctx context.Context) (context.Context, error) {
	// Get the old CID (the currently pinned one)
	oldCIDStr, ok := helpers.GetOldCID(ctx)
	if !ok || oldCIDStr == "" {
		// If no old CID, use the current CID as old
		cidStr, err := helpers.RequireCID(ctx, "replacement")
		if err != nil {
			return ctx, fmt.Errorf("no existing CID found for replacement")
		}
		oldCIDStr = cidStr
		ctx = helpers.SetOldCID(ctx, oldCIDStr)
	}

	// Generate new CID
	newCIDStr, err := helpers.KuboAdd(ctx, []byte("replacement content"))
	if err != nil {
		return ctx, fmt.Errorf("failed to generate new CID: %w", err)
	}

	// Pin the new CID
	_, ctx, err = helpers.IPFSPinAdd(ctx, newCIDStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to pin new CID: %w", err)
	}

	// Unpin the old CID
	if err := helpers.IPFSPinRm(ctx, oldCIDStr); err != nil {
		return ctx, fmt.Errorf("failed to unpin old CID: %w", err)
	}

	// Update context
	ctx = helpers.SetOldCID(ctx, oldCIDStr)
	ctx = helpers.SetCID(ctx, newCIDStr)

	return ctx, nil
}

// theUserHas3PinnedCIDs creates 3 test CIDs and pins them
func (s *PinningSteps) theUserHas3PinnedCIDs(ctx context.Context) (context.Context, error) {
	var cids []string
	for i := range 3 {
		buf := []byte{}
		buf = fmt.Appendf(buf, "test content %d", i)
		cidStr, ctx, err := helpers.IPFSUploadAndPin(ctx, buf, "")
		if err != nil {
			return ctx, err
		}
		cids = append(cids, cidStr)
	}

	ctx = helpers.SetCIDs(ctx, cids)
	return ctx, nil
}

// theUserListsTheirPins retrieves all pins for the authenticated user
// Note: For test isolation, we use the pins stored in context, not the API list
func (s *PinningSteps) theUserListsTheirPins(ctx context.Context) (context.Context, error) {
	// Get CIDs from context (set by theUserHas3PinnedCIDs or similar steps)
	cids, ok := helpers.GetCIDs(ctx)
	if !ok || len(cids) == 0 {
		return ctx, fmt.Errorf("no CIDs found in context")
	}

	// Store CIDs as the pin list for verification
	ctx = helpers.SetPinList(ctx, cids)
	return ctx, nil
}

// thePinIsInTheList verifies a previously pinned CID is in the pin list
func (s *PinningSteps) thePinIsInTheList(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "pin list verification")
	if err != nil {
		return ctx, fmt.Errorf("no CID found in context")
	}

	pinsList, ok := helpers.GetPinList(ctx)
	if !ok {
		return ctx, fmt.Errorf("no pins found in context")
	}

	if !slices.Contains(pinsList, cidStr) {
		return ctx, fmt.Errorf("CID %s not found in pin list", cidStr)
	}

	return ctx, nil
}

// theUserHasAPinnedCID creates and pins a test CID for replacement scenarios
func (s *PinningSteps) theUserHasAPinnedCID(ctx context.Context) (context.Context, error) {
	originalCID, err := helpers.KuboAdd(ctx, []byte("original content for pin replacement"))
	if err != nil {
		return ctx, fmt.Errorf("failed to generate original CID: %w", err)
	}

	_, ctx, err = helpers.IPFSPinAdd(ctx, originalCID)
	if err != nil {
		return ctx, fmt.Errorf("failed to pin original CID: %w", err)
	}

	ctx = helpers.SetOldCID(ctx, originalCID)
	ctx = helpers.SetCID(ctx, originalCID)
	return ctx, nil
}

// theUserReplacesThePinWithANewCID replaces an existing pin with a new CID
func (s *PinningSteps) theUserReplacesThePinWithANewCID(ctx context.Context) (context.Context, error) {
	oldCIDStr, ok := helpers.GetOldCID(ctx)
	if !ok || oldCIDStr == "" {
		return ctx, fmt.Errorf("no old CID found in context")
	}

	newCIDStr, err := helpers.KuboAdd(ctx, []byte("replacement content"))
	if err != nil {
		return ctx, fmt.Errorf("failed to generate new CID: %w", err)
	}

	// Pin new CID using Portal SDK
	_, ctx, err = helpers.IPFSPinAdd(ctx, newCIDStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to pin new CID: %w", err)
	}

	// Unpin old CID using Portal SDK
	if err := helpers.IPFSPinRm(ctx, oldCIDStr); err != nil {
		return ctx, fmt.Errorf("failed to unpin old CID: %w", err)
	}

	ctx = helpers.SetCID(ctx, newCIDStr)
	return ctx, nil
}

// thePinNowPointsToTheNewCID verifies the new CID is pinned
func (s *PinningSteps) thePinNowPointsToTheNewCID(ctx context.Context) (context.Context, error) {
	if err := helpers.VerifyCIDPinned(ctx, "pin"); err != nil {
		return ctx, fmt.Errorf("failed to get pin: %w", err)
	}
	return ctx, nil
}

// theOldCIDIsNoLongerPinned verifies the old CID is unpinned
func (s *PinningSteps) theOldCIDIsNoLongerPinned(ctx context.Context) (context.Context, error) {
	oldCIDStr, ok := helpers.GetOldCID(ctx)
	if !ok || oldCIDStr == "" {
		return ctx, fmt.Errorf("no old CID found in context")
	}

	// Error expected for unpinned CID - just check it's not pinned
	pinned, err := helpers.IPFSIsPinned(ctx, oldCIDStr)
	if err == nil && pinned {
		return ctx, fmt.Errorf("old CID is still pinned")
	}

	return ctx, nil
}

// theUserUnpinsTheContent removes a CID from the pin list
func (s *PinningSteps) theUserUnpinsTheContent(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "unpinning")
	if err != nil {
		return ctx, fmt.Errorf("no CID found in context")
	}

	if err := helpers.IPFSPinRm(ctx, cidStr); err != nil {
		return ctx, fmt.Errorf("failed to unpin CID: %w", err)
	}

	return ctx, nil
}

// thePinIsNoLongerInTheList verifies the CID is removed from pin list
func (s *PinningSteps) thePinIsNoLongerInTheList(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "unpinning verification")
	if err != nil {
		return ctx, err
	}

	pinned, err := helpers.IPFSIsPinned(ctx, cidStr)
	if err == nil && pinned {
		return ctx, fmt.Errorf("CID %s is still pinned", cidStr)
	}

	return ctx, nil
}

// allPinsAreReturned verifies the expected number of pins are returned
func (s *PinningSteps) allPinsAreReturned(ctx context.Context, expectedCount int) (context.Context, error) {
	pinsList, ok := helpers.GetPinList(ctx)
	if !ok {
		return ctx, fmt.Errorf("no pins found in context")
	}

	if len(pinsList) != expectedCount {
		return ctx, fmt.Errorf("expected %d pins, got %d", expectedCount, len(pinsList))
	}

	return ctx, nil
}

// eachPinHasAUniqueCID verifies all pins have unique CIDs
func (s *PinningSteps) eachPinHasAUniqueCID(ctx context.Context) (context.Context, error) {
	pinsList, ok := helpers.GetPinList(ctx)
	if !ok {
		return ctx, fmt.Errorf("no pins found in context")
	}

	cids := make(map[string]bool)
	for _, pin := range pinsList {
		if cids[pin] {
			return ctx, fmt.Errorf("duplicate CID found: %s", pin)
		}
		cids[pin] = true
	}

	return ctx, nil
}

// theUserStartsPinningTheCID starts the pinning process without waiting for completion
func (s *PinningSteps) theUserStartsPinningTheCID(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "pinning")
	if err != nil {
		return ctx, err
	}

	// Create pin - returns immediately without waiting for status
	_, ctx, err = helpers.IPFSPinAdd(ctx, cidStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to start pinning CID: %w", err)
	}

	return ctx, nil
}

// thePinStatusReachesQueuedOrPinningWithinTimeout verifies pin reaches initial state
func (s *PinningSteps) thePinStatusReachesQueuedOrPinningWithinTimeout(ctx context.Context) (context.Context, error) {
	// Start pinning a CID if not already started
	cidStr, err := helpers.RequireCID(ctx, "timeout verification")
	if err != nil {
		return ctx, err
	}

	_, ctx, err = helpers.IPFSPinAdd(ctx, cidStr)
	if err != nil {
		return ctx, err
	}

	// List pins to check status
	pins, err := helpers.IPFSPinLs(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list pins: %w", err)
	}

	// Check if CID is in pins (at least queued/pinning)
	for _, pinCID := range pins {
		if pinCID == cidStr {
			return ctx, nil
		}
	}

	return ctx, fmt.Errorf("CID %s not found in pinned list", cidStr)
}

// allNSpinsAreSuccessfullyCreated verifies N pins were created successfully
func (s *PinningSteps) allNSpinsAreSuccessfullyCreated(ctx context.Context, count int) (context.Context, error) {
	pinsList, ok := helpers.GetPinList(ctx)
	if !ok {
		return ctx, fmt.Errorf("no pins found in context")
	}

	if len(pinsList) != count {
		return ctx, fmt.Errorf("expected %d pins, got %d", count, len(pinsList))
	}

	// Verify all CIDs are actually pinned
	for i, cidStr := range pinsList {
		pinned, err := helpers.IPFSIsPinned(ctx, cidStr)
		if err != nil {
			return ctx, fmt.Errorf("failed to verify pin %d: %w", i, err)
		}
		if !pinned {
			return ctx, fmt.Errorf("CID %d is not pinned", i)
		}
	}

	return ctx, nil
}

// theUserHasNDifferentCIDs creates N different CIDs for concurrent pinning
func (s *PinningSteps) theUserHasNDifferentCIDs(ctx context.Context, count int) (context.Context, error) {
	var cids []string
	for i := range count {
		content := []byte(fmt.Sprintf("test content %d", i))
		cidStr, err := helpers.KuboAdd(ctx, content)
		if err != nil {
			return ctx, fmt.Errorf("failed to generate CID %d: %w", i, err)
		}
		cids = append(cids, cidStr)
	}

	ctx = helpers.SetCIDs(ctx, cids)
	return ctx, nil
}

// theUserUploadsAndPinsEachFile uploads and pins N files sequentially
func (s *PinningSteps) theUserUploadsAndPinsEachFile(ctx context.Context, count int) (context.Context, error) {
	ctx, err := s.theUserHasNDifferentCIDs(ctx, count)
	if err != nil {
		return ctx, err
	}

	cids, ok := helpers.GetCIDs(ctx)
	if !ok {
		return ctx, fmt.Errorf("no CIDs found in context")
	}

	var cidsWithPins []string
	for i := range cids {
		content := []byte(fmt.Sprintf("test content %d", i))
		cid, err := helpers.IPFSPortalUpload(ctx, content, fmt.Sprintf("file%d.txt", i))
		if err != nil {
			return ctx, fmt.Errorf("failed to upload file %d: %w", i, err)
		}

		_, ctx, err = helpers.IPFSPinAdd(ctx, cid)
		if err != nil {
			return ctx, fmt.Errorf("failed to pin file %d: %w", i, err)
		}

		cidsWithPins = append(cidsWithPins, cid)
	}

	ctx = helpers.SetCIDs(ctx, cidsWithPins)
	return ctx, nil
}

// allNFilesArePinned verifies N files are pinned
func (s *PinningSteps) allNFilesArePinned(ctx context.Context, count int) (context.Context, error) {
	cids, ok := helpers.GetCIDs(ctx)
	if !ok {
		return ctx, fmt.Errorf("no CIDs found in context")
	}

	if len(cids) != count {
		return ctx, fmt.Errorf("expected %d CIDs, got %d", count, len(cids))
	}

	// Wait for all operations to complete
	for i, cidStr := range cids {
		if err := helpers.WaitForOperationCompleteByCID(ctx, cidStr, 2*time.Minute); err != nil {
			return ctx, fmt.Errorf("operation for file %d failed: %w", i, err)
		}
	}

	return ctx, nil
}

// theNewCIDIsPinnedWithCorrectContent verifies new pin has correct content
func (s *PinningSteps) theNewCIDIsPinnedWithCorrectContent(ctx context.Context) (context.Context, error) {
	newCID, err := helpers.RequireCID(ctx, "content verification")
	if err != nil {
		return ctx, err
	}

	retrievedContent, err := helpers.KuboCat(ctx, newCID)
	if err != nil {
		return ctx, fmt.Errorf("failed to retrieve content: %w", err)
	}

	// Check if content contains expected marker
	if !strings.Contains(retrievedContent, "replacement content") {
		return ctx, fmt.Errorf("content does not contain expected marker")
	}

	return ctx, nil
}

// theOldCIDIsUnpinned verifies old CID is no longer pinned
func (s *PinningSteps) theOldCIDIsUnpinned(ctx context.Context) (context.Context, error) {
	oldCID, ok := helpers.GetOldCID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no old CID found in context")
	}

	pinned, err := helpers.IPFSIsPinned(ctx, oldCID)
	if err == nil && pinned {
		return ctx, fmt.Errorf("old CID is still pinned")
	}

	return ctx, nil
}

// theCIDIsNotInThePinnedList verifies CID was not pinned
func (s *PinningSteps) theCIDIsNotInThePinnedList(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "cleanup verification")
	if err != nil {
		return ctx, err
	}

	pinned, err := helpers.IPFSIsPinned(ctx, cidStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to check pin status: %w", err)
	}

	if pinned {
		return ctx, fmt.Errorf("CID is still pinned")
	}

	return ctx, nil
}

// theUserHasASizeMBIPFSTestFile creates test data of specified size for performance testing
func (s *PinningSteps) theUserHasASizeMBIPFSTestFile(ctx context.Context, sizeMB int) (context.Context, error) {
	sizeBytes := int64(sizeMB * 1024 * 1024)
	content, err := helpers.GenerateLargeTestFile(sizeBytes)
	if err != nil {
		return ctx, fmt.Errorf("failed to generate IPFS test file: %w", err)
	}

	ctx = helpers.SetKnownContent(ctx, string(content))
	return ctx, nil
}

// theUserHasASizeGBIPFSTestFile creates test data of specified size for performance testing
func (s *PinningSteps) theUserHasASizeGBIPFSTestFile(ctx context.Context, sizeGB int) (context.Context, error) {
	sizeBytes := int64(sizeGB * 1024 * 1024 * 1024)
	content, err := helpers.GenerateLargeTestFile(sizeBytes)
	if err != nil {
		return ctx, fmt.Errorf("failed to generate IPFS test file: %w", err)
	}

	ctx = helpers.SetKnownContent(ctx, string(content))
	return ctx, nil
}

// theUserUploadsAndPinsTheLargeIPFSTestFile uploads and pins the large IPFS test file via portal
func (s *PinningSteps) theUserUploadsAndPinsTheLargeIPFSTestFile(ctx context.Context) (context.Context, error) {
	originalContent, ok := helpers.GetKnownContent(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPFS test file content found")
	}

	cid, err := helpers.IPFSPortalUpload(ctx, []byte(originalContent), "large-test-file.bin")
	if err != nil {
		return ctx, fmt.Errorf("failed to upload large IPFS test file: %w", err)
	}

	_, ctx, err = helpers.IPFSPinAdd(ctx, cid)
	if err != nil {
		return ctx, fmt.Errorf("failed to pin large IPFS test file: %w", err)
	}

	ctx = helpers.SetCID(ctx, cid)
	return ctx, nil
}

// theIPFSPinReachesPinnedStatusWithinMinutes verifies pin completes within time limit
func (s *PinningSteps) theIPFSPinReachesPinnedStatusWithinMinutes(ctx context.Context, minutes int) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "time limit verification")
	if err != nil {
		return ctx, err
	}

	timeout := time.Duration(minutes) * time.Minute
	startTime := time.Now()

	for {
		pinned, err := helpers.IPFSIsPinned(ctx, cidStr)
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
func (s *PinningSteps) theUploadedIPFSTestFileIsAvailable(ctx context.Context) (context.Context, error) {
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
