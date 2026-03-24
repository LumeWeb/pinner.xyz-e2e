package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
	"go.lumeweb.com/ipfs-sdk"
)

// WebsiteDNSSteps provides step definitions for website DNS validation operations
type WebsiteDNSSteps struct{}

// NewWebsiteDNSSteps creates a new WebsiteDNSSteps instance
func NewWebsiteDNSSteps() *WebsiteDNSSteps {
	return &WebsiteDNSSteps{}
}

// InitializeScenario registers DNS validation steps with godog
func (s *WebsiteDNSSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// DNS validation operations
	ctx.Step(`^the user validates the DNS records$`, s.theUserValidatesTheDNSRecords)
	ctx.Step(`^the website becomes active$`, s.theWebsiteBecomesActive)
	
	// DNS zone linking operations
	ctx.Step(`^the user creates a website with DNS zone linking$`, s.theUserCreatesAWebsiteWithDNSZoneLinking)
	ctx.Step(`^the user has a website without DNS hosting$`, s.theUserHasAWebsiteWithoutDNSHosting)
	ctx.Step(`^the website DNS zone ID is stored in context$`, s.theWebsiteDNSZoneIDIsStoredInContext)
	
	// DNS hosting toggle operations
	ctx.Step(`^the user updates the website to enable DNS hosting$`, s.theUserUpdatesTheWebsiteToEnableDNSHosting)
	ctx.Step(`^the user updates the website to disable DNS hosting$`, s.theUserUpdatesTheWebsiteToDisableDNSHosting)
}

// theUserValidatesTheDNSRecords triggers DNS validation for a website
// This simulates the POST /api/websites/:id/validate endpoint call
func (s *WebsiteDNSSteps) theUserValidatesTheDNSRecords(ctx context.Context) (context.Context, error) {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return ctx, err
	}

	id, err := helpers.RequireWebsiteID(ctx)
	if err != nil {
		return ctx, err
	}

	// Call the DNS validation endpoint
	idStr := fmt.Sprintf("%d", id)
	
	validateResp, err := websiteService.ValidateDNS(ctx, idStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to validate DNS: %w", err)
	}
	
	if !validateResp.Valid {
		return ctx, fmt.Errorf("DNS validation failed: %s", validateResp.Message)
	}

	// For managed DNS, wait for status to become active
	if dnsHostingEnabled, _ := helpers.GetWebsiteDnsHostingEnabled(ctx); dnsHostingEnabled {
		// Add a timeout context to prevent hanging
		timeoutCtx, cancel := context.WithTimeout(ctx, helpers.WebsiteStatusPollTimeout)
		defer cancel()
		err = websiteService.WaitForWebsiteStatus(timeoutCtx, idStr, "active")
		if err != nil {
			return ctx, fmt.Errorf("website did not become active after DNS validation: %w", err)
		}
	}

	// Update website status in context
	ctx = helpers.SetWebsiteStatus(ctx, "active")

	return ctx, nil
}

// theWebsiteBecomesActive verifies that website status is active
func (s *WebsiteDNSSteps) theWebsiteBecomesActive(ctx context.Context) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	id, err := helpers.RequireWebsiteID(ctx)
	if err != nil {
		return err
	}

	// Get the updated website status
	idStr := fmt.Sprintf("%d", id)
	website, err := websiteService.Get(ctx, idStr)
	if err != nil {
		return fmt.Errorf("failed to get website status: %w", err)
	}

	if website == nil {
		return fmt.Errorf("website is nil")
	}

	// Verify status is active
	if website.Status != "active" {
		return fmt.Errorf("expected website status 'active', got '%s'", website.Status)
	}

	ctx = helpers.SetWebsiteStatus(ctx, website.Status)

	return nil
}

// theUserCreatesAWebsiteWithDNSZoneLinking creates a website linked to an existing DNS zone
// This simulates the scenario where a user has an externally-managed DNS zone and wants
// to link their website to it for DNS hosting
func (s *WebsiteDNSSteps) theUserCreatesAWebsiteWithDNSZoneLinking(ctx context.Context) (context.Context, error) {
	return s.createWebsiteWithDnsHosting(ctx, true, true)
}

// theUserHasAWebsiteWithoutDNSHosting creates a website precondition without DNS hosting enabled
// Used as a Given step for scenarios that test DNS hosting toggle
func (s *WebsiteDNSSteps) theUserHasAWebsiteWithoutDNSHosting(ctx context.Context) (context.Context, error) {
	return s.createWebsiteWithDnsHosting(ctx, false, false)
}

// createWebsiteWithDnsHosting is an internal helper to create a website with specified DNS hosting
func (s *WebsiteDNSSteps) createWebsiteWithDnsHosting(ctx context.Context, dnsHostingEnabled bool, enableDevDnsZoneTracking bool) (context.Context, error) {
	domain := helpers.GenerateTestWebsiteDomain()
	
	// Upload sample content
	fixturePath := helpers.GetSampleWebsiteFixturePath()
	cid, err := helpers.IPFSPortalUploadDirFromFS(ctx, fixturePath)
	if err != nil {
		return ctx, fmt.Errorf("failed to upload sample website: %w", err)
	}
	
	// Wait for upload
	err = helpers.WaitForOperation(ctx, cid)
	if err != nil {
		return ctx, fmt.Errorf("operation failed to complete: %w", err)
	}
	
	// Create website
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return ctx, err
	}
	
	website, err := websiteService.CreateWithOptions(ctx, ipfs.WebsiteRequest{
		Domain:            domain,
		TargetHash:       cid,
		TargetType:       "ipfs",
		DnsHostingEnabled: &dnsHostingEnabled,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to create website: %w", err)
	}
	
	// Store website details in context
	ctx = helpers.StoreWebsiteID(ctx, website.Id)
	ctx = helpers.SetWebsiteDomain(ctx, website.Domain)
	ctx = helpers.SetWebsiteTargetHash(ctx, website.TargetHash)
	ctx = helpers.SetWebsiteTargetType(ctx, website.TargetType)
	ctx = helpers.SetWebsiteStatus(ctx, website.Status)
	ctx = helpers.SetWebsiteDnsHostingEnabled(ctx, website.DnsHostingEnabled)
	
	// Add DNS records to dev server for DNS validation
	// DNS validation is required regardless of DNS hosting status to prove domain ownership.
	// The dnsHostingEnabled flag only controls whether the portal manages DNS, not whether
	// validation is required.
	// Build DNS link considering IPNS auto-conversion
	dnslink := helpers.BuildWebsiteDNSLink(website.TargetHash, website.TargetType)
	err = helpers.AddWebsiteDNSZoneToDevServer(ctx, website.Domain, dnslink, website.ValidationToken)
	if err != nil {
		return ctx, fmt.Errorf("failed to add DNS zone to dev server: %w", err)
	}
	
	// Store DNS zone ID if DNS hosting enabled and tracking requested
	if enableDevDnsZoneTracking && dnsHostingEnabled {
		if website.DnsZoneId != nil {
			zoneIDStr := fmt.Sprintf("%d", *website.DnsZoneId)
			ctx = helpers.SetWebsiteDevDnsZoneID(ctx, zoneIDStr)
		} else {
			// Attempt to get zone ID from context if available
			zoneID, hasExistingZone := helpers.GetDNSZoneID(ctx)
			if hasExistingZone {
				ctx = helpers.SetWebsiteDevDnsZoneID(ctx, zoneID)
			}
		}
	}
	
	// Add to cleanup
	ctx = helpers.AddWebsiteIDCleanup(ctx, website.Id)
	
	return ctx, nil
}

// theWebsiteDNSZoneIDIsStoredInContext verifies that DNS zone ID is available in context
// This ensures the zone linking was successful and the zone ID is retrievable
func (s *WebsiteDNSSteps) theWebsiteDNSZoneIDIsStoredInContext(ctx context.Context) error {
	zoneID, ok := helpers.GetWebsiteDevDnsZoneID(ctx)
	if !ok {
		return fmt.Errorf("DNS zone ID not found in context - zone linking may have failed")
	}
	
	if zoneID == "" {
		return fmt.Errorf("DNS zone ID is empty in context")
	}
	
	return nil
}

// theUserUpdatesTheWebsiteToEnableDNSHosting updates a website to enable DNS hosting
// This is used when a user toggles DNS hosting on an existing website
// Note: The portal API's Update endpoint doesn't currently support changing DNS HostingEnabled status.
// This step calls Update to follow the API pattern, but actual DNS hosting toggle may require portal changes.
func (s *WebsiteDNSSteps) theUserUpdatesTheWebsiteToEnableDNSHosting(ctx context.Context) (context.Context, error) {
	return s.updateWebsiteWithDnsHosting(ctx, true)
}

// theUserUpdatesTheWebsiteToDisableDNSHosting updates a website to disable DNS hosting
// This is used when a user toggles off DNS hosting on an existing website
func (s *WebsiteDNSSteps) theUserUpdatesTheWebsiteToDisableDNSHosting(ctx context.Context) (context.Context, error) {
	return s.updateWebsiteWithDnsHosting(ctx, false)
}

// updateWebsiteWithDnsHosting is an internal helper to update a website with specified DNS hosting state
// Consolidates common update logic to avoid code duplication
func (s *WebsiteDNSSteps) updateWebsiteWithDnsHosting(ctx context.Context, dnsHostingEnabled bool) (context.Context, error) {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return ctx, err
	}

	id, err := helpers.RequireWebsiteID(ctx)
	if err != nil {
		return ctx, err
	}

	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return ctx, fmt.Errorf("no website domain in context")
	}

	targetHash, ok := helpers.GetWebsiteTargetHash(ctx)
	if !ok {
		return ctx, fmt.Errorf("no website target hash in context")
	}

	targetType, ok := helpers.GetWebsiteTargetType(ctx)
	if !ok {
		return ctx, fmt.Errorf("no website target type in context")
	}

	idStr := fmt.Sprintf("%d", id)
	
	// Update website with DNS hosting setting
	updatedWebsite, err := websiteService.UpdateWithOptions(ctx, idStr, ipfs.WebsiteRequest{
		Domain:            domain,
		TargetHash:        targetHash,
		TargetType:        targetType,
		DnsHostingEnabled: &dnsHostingEnabled,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to update website: %w", err)
	}

	// Update status if changed
	if updatedWebsite.Status != "" {
		ctx = helpers.SetWebsiteStatus(ctx, updatedWebsite.Status)
	}
	
	// Update DNS hosting enabled state
	ctx = helpers.SetWebsiteDnsHostingEnabled(ctx, dnsHostingEnabled)
	
	// Manage DNS records in dev server
	// Add or update DNS records for validation regardless of DNS hosting status.
	// The dnsHostingEnabled flag controls whether the portal manages the DNS records,
	// but DNS validation always requires records to be present.
	dnslink := helpers.BuildWebsiteDNSLink(updatedWebsite.TargetHash, updatedWebsite.TargetType)
	err = helpers.AddWebsiteDNSZoneToDevServer(ctx, updatedWebsite.Domain, dnslink, updatedWebsite.ValidationToken)
	if err != nil {
		return ctx, fmt.Errorf("failed to add DNS zone to dev server: %w", err)
	}
	// Store the zone ID if DNS hosting is enabled
	if dnsHostingEnabled && updatedWebsite.DnsZoneId != nil {
		zoneIDStr := fmt.Sprintf("%d", *updatedWebsite.DnsZoneId)
		ctx = helpers.SetWebsiteDevDnsZoneID(ctx, zoneIDStr)
	}

	return ctx, nil
}
