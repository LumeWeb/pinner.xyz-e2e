package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"pinner.xyz-e2e/helpers"
)

// Common error constants for DNS background steps
var (
	errNoJWTToken     = fmt.Errorf("no JWT token found in context - user not authenticated")
	errIPFSClientNil  = fmt.Errorf("IPFS client is nil after creation")
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
	// Common authentication and IPFS connection steps used across all DNS features
	ctx.Step(`^the user has an authenticated API key$`, s.theUserHasAnAuthenticatedAPIKey)
	ctx.Step(`^the user has an IPFS client connection$`, s.theUserHasIPFSClientConnection)
}

// theUserHasAnAuthenticatedAPIKey ensures the user has an authenticated API key
func (s *DNSBackgroundSteps) theUserHasAnAuthenticatedAPIKey(ctx context.Context) (context.Context, error) {
	// Verify JWT token exists in context
	token, ok := helpers.GetJWTToken(ctx)
	if !ok || token == "" {
		return ctx, errNoJWTToken
	}

	// Verify IPFS client can be created with this token
	client, err := helpers.GetIPFSClient(ctx)
	if err != nil {
		return ctx, err
	}

	if client == nil {
		return ctx, errIPFSClientNil
	}

	return ctx, nil
}

// theUserHasIPFSClientConnection verifies the user has an IPFS client connection
func (s *DNSBackgroundSteps) theUserHasIPFSClientConnection(ctx context.Context) (context.Context, error) {
	_, err := helpers.GetIPFSClient(ctx)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}
