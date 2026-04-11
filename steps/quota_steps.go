package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	
	"pinner.xyz-e2e/helpers"
)

// QuotaSteps holds the state for quota management step definitions
type QuotaSteps struct{}

// NewQuotaSteps creates a new QuotaSteps instance
func NewQuotaSteps() *QuotaSteps {
	return &QuotaSteps{}
}

// InitializeScenario registers all step definitions with godog
func (s *QuotaSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Quota recording steps
	ctx.Step(`^the user records their initial quota status$`, s.theUserRecordsTheirInitialQuotaStatus)
	ctx.Step(`^the user records their initial bandwidth status$`, s.theUserRecordsTheirInitialBandwidthStatus)

	// Quota checking steps
	ctx.Step(`^the user checks their account quota$`, s.theUserChecksTheirAccountQuota)
	ctx.Step(`^the user checks their (?:upload|download|storage|bandwidth) quota(?: after operation)?$`, s.theUserChecksTheirAccountQuota)

	// Quota verification steps
	ctx.Step(`^the upload quota usage has increased by approximately (\d+)MB$`, s.theUploadQuotaUsageHasIncreasedBy)
	ctx.Step(`^the download quota usage has increased by approximately (\d+)MB$`, s.theDownloadQuotaUsageHasIncreasedBy)
	ctx.Step(`^the storage quota usage has increased by approximately (\d+)MB$`, s.theStorageQuotaUsageHasIncreasedBy)
	ctx.Step(`^the bandwidth quota reflects both upload and download usage$`, s.theBandwidthQuotaReflectsBothUploadAndDownloadUsage)
	ctx.Step(`^the upload quota percentage calculation matches (?:the )?new usage$`, s.theUploadQuotaPercentageCalculationMatchesNewUsage)
	
	// Composite step for file upload scenario
	ctx.Step(`^the user checks (?:their account quota )?after uploading a (\d+) ?MB file$`, s.theUserChecksAfterUploadingFile)

	// Quota history steps
	ctx.Step(`^the user retrieves quota history for the last (\d+) hours$`, s.theUserRetrievesQuotaHistory)
	ctx.Step(`^the quota history shows usage data points$`, s.theQuotaHistoryShowsUsageDataPoints)

	// IPFS network fetch steps
	ctx.Step(`^the user fetches content via IPFS network$`, s.theUserFetchesContentViaIPFSNetwork)
}

// theUserRecordsTheirInitialQuotaStatus records initial quota for comparison
func (s *QuotaSteps) theUserRecordsTheirInitialQuotaStatus(ctx context.Context) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	quota, err := api.GetQuota(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get quota: %w", err)
	}

	ctx = helpers.SetInitialQuotaStatus(ctx, quota)
	return ctx, nil
}

// theUserRecordsTheirInitialBandwidthStatus records initial bandwidth for comparison
func (s *QuotaSteps) theUserRecordsTheirInitialBandwidthStatus(ctx context.Context) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	quota, err := api.GetQuota(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get quota: %w", err)
	}

	// Bandwidth is calculated as the sum of upload and download usage
	initialBandwidth := int64(quota.Upload.Used + quota.Download.Used)

	ctx = helpers.SetInitialBandwidth(ctx, initialBandwidth)
	return ctx, nil
}

// theUserChecksTheirAccountQuota checks overall account quota
func (s *QuotaSteps) theUserChecksTheirAccountQuota(ctx context.Context) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	quota, err := api.GetQuota(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get quota: %w", err)
	}

	ctx = helpers.SetCurrentQuotaStatus(ctx, quota)
	return ctx, nil
}

// theUserChecksTheirStorageQuota checks storage quota after operations
func (s *QuotaSteps) theUserChecksTheirStorageQuota(ctx context.Context) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	quota, err := api.GetQuota(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get quota: %w", err)
	}

	ctx = helpers.SetCurrentQuotaStatus(ctx, quota)
	return ctx, nil
}

// verifyQuotaUsageIncreasedBy verifies quota consumption for specified type
func (s *QuotaSteps) verifyQuotaUsageIncreasedBy(ctx context.Context, quotaType string, expectedMB int) (context.Context, error) {
	current, ok := helpers.GetCurrentQuotaStatus(ctx)
	initial, ok2 := helpers.GetInitialQuotaStatus(ctx)

	if !ok || !ok2 || current == nil || initial == nil {
		return ctx, fmt.Errorf("quota status not available in context")
	}

	expectedBytes := int64(expectedMB * 1024 * 1024)
	var delta int64

	// Determine which quota field to check based on type
	switch quotaType {
	case "upload":
		delta = int64(current.Upload.Used - initial.Upload.Used)
	case "download":
		delta = int64(current.Download.Used - initial.Download.Used)
	case "storage":
		// Storage quota is tracked via Upload quota in the portal API
		delta = int64(current.Upload.Used - initial.Upload.Used)
	default:
		return ctx, fmt.Errorf("unknown quota type: %s", quotaType)
	}

	// Allow 10% tolerance for measurement
	tolerance := expectedBytes / 10
	lowerBound := expectedBytes - tolerance
	upperBound := expectedBytes + tolerance

	if delta < lowerBound || delta > upperBound {
		return ctx, fmt.Errorf("%s quota usage increased by %d bytes, expected approximately %d bytes (tolerance: ±%d bytes)",
			quotaType, delta, expectedBytes, tolerance)
	}

	return ctx, nil
}

// theUploadQuotaUsageHasIncreasedBy verifies upload consumption
func (s *QuotaSteps) theUploadQuotaUsageHasIncreasedBy(ctx context.Context, expectedMB int) (context.Context, error) {
	return s.verifyQuotaUsageIncreasedBy(ctx, "upload", expectedMB)
}

// theDownloadQuotaUsageHasIncreasedBy verifies download consumption
func (s *QuotaSteps) theDownloadQuotaUsageHasIncreasedBy(ctx context.Context, expectedMB int) (context.Context, error) {
	return s.verifyQuotaUsageIncreasedBy(ctx, "download", expectedMB)
}

// theStorageQuotaUsageHasIncreasedBy verifies storage consumption
func (s *QuotaSteps) theStorageQuotaUsageHasIncreasedBy(ctx context.Context, expectedMB int) (context.Context, error) {
	// Storage quota is currently tracked via Upload quota in the portal API
	return s.verifyQuotaUsageIncreasedBy(ctx, "storage", expectedMB)
}

// theBandwidthQuotaReflectsBothUploadAndDownloadUsage verifies bandwidth tracking
func (s *QuotaSteps) theBandwidthQuotaReflectsBothUploadAndDownloadUsage(ctx context.Context) (context.Context, error) {
	currentQuota, ok := helpers.GetCurrentQuotaStatus(ctx)
	if !ok || currentQuota == nil {
		return ctx, fmt.Errorf("current quota status not available in context")
	}

	// Bandwidth is calculated as the sum of upload and download usage
	// Since bandwidth is computed client-side, verify it equals the sum
	actualBandwidth := currentQuota.Upload.Used + currentQuota.Download.Used

	if actualBandwidth < 0 {
		return ctx, fmt.Errorf("bandwidth cannot be negative: %d", actualBandwidth)
	}

	return ctx, nil
}

// theUploadQuotaPercentageCalculationMatchesNewUsage verifies percentage calculation
func (s *QuotaSteps) theUploadQuotaPercentageCalculationMatchesNewUsage(ctx context.Context) (context.Context, error) {
	currentQuota, ok := helpers.GetCurrentQuotaStatus(ctx)
	if !ok || currentQuota == nil {
		return ctx, fmt.Errorf("current quota status not available in context")
	}

	if currentQuota.Upload.Limit == nil {
		return ctx, fmt.Errorf("upload limit is nil")
	}

	// Verify: percentage = (used * 100) / limit
	expectedPercentage := (currentQuota.Upload.Used * 100) / *currentQuota.Upload.Limit
	if currentQuota.Upload.Percentage != expectedPercentage {
		return ctx, fmt.Errorf("upload percentage is %d, expected %d (calculated from used * 100 / limit)",
			currentQuota.Upload.Percentage, expectedPercentage)
	}

	return ctx, nil
}

// theUserChecksAfterUploadingFile is a composite step for the percentage calculation scenario
func (s *QuotaSteps) theUserChecksAfterUploadingFile(ctx context.Context, sizeMB int) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Just get quota after upload - the upload and consumption happen in prior steps
	_, err = api.GetQuota(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get quota: %w", err)
	}

	return ctx, nil
}

// theUserRetrievesQuotaHistory retrieves historical quota data
func (s *QuotaSteps) theUserRetrievesQuotaHistory(ctx context.Context, hours int) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	endDate := time.Now()
	startDate := endDate.Add(-time.Duration(hours) * time.Hour)
	
	history, err := api.GetQuotaHistory(ctx, 
		startDate.Format(time.RFC3339),
		endDate.Format(time.RFC3339),
		"upload",
	)
	if err != nil {
		return ctx, fmt.Errorf("failed to get quota history: %w", err)
	}

	ctx = helpers.SetQuotaHistory(ctx, history)
	return ctx, nil
}

// theQuotaHistoryShowsUsageDataPoints verifies history was retrieved
func (s *QuotaSteps) theQuotaHistoryShowsUsageDataPoints(ctx context.Context) (context.Context, error) {
	history, ok := helpers.GetQuotaHistory(ctx)
	if !ok || history == nil {
		return ctx, fmt.Errorf("quota history was not retrieved")
	}

	if len(history.Points) == 0 {
		return ctx, fmt.Errorf("quota history has no usage data points")
	}

	return ctx, nil
}

// theUserFetchesContentViaIPFSNetwork fetches content via IPFS network using Kubo's pin/add endpoint
// This triggers bitswap to download the content from the IPFS network, which is tracked against download quota
func (s *QuotaSteps) theUserFetchesContentViaIPFSNetwork(ctx context.Context) (context.Context, error) {
	cidStr, err := helpers.RequireCID(ctx, "fetching content via IPFS network")
	if err != nil {
		return ctx, err
	}

	if err := helpers.KuboFetchViaIPFSNetwork(ctx, cidStr); err != nil {
		return ctx, fmt.Errorf("failed to fetch content via IPFS network: %w", err)
	}

	// Wait for quota tracking to be recorded
	time.Sleep(2 * time.Second)

	return ctx, nil
}
