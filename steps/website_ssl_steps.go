package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	"go.lumeweb.com/ipfs-sdk"
	"pinner.xyz-e2e/helpers"
)

// WebsiteSSLSteps provides step definitions for website SSL status operations
type WebsiteSSLSteps struct{}

// NewWebsiteSSLSteps creates a new WebsiteSSLSteps instance
func NewWebsiteSSLSteps() *WebsiteSSLSteps {
	return &WebsiteSSLSteps{}
}

// InitializeScenario registers SSL verification steps with godog
func (s *WebsiteSSLSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Website setup steps
	ctx.Step(`^the user has a website with DNS hosting enabled$`, s.theUserHasAWebsiteWithDNSHostingEnabled)
	
	// SSL status update operations (simulating Caddy webhook calls)
	ctx.Step(`^the SSL status is updated to "([^"]*)"$`, s.theSSLStatusIsUpdatedTo)
	ctx.Step(`^the SSL status update fails with error$`, s.theSSLStatusUpdateFailsWithError)
	
	// SSL status verification
	ctx.Step(`^the SSL status is stored in context$`, s.theSSLStatusIsStoredInContext)
	ctx.Step(`^the SSL certificate is issued$`, s.theSSLCertificateIsIssued)
	ctx.Step(`^the website has valid SSL certificate$`, s.theWebsiteHasValidSSLCertificate)
	ctx.Step(`^the website SSL status is "([^"]*)"$`, s.theWebsiteSSLStatusIs)
	ctx.Step(`^the SSL status contains error details$`, s.theSSLStatusContainsErrorDetails)
}

// theSSLStatusIsUpdatedTo updates SSL status via internal API endpoint
// Simulates Caddy webhook calling POST /internal/websites/:domain/ssl-status with X-Gateway-Secret
func (s *WebsiteSSLSteps) theSSLStatusIsUpdatedTo(ctx context.Context, status string) (context.Context, error) {
	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return ctx, fmt.Errorf("no website domain in context")
	}

	// Call internal API endpoint with gateway secret auth (handled automatically by SDK)
	timestamp := time.Now().Format(time.RFC3339)

	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return ctx, err
	}

	err = websiteService.UpdateSSLStatusInternal(ctx, domain, ipfs.SSLStatusUpdateRequest{
		Status:    status,
		Timestamp: &timestamp,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to update SSL status via internal API: %w", err)
	}

	// Store SSL status in context
	ctx = helpers.SetWebsiteSslStatus(ctx, status)

	return ctx, nil
}

// theSSLStatusUpdateFailsWithError updates SSL status to failed state with error message
func (s *WebsiteSSLSteps) theSSLStatusUpdateFailsWithError(ctx context.Context) (context.Context, error) {
	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return ctx, fmt.Errorf("no website domain in context")
	}

	// Call internal API endpoint with failed status and error message
	// Gateway secret auth is handled automatically by SDK
	errorMsg := "ACME challenge failed: DNS record not found"
	timestamp := time.Now().Format(time.RFC3339)

	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return ctx, err
	}

	err = websiteService.UpdateSSLStatusInternal(ctx, domain, ipfs.SSLStatusUpdateRequest{
		Error:     &errorMsg,
		Status:    "failed",
		Timestamp: &timestamp,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to update SSL status via internal API: %w", err)
	}

	// Store SSL status in context
	ctx = helpers.SetWebsiteSslStatus(ctx, "failed")

	return ctx, nil
}

// theWebsiteHasValidSSLCertificate verifies that the website has a valid SSL certificate
// This checks that SSL status is ready and certificate is issued
func (s *WebsiteSSLSteps) theWebsiteHasValidSSLCertificate(ctx context.Context) error {
	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}

	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	// Use SDK's wait method to poll for SSL ready status
	// Default timeout is applied by the SDK
	finalStatus, err := websiteService.WaitForSSLStatusReady(ctx, domain)
	if err != nil {
		return fmt.Errorf("SSL status did not reach ready state: %w", err)
	}

	if finalStatus != "ready" {
		return fmt.Errorf("expected SSL status 'ready', got '%s'", finalStatus)
	}

	return nil
}

// theUserHasAWebsiteWithDNSHostingEnabled creates a website with DNS hosting enabled as a precondition
func (s *WebsiteSSLSteps) theUserHasAWebsiteWithDNSHostingEnabled(ctx context.Context) (context.Context, error) {
	websiteSteps := NewWebsiteSteps()
	return websiteSteps.theUserCreatesAWebsiteWithDnsHostingEnabled(ctx)
}

// theSSLStatusIsStoredInContext verifies that SSL status was stored in context
func (s *WebsiteSSLSteps) theSSLStatusIsStoredInContext(ctx context.Context) error {
	_, ok := helpers.GetWebsiteSslStatus(ctx)
	if !ok {
		return fmt.Errorf("SSL status not stored in context")
	}
	return nil
}

// theSSLCertificateIsIssued verifies that the SSL certificate has been issued
// This checks that SSLIssuedAt timestamp is set on the SSL status
func (s *WebsiteSSLSteps) theSSLCertificateIsIssued(ctx context.Context) error {
	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}

	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	sslStatus, err := websiteService.GetSSLStatus(ctx, domain)
	if err != nil {
		return fmt.Errorf("failed to get SSL status: %w", err)
	}

	if sslStatus.Ssl == nil {
		return fmt.Errorf("SSL status is nil")
	}

	// Check if IssuedAt timestamp is set
	if sslStatus.Ssl.IssuedAt == nil || sslStatus.Ssl.IssuedAt.IsZero() {
		return fmt.Errorf("SSL certificate has not been issued (IssuedAt not set)")
	}

	return nil
}

// theWebsiteSSLStatusIs verifies the website SSL status matches expected value
func (s *WebsiteSSLSteps) theWebsiteSSLStatusIs(ctx context.Context, expectedStatus string) error {
	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}

	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	sslStatus, err := websiteService.GetSSLStatus(ctx, domain)
	if err != nil {
		return fmt.Errorf("failed to get SSL status: %w", err)
	}

	if sslStatus.Ssl == nil {
		return fmt.Errorf("SSL status is nil")
	}

	if sslStatus.Ssl.Status != expectedStatus {
		return fmt.Errorf("expected SSL status '%s', got '%s'", expectedStatus, sslStatus.Ssl.Status)
	}

	return nil
}

// theSSLStatusContainsErrorDetails verifies that SSL status contains error information
func (s *WebsiteSSLSteps) theSSLStatusContainsErrorDetails(ctx context.Context) error {
	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return fmt.Errorf("no website domain in context")
	}

	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	sslStatus, err := websiteService.GetSSLStatus(ctx, domain)
	if err != nil {
		return fmt.Errorf("failed to get SSL status: %w", err)
	}

	if sslStatus.Ssl == nil {
		return fmt.Errorf("SSL status is nil")
	}

	if sslStatus.Ssl.Status != "failed" {
		return fmt.Errorf("expected SSL status 'failed', got '%s'", sslStatus.Ssl.Status)
	}

	if sslStatus.Ssl.Error == nil || *sslStatus.Ssl.Error == "" {
		return fmt.Errorf("SSL status should contain error details but doesn't")
	}

	return nil
}
