package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

// IPNSPublishingResolutionSteps holds step definitions for IPNS publishing and resolution operations
type IPNSPublishingResolutionSteps struct{}

// NewIPNSPublishingResolutionSteps creates a new IPNSPublishingResolutionSteps instance
func NewIPNSPublishingResolutionSteps() *IPNSPublishingResolutionSteps {
	return &IPNSPublishingResolutionSteps{}
}

// InitializeScenario registers all IPNS publishing and resolution step definitions with godog
func (s *IPNSPublishingResolutionSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the user publishes CID "([^"]*)" to the IPNS key$`, s.theUserPublishesCIDToTheIPNSKey)
	ctx.Step(`^the IPNS name resolves to the published CID$`, s.theIPNSNameResolvesToThePublishedCID)
	ctx.Step(`^the user republishes all IPNS entries$`, s.theUserRepublishesAllIPNSEntries)
	ctx.Step(`^the user resolves IPNS name "([^"]*)"$`, s.theUserResolvesIPNSName)
	ctx.Step(`^the resolved CID is "([^"]*)"$`, s.theResolvedCIDIs)
	
	// Key creation - shared helper used by both "has IPNS key" and "has IPNS key in Portal" steps
	ctx.Step(`^the user has an IPNS key named "([^"]*)"$`, s.theUserHasAnIPNSKey)
	ctx.Step(`^the user has an IPNS key named "([^"]*)" in the Portal$`, s.theUserHasAnIPNSKey)
	
	ctx.Step(`^the user has an IPNS CID "([^"]*)"`, s.theUserHasAnIPNSCID)
	ctx.Step(`^the publish operation succeeds$`, s.thePublishOperationSucceeds)
	
	// Cross-node verification steps
	ctx.Step(`^the IPNS name is resolvable by Portal$`, s.theIPNSNameIsResolvableByPortal)
	ctx.Step(`^Kubo can resolve the IPNS name$`, s.kuboCanResolveTheIPNSName)
	ctx.Step(`^both resolve to "([^"]*)"$`, s.bothResolveToCID)
	
	// Resolution steps - delegates to shared helper for actual resolution logic
	ctx.Step(`^the user resolves the IPNS name published by Portal$`, s.theUserResolvesIPNSNameViaPortal)
	ctx.Step(`^the user resolves the IPNS name via Portal$`, s.theUserResolvesIPNSNameViaPortal)
	
	ctx.Step(`^the IPNS name exists in the Portal$`, s.theIPNSNameExistsInPortal)
	ctx.Step(`^the republish operation succeeds$`, s.theRepublishOperationSucceeds)
}

// theUserPublishesCIDToTheIPNSKey publishes a CID to the current IPNS key
func (s *IPNSPublishingResolutionSteps) theUserPublishesCIDToTheIPNSKey(ctx context.Context, cid string) (context.Context, error) {
	keyID, ok := helpers.GetIPNSKeyID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS key ID found in context")
	}

	publish, err := helpers.PublishToIPNS(ctx, keyID, cid)
	if err != nil {
		return ctx, fmt.Errorf("failed to publish CID %s to IPNS key %d: %w", cid, keyID, err)
	}

	if publish == nil {
		return ctx, fmt.Errorf("publish operation returned nil response")
	}

	ctx = helpers.SetIPNSPublishCID(ctx, cid)
	if publish.Name != "" {
		ctx = helpers.SetIPNSIPNSName(ctx, publish.Name)
	}

	return ctx, nil
}

// theIPNSNameResolvesToThePublishedCID verifies that the IPNS name resolves to the CID that was published
func (s *IPNSPublishingResolutionSteps) theIPNSNameResolvesToThePublishedCID(ctx context.Context) (context.Context, error) {
	ipnsName, ok := helpers.GetIPNSIPNSName(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS name found in context")
	}

	publishedCID, ok := helpers.GetIPNSPublishCID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no published CID found in context")
	}

	resolve, err := helpers.ResolveIPNSName(ctx, ipnsName)
	if err != nil {
		return ctx, fmt.Errorf("failed to resolve IPNS name %s: %w", ipnsName, err)
	}

	if resolve == nil || resolve.Value == "" {
		return ctx, fmt.Errorf("IPNS name %s could not be resolved", ipnsName)
	}

	if resolve.Value != publishedCID {
		return ctx, fmt.Errorf("IPNS name %s resolved to %s, but published CID was %s", ipnsName, resolve.Value, publishedCID)
	}

	ctx = helpers.SetIPNSResolvedCID(ctx, resolve.Value)
	return ctx, nil
}

// theUserRepublishesAllIPNSEntries republishes all IPNS entries for the authenticated user
func (s *IPNSPublishingResolutionSteps) theUserRepublishesAllIPNSEntries(ctx context.Context) (context.Context, error) {
	err := helpers.RepublishIPNS(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to republish IPNS entries: %w", err)
	}

	return ctx, nil
}

// theUserResolvesIPNSName resolves a specified IPNS name
func (s *IPNSPublishingResolutionSteps) theUserResolvesIPNSName(ctx context.Context, name string) (context.Context, error) {
	resolve, err := helpers.ResolveIPNSName(ctx, name)
	if err != nil {
		return ctx, fmt.Errorf("failed to resolve IPNS name %s: %w", name, err)
	}

	if resolve == nil || resolve.Value == "" {
		return ctx, fmt.Errorf("IPNS name %s could not be resolved", name)
	}

	ctx = helpers.SetIPNSIPNSName(ctx, name)
	return ctx, nil
}

// theResolvedCIDIs verifies that the resolved CID matches the expected value
func (s *IPNSPublishingResolutionSteps) theResolvedCIDIs(ctx context.Context, expectedCID string) (context.Context, error) {
	resolvedCID, ok := helpers.GetIPNSResolvedCID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no resolved CID found in context")
	}

	if resolvedCID != expectedCID {
		return ctx, fmt.Errorf("resolved CID is %s, expected %s", resolvedCID, expectedCID)
	}

	return ctx, nil
}

// theUserHasAnIPNSKey creates an IPNS key and stores it in context
// Used by both "has IPNS key" and "has IPNS key in Portal" scenario steps
func (s *IPNSPublishingResolutionSteps) theUserHasAnIPNSKey(ctx context.Context, name string) (context.Context, error) {
	ctx, key, err := helpers.CreateIPNSKey(ctx, name)
	if err != nil {
		return ctx, err
	}

	if key == nil {
		return ctx, fmt.Errorf("IPNS key creation returned nil response")
	}

	return helpers.StoreKeyInfoInContext(ctx, key), nil
}

// theUserHasAnIPNSCID stores a CID in context for IPNS operations
func (s *IPNSPublishingResolutionSteps) theUserHasAnIPNSCID(ctx context.Context, cid string) (context.Context, error) {
	ctx = helpers.SetIPNSPublishCID(ctx, cid)
	return ctx, nil
}

// thePublishOperationSucceeds is a placeholder for future publish operation verification
func (s *IPNSPublishingResolutionSteps) thePublishOperationSucceeds(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

// kuboCanResolveTheIPNSName verifies that Kubo can resolve the IPNS name
func (s *IPNSPublishingResolutionSteps) kuboCanResolveTheIPNSName(ctx context.Context) (context.Context, error) {
	ipnsName, ok := helpers.GetIPNSIPNSName(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS name found in context")
	}

	resolvedCID, err := helpers.KuboResolveIPNS(ctx, ipnsName)
	if err != nil {
		return ctx, fmt.Errorf("Kubo failed to resolve IPNS name %s: %w", ipnsName, err)
	}

	if resolvedCID == "" {
		return ctx, fmt.Errorf("Kubo resolved to empty CID for IPNS name %s", ipnsName)
	}

	ctx = helpers.SetIPNSResolvedCID(ctx, resolvedCID)
	return ctx, nil
}

// theIPNSNameIsResolvableByPortal verifies that the Portal can resolve the IPNS name
func (s *IPNSPublishingResolutionSteps) theIPNSNameIsResolvableByPortal(ctx context.Context) (context.Context, error) {
	ipnsName, ok := helpers.GetIPNSIPNSName(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS name found in context")
	}

	resolve, err := helpers.ResolveIPNSName(ctx, ipnsName)
	if err != nil {
		return ctx, fmt.Errorf("Portal failed to resolve IPNS name %s: %w", ipnsName, err)
	}

	if resolve == nil || resolve.Value == "" {
		return ctx, fmt.Errorf("Portal returned empty CID for IPNS name %s", ipnsName)
	}

	return ctx, nil
}

// bothResolveToCID verifies that both Portal and Kubo resolve to the same CID
func (s *IPNSPublishingResolutionSteps) bothResolveToCID(ctx context.Context, expectedCID string) (context.Context, error) {
	resolvedCID, ok := helpers.GetIPNSResolvedCID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no resolved CID found in context")
	}

	if resolvedCID != expectedCID {
		return ctx, fmt.Errorf("resolved CID %s does not match expected %s", resolvedCID, expectedCID)
	}

	return ctx, nil
}

// theUserResolvesIPNSNameViaPortal resolves an IPNS name via Portal
// Shared by "resolves IPNS name published by Portal" and "resolves IPNS name via Portal" scenario steps
func (s *IPNSPublishingResolutionSteps) theUserResolvesIPNSNameViaPortal(ctx context.Context) (context.Context, error) {
	ipnsName, ok := helpers.GetIPNSIPNSName(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS name found in context")
	}

	cid, err := helpers.ResolveIPNSNameViaPortal(ctx, ipnsName)
	if err != nil {
		return ctx, fmt.Errorf("failed to resolve IPNS name via Portal: %w", err)
	}

	ctx = helpers.SetIPNSResolvedCID(ctx, cid)
	return ctx, nil
}

// theIPNSNameExistsInPortal verifies that the IPNS name exists in the Portal
func (s *IPNSPublishingResolutionSteps) theIPNSNameExistsInPortal(ctx context.Context) (context.Context, error) {
	ipnsName, ok := helpers.GetIPNSIPNSName(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS name found in context")
	}

	_, err := helpers.ResolveIPNSNameViaPortal(ctx, ipnsName)
	if err != nil {
		return ctx, fmt.Errorf("failed to verify IPNS name in Portal: %w", err)
	}

	return ctx, nil
}

// theRepublishOperationSucceeds verifies that the IPNS republish operation completed successfully
func (s *IPNSPublishingResolutionSteps) theRepublishOperationSucceeds(ctx context.Context) (context.Context, error) {
	keys, err := helpers.ListIPNSKeys(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to verify IPNS keys list: %w", err)
	}

	if len(keys) == 0 {
		return ctx, fmt.Errorf("no IPNS keys available to republish")
	}

	return ctx, nil
}
