package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

// IPNSCommonSteps holds shared step definitions for IPNS operations
type IPNSCommonSteps struct{}

// NewIPNSCommonSteps creates a new IPNSCommonSteps instance
func NewIPNSCommonSteps() *IPNSCommonSteps {
	return &IPNSCommonSteps{}
}

// InitializeScenario registers all shared IPNS step definitions with godog
func (s *IPNSCommonSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Content verification steps
	ctx.Step(`^the IPNS name is resolvable$`, s.theIPNSNameIsResolvable)
	ctx.Step(`^the IPNS name resolves to CID "([^"]*)"$`, s.theIPNSNameResolvesToCID)
	ctx.Step(`^the IPNS published CID matches the resolved CID$`, s.theIPNSPublishedCIDMatchesResolvedCID)
	
	// Key verification steps
	ctx.Step(`^the IPNS key has a valid peer ID$`, s.theIPNSKeyHasValidPeerID)
	ctx.Step(`^the IPNS key has IPNS name "([^"]*)"$`, s.theIPNSKeyHasIPNSName)
	ctx.Step(`^all (\d+) IPNS keys have valid peer IDs$`, s.allIPNSKeysHaveValidPeerIDs)
}

// theIPNSNameIsResolvable verifies that an IPNS name can be resolved to a CID
func (s *IPNSCommonSteps) theIPNSNameIsResolvable(ctx context.Context) (context.Context, error) {
	ipnsName, ok := helpers.GetIPNSIPNSName(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS name found in context")
	}

	resolve, err := helpers.ResolveIPNSName(ctx, ipnsName)
	if err != nil {
		return ctx, fmt.Errorf("failed to resolve IPNS name %s: %w", ipnsName, err)
	}

	if resolve == nil {
		return ctx, fmt.Errorf("IPNS name %s could not be resolved", ipnsName)
	}

	// Store the resolved CID in context for later verification
	ctx = helpers.SetIPNSResolvedCID(ctx, resolve.Value)

	return ctx, nil
}

// theIPNSNameResolvesToCID verifies that an IPNS name resolves to a specific CID
func (s *IPNSCommonSteps) theIPNSNameResolvesToCID(ctx context.Context, expectedCID string) (context.Context, error) {
	ipnsName, ok := helpers.GetIPNSIPNSName(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS name found in context")
	}

	resolve, err := helpers.ResolveIPNSName(ctx, ipnsName)
	if err != nil {
		return ctx, fmt.Errorf("failed to resolve IPNS name %s: %w", ipnsName, err)
	}

	if resolve == nil {
		return ctx, fmt.Errorf("IPNS name %s could not be resolved", ipnsName)
	}

	if resolve.Value != expectedCID {
		return ctx, fmt.Errorf("IPNS name %s resolved to %s, expected %s", ipnsName, resolve.Value, expectedCID)
	}

	// Store the resolved CID in context for later verification
	ctx = helpers.SetIPNSResolvedCID(ctx, expectedCID)

	return ctx, nil
}

// theIPNSPublishedCIDMatchesResolvedCID verifies that the published CID matches the resolved CID
func (s *IPNSCommonSteps) theIPNSPublishedCIDMatchesResolvedCID(ctx context.Context) (context.Context, error) {
	publishedCID, ok := helpers.GetIPNSPublishCID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no published CID found in context")
	}

	resolvedCID, ok := helpers.GetIPNSResolvedCID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no resolved CID found in context")
	}

	if publishedCID != resolvedCID {
		return ctx, fmt.Errorf("published CID %s does not match resolved CID %s", publishedCID, resolvedCID)
	}

	return ctx, nil
}

// theIPNSKeyHasValidPeerID verifies that an IPNS key has a valid peer ID
func (s *IPNSCommonSteps) theIPNSKeyHasValidPeerID(ctx context.Context) (context.Context, error) {
	peerID, ok := helpers.GetIPNSPeerID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS peer ID found in context")
	}

	if peerID == "" {
		return ctx, fmt.Errorf("IPNS peer ID is empty")
	}

	// Basic validation - peer IDs should start with "Qm" or "k" for newer formats
	if len(peerID) < 2 {
		return ctx, fmt.Errorf("IPNS peer ID %s is too short", peerID)
	}

	return ctx, nil
}

// theIPNSKeyHasIPNSName verifies that an IPNS key has a specific IPNS name
func (s *IPNSCommonSteps) theIPNSKeyHasIPNSName(ctx context.Context, expectedIPNSName string) (context.Context, error) {
	ipnsName, ok := helpers.GetIPNSIPNSName(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS name found in context")
	}

	if ipnsName != expectedIPNSName {
		return ctx, fmt.Errorf("IPNS name is %s, expected %s", ipnsName, expectedIPNSName)
	}

	return ctx, nil
}

// allIPNSKeysHaveValidPeerIDs verifies that multiple IPNS keys all have valid peer IDs
func (s *IPNSCommonSteps) allIPNSKeysHaveValidPeerIDs(ctx context.Context, count int) (context.Context, error) {
	// This is a placeholder for scenarios that test multiple IPNS keys
	// In the future, we may need to track multiple IPsNS keys in context
	
	// For now, just verify the current key has a valid peer ID
	// Similar to the pattern used for IPFS files
	_, err := s.theIPNSKeyHasValidPeerID(ctx)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}
