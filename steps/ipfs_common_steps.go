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
