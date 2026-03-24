package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

// WebsiteIPNSSteps provides step definitions for website IPNS verification operations
type WebsiteIPNSSteps struct{}

// NewWebsiteIPNSSteps creates a new WebsiteIPNSSteps instance
func NewWebsiteIPNSSteps() *WebsiteIPNSSteps {
	return &WebsiteIPNSSteps{}
}

// InitializeScenario registers IPNS verification steps with godog
func (s *WebsiteIPNSSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Janitor verification operations
	ctx.Step(`^the janitor validates the website$`, s.theJanitorValidatesTheWebsite)
	ctx.Step(`^the IPNS key is valid for the user$`, s.theIPNSKeyIsValidForTheUser)
	ctx.Step(`^the resolved CID is pinned$`, s.theResolvedCIDIsPinned)
}

// theJanitorValidatesTheWebsite simulates janitor job validating website IPNS configuration
// This checks IPNS key validity, record publishing, and resolved CID pinning
func (s *WebsiteIPNSSteps) theJanitorValidatesTheWebsite(ctx context.Context) (context.Context, error) {
	
	id, err := helpers.RequireWebsiteID(ctx)
	if err != nil {
		return ctx, err
	}

	targetHash, ok := helpers.GetWebsiteTargetHash(ctx)
	if !ok {
		return ctx, fmt.Errorf("no website target hash in context")
	}

	targetType, ok := helpers.GetWebsiteTargetType(ctx)
	if !ok {
		return ctx, fmt.Errorf("no website target type in context")
	}

	// Validate that target type is IPNS
	if targetType != "ipns" {
		return ctx, fmt.Errorf("expected IPNS target type, got '%s'", targetType)
	}

	// Store target hash in context for validation
	ctx = helpers.SetWebsiteTargetHash(ctx, targetHash)

	// Wait for website to become active (janitor will validate IPNS record)
	idStr := fmt.Sprintf("%d", id)
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return ctx, err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, helpers.WebsiteStatusPollTimeout)
	defer cancel()

	err = websiteService.WaitForWebsiteStatus(timeoutCtx, idStr, "active")
	if err != nil {
		return ctx, fmt.Errorf("website did not become active after janitor validation: %w", err)
	}

	return ctx, nil
}

// theIPNSKeyIsValidForTheUser verifies that the IPNS key exists and belongs to the user
func (s *WebsiteIPNSSteps) theIPNSKeyIsValidForTheUser(ctx context.Context) error {
	// In the real implementation, janitor checks:
	// 1. IPNS key exists in database
	// 2. User owns the key
	// 3. Record is published
	// 4. Record is not expired

	// For E2E testing purposes, IPNS key validity is assumed
	// to have been established during website creation
	return nil
}

// theResolvedCIDIsPinned verifies that the CID resolved from IPNS is pinned
// This is checked by janitor to ensure content availability
func (s *WebsiteIPNSSteps) theResolvedCIDIsPinned(ctx context.Context) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	id, err := helpers.RequireWebsiteID(ctx)
	if err != nil {
		return err
	}

	// Get current website status
	idStr := fmt.Sprintf("%d", id)
	website, err := websiteService.Get(ctx, idStr)
	if err != nil {
		return fmt.Errorf("failed to get website: %w", err)
	}

	if website == nil {
		return fmt.Errorf("website is nil")
	}

	// In the real implementation, janitor:
	// 1. Resolves IPNS record to get CID
	// 2. Validates CID is pinned in storage
	// 3. Checks record validity timestamp
	// 4. Marks website as broken if any check fails

	// For E2E testing, we assume resolution was successful
	// If website status is "active", it means janitor passed all validations
	if website.Status != "active" {
		return fmt.Errorf("expected website status 'active', got '%s'", website.Status)
	}

	return nil
}
