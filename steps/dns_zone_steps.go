package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"pinner.xyz-e2e/helpers"
)

const (
	// DefaultTestDomain is the standard test domain used in feature files
	// Using a constant makes it easier to identify and change later
	DefaultTestDomain = "test.example.com"
)

// DNSZoneSteps provides steps for DNS zone CRUD operations
type DNSZoneSteps struct{}

// NewDNSZoneSteps creates a new DNSZoneSteps instance
func NewDNSZoneSteps() *DNSZoneSteps {
	return &DNSZoneSteps{}
}

// InitializeScenario registers DNS zone steps with godog
func (s *DNSZoneSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the user creates a DNS zone with domain "([^"]*)"$`, s.theUserCreatesADNSZoneWithDomain)
	ctx.Step(`^the user has a DNS zone$`, s.theUserHasADNSZone)
	ctx.Step(`^the user gets DNS zone by ID$`, s.theUserGetsDNSZoneByID)
	ctx.Step(`^the user lists DNS zones$`, s.theUserListsDNSZones)
	ctx.Step(`^the user deletes the DNS zone$`, s.theUserDeletesTheDNSZone)
}

// theUserCreatesADNSZoneWithDomain creates a DNS zone with the specified domain
func (s *DNSZoneSteps) theUserCreatesADNSZoneWithDomain(ctx context.Context, domain string) (context.Context, error) {
	// For default test domain, use a unique domain to avoid conflicts from existing data
	if domain == DefaultTestDomain {
		domain = helpers.GenerateTestDomain()
	}
	
	nameservers := []string{}
	newCtx, _, err := helpers.CreateDNSZone(ctx, domain, nameservers)
	return newCtx, err
}

// theUserHasADNSZone creates a DNS zone for scenarios that require a zone precondition
func (s *DNSZoneSteps) theUserHasADNSZone(ctx context.Context) (context.Context, error) {
	domain := helpers.GenerateTestDomain()
	nameservers := []string{}
	newCtx, _, err := helpers.CreateDNSZone(ctx, domain, nameservers)
	return newCtx, err
}

// theUserGetsDNSZoneByID retrieves a DNS zone by ID from context
func (s *DNSZoneSteps) theUserGetsDNSZoneByID(ctx context.Context) (context.Context, error) {
	zoneID, err := helpers.RequireDNSZoneID(ctx)
	if err != nil {
		return ctx, err
	}

	zone, err := helpers.GetDNSZone(ctx, zoneID)
	if err != nil {
		return ctx, fmt.Errorf("failed to get DNS zone: %w", err)
	}

	if zone == nil {
		return ctx, fmt.Errorf("DNS zone is nil")
	}

	// Store zone details in context
	zoneIDStr := zoneID
	ctx = helpers.SetDNSZoneID(ctx, zoneIDStr)
	ctx = helpers.SetDNSZoneDomain(ctx, zone.Domain)

	return ctx, nil
}

// theUserListsDNSZones lists all DNS zones for the authenticated user
func (s *DNSZoneSteps) theUserListsDNSZones(ctx context.Context) (context.Context, error) {
	zones, err := helpers.ListDNSZones(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list DNS zones: %w", err)
	}

	if zones == nil {
		return ctx, fmt.Errorf("DNS zone list is nil")
	}

	// Store zone count in context for verification
	zoneCount := len(zones)
	ctx = helpers.SetContextValue(ctx, helpers.DNSZoneCleanupKey+"_count", zoneCount)

	return ctx, nil
}

// theUserDeletesTheDNSZone deletes the DNS zone stored in context
func (s *DNSZoneSteps) theUserDeletesTheDNSZone(ctx context.Context) (context.Context, error) {
	zoneID, err := helpers.RequireDNSZoneID(ctx)
	if err != nil {
		return ctx, err
	}

	err = helpers.DeleteDNSZone(ctx, zoneID)
	if err != nil {
		return ctx, fmt.Errorf("failed to delete DNS zone: %w", err)
	}

	return ctx, nil
}
