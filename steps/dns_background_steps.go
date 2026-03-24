package steps

import (
	"context"

	"github.com/cucumber/godog"

	"pinner.xyz-e2e/helpers"
)

// DNSBackgroundSteps provides shared background steps for DNS testing
// These are common steps that are used across all DNS feature files
type DNSBackgroundSteps struct{}

// NewDNSBackgroundSteps creates a new DNSBackgroundSteps instance
func NewDNSBackgroundSteps() *DNSBackgroundSteps {
	return &DNSBackgroundSteps{}
}

// InitializeScenario registers DNS background steps with godog
func (s *DNSBackgroundSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Common IPFS connection steps used across all DNS features
	ctx.Step(`^the user has an IPFS client connection$`, s.theUserHasIPFSClientConnection)
	
	// DNS dev server client initialization
	ctx.Step(`^the DNS dev server is available$`, s.theDNSDevServerIsAvailable)
}

// theUserHasIPFSClientConnection verifies the user has an IPFS client connection
func (s *DNSBackgroundSteps) theUserHasIPFSClientConnection(ctx context.Context) (context.Context, error) {
	_, err := helpers.GetIPFSClient(ctx)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

// theDNSDevServerIsAvailable ensures the DNS dev server client is initialized
// and the DNS dev server API is responsive
func (s *DNSBackgroundSteps) theDNSDevServerIsAvailable(ctx context.Context) (context.Context, error) {
	// Ensure DNS dev client exists in context
	ctx = helpers.EnsureDNSDevClient(ctx)
	
	// Check server health
	if err := helpers.CheckDNSDevServerHealth(ctx); err != nil {
		return ctx, err
	}
	
	return ctx, nil
}
