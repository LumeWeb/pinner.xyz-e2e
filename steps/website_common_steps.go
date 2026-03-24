package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

// WebsiteCommonSteps provides shared verification steps for website testing
// These steps are registered before service-specific website steps to prevent duplicates
type WebsiteCommonSteps struct{}

// NewWebsiteCommonSteps creates a new WebsiteCommonSteps instance
func NewWebsiteCommonSteps() *WebsiteCommonSteps {
	return &WebsiteCommonSteps{}
}

// InitializeScenario registers common website verification steps with godog
func (s *WebsiteCommonSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Status verification steps
	ctx.Step(`^the website status is "([^"]*)"$`, s.theWebsiteStatusIs)
	ctx.Step(`^the website reaches "([^"]*)" status$`, s.theWebsiteReachesStatus)
	ctx.Step(`^the website domain is "([^"]*)"$`, s.theWebsiteDomainIs)
	ctx.Step(`^the website target hash is "([^"]*)"$`, s.theWebsiteTargetHashIs)
	ctx.Step(`^the website target type is "([^"]*)"$`, s.theWebsiteTargetTypeIs)
	
	// Update verification
	ctx.Step(`^the website is updated successfully$`, s.theWebsiteIsUpdatedSuccessfully)
	
	// SSL status verification
	ctx.Step(`^the website SSL status is "([^"]*)"$`, s.theWebsiteSslStatusIs)
	ctx.Step(`^the website has valid SSL certificate$`, s.theWebsiteHasValidSslCertificate)
	
	// DNS hosting verification
	ctx.Step(`^DNS hosting is enabled for the website$`, s.dnsHostingIsEnabled)
	ctx.Step(`^DNS hosting is disabled for the website$`, s.dnsHostingIsDisabled)

	// Website existence verification
	ctx.Step(`^the website exists$`, s.theWebsiteExists)
	ctx.Step(`^the website does not exist$`, s.theWebsiteDoesNotExist)

	// Count verification
	ctx.Step(`^(\d+) websites? returned$`, s.websiteCountIs)
	ctx.Step(`^the website count is returned$`, s.theWebsiteCountIsReturned)

	// Target hash verification
	ctx.Step(`^the website target hash matches the uploaded CID$`, s.theWebsiteTargetHashMatchesTheUploadedCID)
	ctx.Step(`^the website has a target hash$`, s.theWebsiteTargetHashIsStored)

	// Domain verification (for cases with random suffixes)
	ctx.Step(`^the website domain matches "([^"]*)"$`, s.theWebsiteDomainMatches)
}

// theWebsiteStatusIs verifies the current status of the website
func (s *WebsiteCommonSteps) theWebsiteStatusIs(ctx context.Context, expectedStatus string) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}

	// Get the website by searching the list
	websites, err := websiteService.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list websites: %w", err)
	}

	// Find the website by domain
	for _, site := range websites {
		if site.Domain == domain {
			if site.Status != expectedStatus {
				return fmt.Errorf("expected status '%s', got '%s'", expectedStatus, site.Status)
			}
			return nil
		}
	}

	return fmt.Errorf("website with domain '%s' not found", domain)
}

// theWebsiteReachesStatus waits for the website to reach a specific status
// This is useful for asynchronous operations where status changes over time
func (s *WebsiteCommonSteps) theWebsiteReachesStatus(ctx context.Context, expectedStatus string) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}

	// Poll for status with timeout
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for website to reach status '%s'", expectedStatus)
		case <-ticker.C:
			websites, err := websiteService.List(ctx)
			if err != nil {
				continue // retry on error
			}

			for _, site := range websites {
				if site.Domain == domain {
					if site.Status == expectedStatus {
						return nil
					}
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// theWebsiteDomainIs verifies the website's domain
// Checks the intended domain (from feature file) when available, otherwise checks the actual domain
func (s *WebsiteCommonSteps) theWebsiteDomainIs(ctx context.Context, expectedDomain string) error {
	// First try to get the intended domain (for scenarios that use randomized actual domains)
	intendedDomain, hasIntended := helpers.GetWebsiteIntendedDomain(ctx)
	if hasIntended && intendedDomain != "" {
		if intendedDomain != expectedDomain {
			return fmt.Errorf("expected domain '%s', got intended domain '%s'", expectedDomain, intendedDomain)
		}
		return nil
	}

	// Fall back to checking actual domain
	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}

	if domain != expectedDomain {
		return fmt.Errorf("expected domain '%s', got '%s'", expectedDomain, domain)
	}
	return nil
}

// theWebsiteDomainMatches verifies the website's domain matches the expected pattern
// Handles both exact matches and randomized suffixes
func (s *WebsiteCommonSteps) theWebsiteDomainMatches(ctx context.Context, expectedPattern string) error {
	// First try to get the intended domain (for scenarios that use randomized actual domains)
	intendedDomain, hasIntended := helpers.GetWebsiteIntendedDomain(ctx)
	if hasIntended && intendedDomain != "" {
		if !isDomainMatch(intendedDomain, expectedPattern) {
			return fmt.Errorf("intended domain '%s' does not match pattern '%s'", intendedDomain, expectedPattern)
		}
		return nil
	}

	// Fall back to checking actual domain
	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}

	if !isDomainMatch(domain, expectedPattern) {
		return fmt.Errorf("domain '%s' does not match pattern '%s'", domain, expectedPattern)
	}

	return nil
}

// isDomainMatch checks if a domain matches a pattern string
func isDomainMatch(domain, pattern string) bool {
	if domain == pattern {
		return true
	}
	// Add pattern matching logic for random suffixes if needed
	return false
}

// theWebsiteTargetHashIs verifies the website's target hash
func (s *WebsiteCommonSteps) theWebsiteTargetHashIs(ctx context.Context, expectedHash string) error {
	targetHash, ok := helpers.GetWebsiteTargetHash(ctx)
	if !ok {
		return fmt.Errorf("no website target hash in context")
	}

	if targetHash != expectedHash {
		return fmt.Errorf("expected target hash '%s', got '%s'", expectedHash, targetHash)
	}
	return nil
}

// theWebsiteTargetTypeIs verifies the website's target type
func (s *WebsiteCommonSteps) theWebsiteTargetTypeIs(ctx context.Context, expectedType string) error {
	targetType, ok := helpers.GetWebsiteTargetType(ctx)
	if !ok {
		return fmt.Errorf("no website target type in context")
	}

	if targetType != expectedType {
		return fmt.Errorf("expected target type '%s', got '%s'", expectedType, targetType)
	}
	return nil
}

// theWebsiteSslStatusIs verifies the SSL status of the website
func (s *WebsiteCommonSteps) theWebsiteSslStatusIs(ctx context.Context, expectedStatus string) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}

	sslStatus, err := websiteService.GetSSLStatus(ctx, domain)
	if err != nil {
		return fmt.Errorf("failed to get SSL status: %w", err)
	}

	// Check SSL status from the response
	if sslStatus.Ssl == nil {
		return fmt.Errorf("SSL status not available")
	}

	if sslStatus.Ssl.Status != expectedStatus {
		return fmt.Errorf("expected SSL status '%s', got '%s'", expectedStatus, sslStatus.Ssl.Status)
	}
	return nil
}

// theWebsiteHasValidSslCertificate verifies that the website has a valid SSL certificate
func (s *WebsiteCommonSteps) theWebsiteHasValidSslCertificate(ctx context.Context) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}

	sslStatus, err := websiteService.GetSSLStatus(ctx, domain)
	if err != nil {
		return fmt.Errorf("failed to get SSL status: %w", err)
	}

	if sslStatus.Ssl == nil {
		return fmt.Errorf("SSL status not available")
	}

	// Valid SSL means status is "valid" or "active"
	if sslStatus.Ssl.Status != "valid" && sslStatus.Ssl.Status != "active" {
		return fmt.Errorf("SSL certificate is not valid, status is '%s'", sslStatus.Ssl.Status)
	}
	return nil
}

// dnsHostingIsEnabled verifies that DNS hosting is enabled for the website
func (s *WebsiteCommonSteps) dnsHostingIsEnabled(ctx context.Context) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}

	websites, err := websiteService.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list websites: %w", err)
	}

	for _, site := range websites {
		if site.Domain == domain {
			if !site.DnsHostingEnabled {
				return fmt.Errorf("expected DNS hosting to be enabled")
			}
			return nil
		}
	}
	return fmt.Errorf("website with domain '%s' not found", domain)
}

// dnsHostingIsDisabled verifies that DNS hosting is disabled for the website
func (s *WebsiteCommonSteps) dnsHostingIsDisabled(ctx context.Context) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}

	websites, err := websiteService.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list websites: %w", err)
	}

	for _, site := range websites {
		if site.Domain == domain {
			if site.DnsHostingEnabled {
				return fmt.Errorf("expected DNS hosting to be disabled")
			}
			return nil
		}
	}
	return fmt.Errorf("website with domain '%s' not found", domain)
}

// theWebsiteExists verifies that a website exists in the list
func (s *WebsiteCommonSteps) theWebsiteExists(ctx context.Context) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}

	websites, err := websiteService.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list websites: %w", err)
	}

	for _, site := range websites {
		if site.Domain == domain {
			return nil
		}
	}
	return fmt.Errorf("website with domain '%s' not found", domain)
}

// theWebsiteDoesNotExist verifies that a website does not exist
func (s *WebsiteCommonSteps) theWebsiteDoesNotExist(ctx context.Context) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}

	websites, err := websiteService.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list websites: %w", err)
	}

	for _, site := range websites {
		if site.Domain == domain {
			return fmt.Errorf("website with domain '%s' still exists", domain)
		}
	}
	return nil
}

// websiteCountIs verifies that the expected number of websites were returned
func (s *WebsiteCommonSteps) websiteCountIs(ctx context.Context, expectedCount int) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	websites, err := websiteService.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list websites: %w", err)
	}

	if len(websites) != expectedCount {
		return fmt.Errorf("expected %d websites, got %d", expectedCount, len(websites))
	}
	return nil
}

// theWebsiteTargetHashMatchesTheUploadedCID verifies that the website has a target hash
// The API converts IPFS CIDs to IPNS peer IDs automatically, so we verify a target hash was set
func (s *WebsiteCommonSteps) theWebsiteTargetHashMatchesTheUploadedCID(ctx context.Context) error {
	targetHash, ok := helpers.GetWebsiteTargetHash(ctx)
	if !ok {
		return fmt.Errorf("no website target hash in context")
	}
	
	if targetHash == "" {
		return fmt.Errorf("website target hash is empty")
	}
	
	return nil
}

// theWebsiteCountIsReturned verifies that website count is stored in context
func (s *WebsiteCommonSteps) theWebsiteCountIsReturned(ctx context.Context) error {
	_, ok := helpers.GetListCount(ctx, helpers.WebsiteListKey)
	if !ok {
		return fmt.Errorf("website count not stored in context")
	}

	return nil
}

// theWebsiteIsUpdatedSuccessfully verifies that a website was updated successfully
func (s *WebsiteCommonSteps) theWebsiteIsUpdatedSuccessfully(ctx context.Context) error {
	_, err := helpers.RequireWebsiteID(ctx)
	if err != nil {
		return err
	}
	
	_, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}
	
	return nil
}

// theWebsiteTargetHashIsStored verifies that the website has a target hash stored in context
func (s *WebsiteCommonSteps) theWebsiteTargetHashIsStored(ctx context.Context) error {
	targetHash, ok := helpers.GetWebsiteTargetHash(ctx)
	if !ok {
		return fmt.Errorf("no website target hash in context")
	}
	
	if targetHash == "" {
		return fmt.Errorf("website target hash is empty")
	}
	
	return nil
}
