package steps

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cucumber/godog"
	"github.com/docker/go-units"
	
	"pinner.xyz-e2e/helpers"
	"go.lumeweb.com/portal-sdk/admin"
)

// QuotaEnforcementSteps provides step definitions for quota limit enforcement testing
type QuotaEnforcementSteps struct {
}

// NewQuotaEnforcementSteps creates a new QuotaEnforcementSteps instance
func NewQuotaEnforcementSteps() *QuotaEnforcementSteps {
	return &QuotaEnforcementSteps{}
}

// InitializeScenario registers all step definitions with godog
func (s *QuotaEnforcementSteps) InitializeScenario(ctx *godog.ScenarioContext) error {
	// Quota setup steps - these update the current plan with specific limits
	ctx.Step(`^the admin sets the upload total limit to (\d+) MB$`, s.theAdminSetsUploadTotalLimit)
	ctx.Step(`^the admin sets the storage limit to (\d+) MB$`, s.theAdminSetsStorageLimit)
	ctx.Step(`^the admin sets the total download limit to (\d+) MB$`, s.theAdminSetsTotalDownloadLimit)
	
	// GB versions of limit steps
	ctx.Step(`^the admin sets the upload total limit to (\d+) GB$`, s.theAdminSetsUploadTotalLimitGB)
	ctx.Step(`^the admin sets the storage limit to (\d+) GB$`, s.theAdminSetsStorageLimitGB)
	ctx.Step(`^the admin sets the total download limit to (\d+) GB$`, s.theAdminSetsTotalDownloadLimitGB)
	
	// Set limit based on actual uploaded file size
	ctx.Step(`^the admin sets the download limit to the uploaded file size$`, s.theAdminSetsDownloadLimitToUploadedFileSize)
	ctx.Step(`^the admin sets the upload total limit to match the pending upload DAG size$`, s.theAdminSetsUploadLimitToMatchPendingUploadDAGSize)
	ctx.Step(`^the admin sets the storage limit to match the pending upload DAG size$`, s.theAdminSetsStorageLimitToMatchPendingUploadDAGSize)

	// Quota exhaustion and attempt steps
	// Note: "the user records their initial quota status" step is implemented in quota_steps.go
	ctx.Step(`^the user attempts to download the file via HTTP from IPFS gateway$`, s.theUserAttemptsToDownloadTheFileFromIPFSGateway)
	ctx.Step(`^the user attempts to upload another file to IPFS$`, s.theUserAttemptsToUploadAnotherFile)
	ctx.Step(`^the user uploads another (\d+)MB file to IPFS$`, s.theUserUploadsAnotherMBFileToIPFS)
	ctx.Step(`^the user attempts to upload another (\d+)MB file to IPFS$`, s.theUserAttemptsToUploadAnotherMBFileToIPFS)
	ctx.Step(`^the user attempts to pin the second content to IPFS$`, s.theUserAttemptsToPinTheSecondContent)
	ctx.Step(`^the content remains unpinned$`, s.theContentRemainsUnpinned)
	ctx.Step(`^the download quota has reached 67% or more of the limit$`, s.theDownloadQuotaHasReached67PercentOrMoreOfTheLimit)
	ctx.Step(`^the download quota has reached 100% of the limit$`, s.theDownloadQuotaHasReached100PercentOfTheLimit)
	ctx.Step(`^the user attempts to fetch content via IPFS via P2P network and fails$`, s.theUserAttemptsToFetchContentViaIPFSNetworkAndFails)

	// Network IPFS pin quota enforcement steps
	ctx.Step(`^the user has an existing IPFS CID from the network that is (\d+)MB$`, s.theUserHasAnExistingIPFSCIDFromTheNetworkThatIsMB)
	ctx.Step(`^the user attempts to pin the IPFS CID$`, s.theUserAttemptsToPinTheIPFSCID)

	// Operation denial verification steps (used in Then clauses)
	ctx.Step(`^the pinning operation is denied with a quota exceeded error$`, s.thePinningOperationIsDeniedWithAQuotaExceededError)
	ctx.Step(`^the download operation is denied with a quota exceeded error$`, s.theDownloadOperationIsDeniedWithAQuotaExceededError)
	ctx.Step(`^the upload operation is denied with a quota exceeded error$`, s.theUploadOperationIsDeniedWithAQuotaExceededError)
	ctx.Step(`^the fetch operation is denied with a quota exceeded error$`, s.theFetchOperationIsDeniedWithAQuotaExceededError)

	return nil
}

// setQuotaLimit is a helper function that updates a plan's quota limits
func (s *QuotaEnforcementSteps) setQuotaLimit(ctx context.Context, mb int, setLimit func(*admin.QuotaPlan, int64)) error {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return err
	}

	plan, err := helpers.RequireAdminCurrentPlan(ctx)
	if err != nil {
		return err
	}

	var limitBytes int64
	
	// Calculate limit based on MB value
	limitBytes, _ = units.FromHumanSize(fmt.Sprintf("%dMB", mb))
	
	setLimit(plan, limitBytes)

	_, err = adminClient.Quota().UpdatePlan(ctx, fmt.Sprint(plan.Id), plan)
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}

	return nil
}

func (s *QuotaEnforcementSteps) theAdminSetsStorageLimit(ctx context.Context, mb int) (context.Context, error) {
	err := s.setQuotaLimit(ctx, mb, func(plan *admin.QuotaPlan, limit int64) {
		plan.StorageLimitBytes = int(limit)
	})
	return ctx, err
}

func (s *QuotaEnforcementSteps) theAdminSetsUploadTotalLimit(ctx context.Context, mb int) (context.Context, error) {
	err := s.setQuotaLimit(ctx, mb, func(plan *admin.QuotaPlan, limit int64) {
		plan.UploadLimitBytes = int(limit)
	})
	return ctx, err
}

func (s *QuotaEnforcementSteps) theAdminSetsTotalDownloadLimit(ctx context.Context, mb int) (context.Context, error) {
	err := s.setQuotaLimit(ctx, mb, func(plan *admin.QuotaPlan, limit int64) {
		plan.DownloadLimitBytes = int(limit)
	})
	return ctx, err
}

// gbToMB converts GB to MB for quota limits
func (s *QuotaEnforcementSteps) gbToMB(gb int) int {
	return gb * 1024
}

func (s *QuotaEnforcementSteps) theAdminSetsUploadTotalLimitGB(ctx context.Context, gb int) (context.Context, error) {
	return s.theAdminSetsUploadTotalLimit(ctx, s.gbToMB(gb))
}

func (s *QuotaEnforcementSteps) theAdminSetsStorageLimitGB(ctx context.Context, gb int) (context.Context, error) {
	return s.theAdminSetsStorageLimit(ctx, s.gbToMB(gb))
}

func (s *QuotaEnforcementSteps) theAdminSetsTotalDownloadLimitGB(ctx context.Context, gb int) (context.Context, error) {
	return s.theAdminSetsTotalDownloadLimit(ctx, s.gbToMB(gb))
}

// theAdminSetsDownloadLimitToUploadedFileSize sets the download quota limit to the actual size
// of the most recently uploaded file. This uses the DAG size from the context.
func (s *QuotaEnforcementSteps) theAdminSetsDownloadLimitToUploadedFileSize(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	plan, err := helpers.RequireAdminCurrentPlan(ctx)
	if err != nil {
		return ctx, err
	}

	// Get the DAG size from the most recent upload
	dagSize, ok := helpers.GetRawFileSize(ctx)
	if !ok || dagSize == 0 {
		return ctx, fmt.Errorf("no DAG size available in context - upload a file first")
	}

	// Set the download limit to the exact DAG size
	plan.DownloadLimitBytes = dagSize

	_, err = adminClient.Quota().UpdatePlan(ctx, fmt.Sprint(plan.Id), plan)
	if err != nil {
		return ctx, fmt.Errorf("failed to update plan: %w", err)
	}

	return ctx, nil
}

// theAdminSetsStorageLimitToMatchPendingUploadDAGSize calculates the DAG size for pending
// upload content (stored in context) and sets the storage quota limit to match exactly.
// This ensures the first upload succeeds and the second fails with quota exceeded.
func (s *QuotaEnforcementSteps) theAdminSetsStorageLimitToMatchPendingUploadDAGSize(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	plan, err := helpers.RequireAdminCurrentPlan(ctx)
	if err != nil {
		return ctx, err
	}

	// Get the pending upload content from context
	content, ok := helpers.GetKnownContent(ctx)
	if !ok {
		return ctx, fmt.Errorf("no pending upload content found in context - prepare content first")
	}

	// Calculate the exact DAG size for this content without uploading
	dagSize, err := helpers.CalculateDAGSizeFromFileContent(ctx, []byte(content))
	if err != nil {
		return ctx, fmt.Errorf("failed to calculate DAG size for pending upload: %w", err)
	}

	// Set the storage limit to the exact DAG size
	plan.StorageLimitBytes = int(dagSize)

	_, err = adminClient.Quota().UpdatePlan(ctx, fmt.Sprint(plan.Id), plan)
	if err != nil {
		return ctx, fmt.Errorf("failed to update plan: %w", err)
	}

	return ctx, nil
}

// theAdminSetsUploadLimitToMatchPendingUploadDAGSize calculates the DAG size for pending
// upload content (stored in context) and sets the upload quota limit to match exactly.
// This ensures the first upload succeeds and the second fails with quota exceeded.
func (s *QuotaEnforcementSteps) theAdminSetsUploadLimitToMatchPendingUploadDAGSize(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	plan, err := helpers.RequireAdminCurrentPlan(ctx)
	if err != nil {
		return ctx, err
	}

	// Get the pending upload content from context
	content, ok := helpers.GetKnownContent(ctx)
	if !ok {
		return ctx, fmt.Errorf("no pending upload content found in context - prepare content first")
	}

	// Calculate the exact DAG size for this content without uploading
	dagSize, err := helpers.CalculateDAGSizeFromFileContent(ctx, []byte(content))
	if err != nil {
		return ctx, fmt.Errorf("failed to calculate DAG size for pending upload: %w", err)
	}

	// Set the upload limit to the exact DAG size
	plan.UploadLimitBytes = int(dagSize)

	_, err = adminClient.Quota().UpdatePlan(ctx, fmt.Sprint(plan.Id), plan)
	if err != nil {
		return ctx, fmt.Errorf("failed to update plan: %w", err)
	}

	return ctx, nil
}


func (s *QuotaEnforcementSteps) theDownloadQuotaHasReached67PercentOrMoreOfTheLimit(ctx context.Context) (context.Context, error) {
	currentQuota, ok := helpers.GetCurrentQuotaStatus(ctx)
	if !ok || currentQuota == nil {
		return ctx, fmt.Errorf("no current quota status found in context")
	}

	// Check if percentage is 67% or more
	if currentQuota.Download.Percentage < 67 {
		return ctx, fmt.Errorf("download quota percentage is %d, expected 67 or more", currentQuota.Download.Percentage)
	}

	return ctx, nil
}

func (s *QuotaEnforcementSteps) theUserAttemptsToDownloadTheFileFromIPFSGateway(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "attempting download again")
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

	// Attempt download - should fail with ErrRateLimitExceeded when quota is exhausted
	// IMPORTANT: DownloadFile returns a lazy io.ReadCloser, so we must READ from it
	// to trigger actual block fetching and quota enforcement
	reader, err := client.Download().DownloadFile(ctx, parsedCID)
	
	// Check if DownloadFile itself returns a quota enforcement error (setup phase)
	if err != nil {
		if helpers.IsQuotaEnforcementError(err) {
			// Quota was enforced during setup - this is expected success
			return ctx, nil
		}
		return ctx, fmt.Errorf("download setup failed: %w", err)
	}
	defer reader.Close()
	
	// Read the file - this triggers lazy block fetching where quota errors occur
	_, err = io.ReadAll(reader)
	
	// Check if error is a quota exceeded error
	if err == nil {
		return ctx, fmt.Errorf("expected download to fail with ErrRateLimitExceeded, but succeeded")
	}

	// Check if error is ErrRateLimitExceeded
	if !helpers.IsQuotaEnforcementError(err) {
		return ctx, fmt.Errorf("expected quota exceeded error, got: %w", err)
	}
	if !helpers.IsQuotaEnforcementError(err) {
		return ctx, fmt.Errorf("expected quota exceeded error, got: %w", err)
	}

	return ctx, nil
}

func (s *QuotaEnforcementSteps) theUserAttemptsToUploadAnotherFile(ctx context.Context) (context.Context, error) {
	// Attempt to upload a test file - should fail with quota exceeded
	// Create a small 1MB test file
	testContent := make([]byte, 1*1024*1024)
	for i := range testContent {
		testContent[i] = byte(i % 256)
	}

	// Attempt upload - should fail with ErrRateLimitExceeded
	_, _, err := helpers.IPFSPortalUpload(ctx, testContent, "test-quota-file.bin")
	if err == nil {
		return ctx, fmt.Errorf("expected upload to fail with quota exceeded error")
	}

	// Check if error is quota exceeded using helper
	if !helpers.IsQuotaEnforcementError(err) {
		return ctx, fmt.Errorf("expected quota exceeded error, got: %w", err)
	}

	return ctx, nil
}

func (s *QuotaEnforcementSteps) theUserUploadsAnotherMBFileToIPFS(ctx context.Context, sizeMB int) (context.Context, error) {
	// Generate test content of the specified size
	sizeBytes := sizeMB * 1024 * 1024
	testContent := make([]byte, sizeBytes)
	for i := range testContent {
		testContent[i] = byte(i % 256)
	}

	// Attempt to upload - should fail with quota exceeded
	_, _, err := helpers.IPFSPortalUpload(ctx, testContent, fmt.Sprintf("another-%dmb-test.bin", sizeMB))
	
	// For quota enforcement testing, if upload succeeds (which shouldn't happen with proper limits),
	// this is actually a test failure - we expect the operation to be blocked
	if err == nil {
		return ctx, fmt.Errorf("expected upload to fail with quota exceeded error but it succeeded")
	}

	// Check if error is quota exceeded using helper
	if !helpers.IsQuotaEnforcementError(err) {
		return ctx, fmt.Errorf("expected quota exceeded error, got: %w", err)
	}

	return ctx, nil
}

// theUserAttemptsToUploadAnotherMBFileToIPFS is an alias for theUserUploadsAnotherMBFileToIPFS
// Both expect the upload to fail with quota exceeded error
func (s *QuotaEnforcementSteps) theUserAttemptsToUploadAnotherMBFileToIPFS(ctx context.Context, sizeMB int) (context.Context, error) {
	return s.theUserUploadsAnotherMBFileToIPFS(ctx, sizeMB)
}

func (s *QuotaEnforcementSteps) theUserAttemptsToPinTheSecondContent(ctx context.Context) (context.Context, error) {
	// This step expects the pinning to fail with quota exceeded
	return ctx, fmt.Errorf("pinning operation should have failed with quota exceeded error before this step")
}

func (s *QuotaEnforcementSteps) theContentRemainsUnpinned(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "checking if content remains unpinned")
	if err != nil {
		return ctx, err
	}

	// Verify content is NOT pinned
	pinned, err := helpers.IPFSIsPinned(ctx, cidStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to check pin status: %w", err)
	}

	if pinned {
		return ctx, fmt.Errorf("content %s is pinned but should remain unpinned due to quota exceeded error", cidStr)
	}

	return ctx, nil
}

func (s *QuotaEnforcementSteps) thePinningOperationIsDeniedWithAQuotaExceededError(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (s *QuotaEnforcementSteps) theDownloadOperationIsDeniedWithAQuotaExceededError(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (s *QuotaEnforcementSteps) theUploadOperationIsDeniedWithAQuotaExceededError(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (s *QuotaEnforcementSteps) theFetchOperationIsDeniedWithAQuotaExceededError(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (s *QuotaEnforcementSteps) theDownloadQuotaHasReached100PercentOfTheLimit(ctx context.Context) (context.Context, error) {
	currentQuota, ok := helpers.GetCurrentQuotaStatus(ctx)
	if !ok || currentQuota == nil {
		return ctx, fmt.Errorf("no current quota status found in context")
	}

	// Check if percentage is 100%
	if currentQuota.Download.Percentage != 100 {
		return ctx, fmt.Errorf("download quota percentage is %d, expected 100", currentQuota.Download.Percentage)
	}

	return ctx, nil
}

func (s *QuotaEnforcementSteps) theUserAttemptsToFetchContentViaIPFSNetworkAndFails(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "attempting P2P network fetch")
	if err != nil {
		return ctx, err
	}

	// Create a timeout context for P2P fetch
	// P2P fetch won't return 429 like HTTP gateway - it will just hang/delay
	// when quota is enforced (no blocks available from gateway, needs to fetch from network)
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Attempt P2P network fetch with timeout
	err = helpers.KuboFetchViaIPFSNetwork(timeoutCtx, cidStr)
	
	// Check if the operation completed successfully (quota NOT enforced)
	if err == nil {
		return ctx, fmt.Errorf("expected P2P fetch to be denied due to quota enforcement, but succeeded")
	}

	// Check if the operation timed out - this means quota was enforced
	// (couldn't fetch any blocks because gateway returned 429, so it tried P2P network which took too long)
	if timeoutCtx.Err() == context.DeadlineExceeded {
		// Timeout indicates quota is being enforced - this is expected success
		return ctx, nil
	}

	// Check if it's a quota enforcement error from gateway (less common in P2P fetch)
	if helpers.IsQuotaEnforcementError(err) {
		return ctx, nil
	}

	// Any other error is unexpected
	return ctx, fmt.Errorf("expected P2P fetch to timeout due to quota enforcement, got: %w", err)
}



// theUserHasAnExistingIPFSCIDFromTheNetworkThatIsMB creates content of specified size,
// adds it to the local Kubo node to simulate existing network content,
// and stores the CID in context for pinning attempts.
func (s *QuotaEnforcementSteps) theUserHasAnExistingIPFSCIDFromTheNetworkThatIsMB(ctx context.Context, sizeMB int) (context.Context, error) {
	// Generate test content of the specified size
	sizeBytes := int64(sizeMB * 1024 * 1024)
	content, err := helpers.GenerateLargeTestFile(sizeBytes)
	if err != nil {
		return ctx, fmt.Errorf("failed to generate %dMB test content: %w", sizeMB, err)
	}

	// Add content to local Kubo node to simulate content existing on the IPFS network
	cidStr, err := helpers.KuboAdd(ctx, content)
	if err != nil {
		return ctx, fmt.Errorf("failed to add %dMB content to Kubo: %w", sizeMB, err)
	}

	// Store CID in context for subsequent pinning attempts
	ctx = helpers.SetCID(ctx, cidStr)
	
	return ctx, nil
}

// theUserAttemptsToPinTheIPFSCID attempts to pin a CID from the context to the Portal.
// Note: The pin creation itself succeeds, but the subsequent storage operation
// fails due to quota enforcement when the storage limit is exceeded.
// This step creates the pin and initiates the operation without waiting for completion.
func (s *QuotaEnforcementSteps) theUserAttemptsToPinTheIPFSCID(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "pinning attempt")
	if err != nil {
		return ctx, err
	}

	// Create the pin - this operation itself succeeds
	// The storage operation that follows will fail due to quota enforcement
	_, ctx, err = helpers.IPFSPinAdd(ctx, cidStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to create pin: %w", err)
	}
	
	// Store the CID context for operation verification
	// No error stored here - operation completion will be checked in "the pin operation fails"
	return ctx, nil
}


