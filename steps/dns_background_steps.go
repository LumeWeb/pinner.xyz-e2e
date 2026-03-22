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
}

// theUserHasIPFSClientConnection verifies the user has an IPFS client connection
func (s *DNSBackgroundSteps) theUserHasIPFSClientConnection(ctx context.Context) (context.Context, error) {
	_, err := helpers.GetIPFSClient(ctx)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}
