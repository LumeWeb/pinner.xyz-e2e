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
