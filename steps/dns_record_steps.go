package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	ipfs_sdk "go.lumeweb.com/ipfs-sdk"

	"pinner.xyz-e2e/helpers"
)

// DNSRecordSteps provides steps for DNS record CRUD operations
type DNSRecordSteps struct{}

// NewDNSRecordSteps creates a new DNSRecordSteps instance
func NewDNSRecordSteps() *DNSRecordSteps {
	return &DNSRecordSteps{}
}

// InitializeScenario registers DNS record steps with godog
func (s *DNSRecordSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the user creates an "([^"]*)" record named "([^"]*)" with value "([^"]*)"$`, s.theUserCreatesARecord)
	ctx.Step(`^the user has a DNS record "([^"]*)" of type "([^"]*)"$`, s.theUserHasADNSRecord)
	ctx.Step(`^the user has an "([^"]*)" record named "([^"]*)" with value "([^"]*)"$`, s.theUserHasAnARecordWithValue)
	ctx.Step(`^the user gets DNS record by name and type$`, s.theUserGetsDNSRecordByType)
	ctx.Step(`^the user lists DNS records$`, s.theUserListsDNSRecords)
	ctx.Step(`^the user updates the DNS record value to "([^"]*)"$`, s.theUserUpdatesTheDNSRecordValueTo)
	ctx.Step(`^the user deletes the DNS record$`, s.theUserDeletesTheDNSRecord)
	ctx.Step(`^the user deletes DNS records for "([^"]*)", "([^"]*)", and "([^"]*)"$`, s.theUserDeletesDNSRecordsFor)
	ctx.Step(`^the user creates "([^"]*)" records for "([^"]*)", "([^"]*)", and "([^"]*)" with value "([^"]*)"$`, s.theUserCreatesRecordsFor)
	ctx.Step(`^the user has "([^"]*)" records for "([^"]*)", "([^"]*)", and "([^"]*)"$`, s.theUserHasRecordsForAnd)
}

// theUserCreatesARecord creates a DNS record with the specified type, name, and value
func (s *DNSRecordSteps) theUserCreatesARecord(ctx context.Context, recordType string, name string, value string) (context.Context, error) {
	zoneID, err := helpers.RequireDNSZoneID(ctx)
	if err != nil {
		return ctx, err
	}

	record := helpers.BuildRecordRequest(recordType, name, value)
	newCtx, createdRecord, err := helpers.CreateDNSRecord(ctx, zoneID, record)
	if err != nil {
		return ctx, fmt.Errorf("failed to create DNS record: %w", err)
	}

	// Store the original input name (subdomain) for verification
	// The server returns FQDN, but verification expects the subdomain name
	newCtx = helpers.SetDNSRecordName(newCtx, name)
	newCtx = helpers.SetDNSRecordType(newCtx, createdRecord.Type)
	newCtx = helpers.StoreDNSRecordValue(newCtx, createdRecord.Content)

	return newCtx, nil
}

// theUserHasADNSRecord creates a DNS record preconditionally for scenarios
// Uses unique record names to prevent conflicts between test runs
func (s *DNSRecordSteps) theUserHasADNSRecord(ctx context.Context, name string, recordType string) (context.Context, error) {
	// Generate unique record name to prevent conflicts between test runs
	uniqueName := helpers.GenerateTestRecordName()
	newCtx, err := s.createDNSRecord(ctx, recordType, uniqueName, "192.0.2.1")
	if err != nil {
		return ctx, err
	}
	// Store the original parameter name for verification expectations
	newCtx = helpers.SetDNSRecordName(newCtx, name)
	return newCtx, nil
}

// theUserHasAnARecordWithValue creates a specific A record with the specified name and value
func (s *DNSRecordSteps) theUserHasAnARecordWithValue(ctx context.Context, recordType string, name string, value string) (context.Context, error) {
	newCtx, err := s.createDNSRecord(ctx, recordType, name, value)
	if err != nil {
		return ctx, err
	}
	// Store the original input name (subdomain) for verification expectations
	newCtx = helpers.SetDNSRecordName(newCtx, name)
	return newCtx, nil
}

// createDNSRecord is a helper that creates a DNS record with the given parameters
// This DRYs up the common record creation pattern
func (s *DNSRecordSteps) createDNSRecord(ctx context.Context, recordType string, name string, value string) (context.Context, error) {
	zoneID, err := helpers.RequireDNSZoneID(ctx)
	if err != nil {
		return ctx, err
	}

	record := helpers.BuildRecordRequest(recordType, name, value)
	newCtx, _, err := helpers.CreateDNSRecord(ctx, zoneID, record)
	if err != nil {
		return ctx, fmt.Errorf("failed to create DNS record: %w", err)
	}

	// Store original input name, type, and value for verification
	newCtx = helpers.SetDNSRecordName(newCtx, name)
	newCtx = helpers.SetDNSRecordType(newCtx, recordType)
	newCtx = helpers.StoreDNSRecordValue(newCtx, value)

	return newCtx, nil
}

// theUserGetsDNSRecordByType retrieves a DNS record by name and type
func (s *DNSRecordSteps) theUserGetsDNSRecordByType(ctx context.Context) (context.Context, error) {
	recordCtx, err := helpers.RequireDNSRecordContext(ctx)
	if err != nil {
		return ctx, err
	}

	record, err := helpers.GetDNSRecord(ctx, recordCtx.ZoneID, recordCtx.RecordFQDN, recordCtx.RecordType)
	if err != nil {
		return ctx, fmt.Errorf("failed to get DNS record: %w", err)
	}

	if record == nil {
		return ctx, fmt.Errorf("DNS record is nil")
	}

	// Store record details in context
	newCtx := helpers.StoreDNSRecordValue(ctx, record.Content)

	return newCtx, nil
}

// theUserListsDNSRecords lists all DNS records for a zone
func (s *DNSRecordSteps) theUserListsDNSRecords(ctx context.Context) (context.Context, error) {
	zoneID, err := helpers.RequireDNSZoneID(ctx)
	if err != nil {
		return ctx, err
	}

	records, err := helpers.ListDNSRecords(ctx, zoneID)
	if err != nil {
		return ctx, fmt.Errorf("failed to list DNS records: %w", err)
	}

	if records == nil {
		return ctx, fmt.Errorf("DNS record list is nil")
	}

	// Store record count in context for verification
	recordCount := len(records)
	ctx = helpers.SetContextValue(ctx, helpers.DNSRecordListCountKey, recordCount)

	return ctx, nil
}

// theUserUpdatesTheDNSRecordValueTo updates a DNS record with a new value
func (s *DNSRecordSteps) theUserUpdatesTheDNSRecordValueTo(ctx context.Context, newValue string) (context.Context, error) {
	recordCtx, err := helpers.RequireDNSRecordContext(ctx)
	if err != nil {
		return ctx, err
	}

	record := helpers.BuildRecordRequest(recordCtx.RecordType, recordCtx.RecordFQDN, newValue)
	newCtx, updatedRecord, err := helpers.UpdateDNSRecord(ctx, recordCtx.ZoneID, recordCtx.RecordFQDN, recordCtx.RecordType, record)
	if err != nil {
		return ctx, fmt.Errorf("failed to update DNS record: %w", err)
	}

	// Update the stored value with new value (may differ from input)
	newCtx = helpers.StoreDNSRecordValue(newCtx, updatedRecord.Content)

	return newCtx, nil
}

// theUserDeletesTheDNSRecord deletes the DNS record from context
func (s *DNSRecordSteps) theUserDeletesTheDNSRecord(ctx context.Context) (context.Context, error) {
	recordCtx, err := helpers.RequireDNSRecordContext(ctx)
	if err != nil {
		return ctx, err
	}

	err = helpers.DeleteDNSRecord(ctx, recordCtx.ZoneID, recordCtx.RecordFQDN, recordCtx.RecordType)
	if err != nil {
		return ctx, fmt.Errorf("failed to delete DNS record: %w", err)
	}

	return ctx, nil
}

// theUserDeletesDNSRecordsFor deletes multiple DNS records by name
func (s *DNSRecordSteps) theUserDeletesDNSRecordsFor(ctx context.Context, name1, name2, name3 string) (context.Context, error) {
	zoneID, err := helpers.RequireDNSZoneID(ctx)
	if err != nil {
		return ctx, err
	}

	// Get record type from the recently created records for consistency
	recordType := "A"
	if rt, ok := helpers.GetDNSRecordType(ctx); ok {
		recordType = rt
	}

	names := []string{name1, name2, name3}
	identifiers := make([]ipfs_sdk.RecordIdentifier, 0, len(names))

	for _, name := range names {
		identifiers = append(identifiers, ipfs_sdk.RecordIdentifier{
			Name: name,
			Type: recordType,
		})
	}

	_, err = helpers.BulkDeleteDNSRecords(ctx, zoneID, identifiers)
	if err != nil {
		return ctx, fmt.Errorf("failed to bulk delete DNS records: %w", err)
	}

	return ctx, nil
}

// theUserCreatesRecordsFor creates multiple DNS records in bulk
func (s *DNSRecordSteps) theUserCreatesRecordsFor(ctx context.Context, recordType, name1, name2, name3, value string) (context.Context, error) {
	return s.buildNameRecords(ctx, recordType, []string{name1, name2, name3}, value)
}

// theUserHasRecordsForAnd creates multiple DNS records in bulk with a default value
// Used as a precondition for scenarios testing bulk deletion operations
func (s *DNSRecordSteps) theUserHasRecordsForAnd(ctx context.Context, recordType, name1, name2, name3 string) (context.Context, error) {
	return s.buildNameRecords(ctx, recordType, []string{name1, name2, name3}, "192.0.2.1")
}

// buildNameRecords creates multiple DNS records from a list of names
// This DRYs up the common pattern used by bulk create operations
func (s *DNSRecordSteps) buildNameRecords(ctx context.Context, recordType string, names []string, value string) (context.Context, error) {
	zoneID, err := helpers.RequireDNSZoneID(ctx)
	if err != nil {
		return ctx, err
	}

	records := make([]ipfs_sdk.RecordRequest, 0, len(names))
	for _, name := range names {
		records = append(records, helpers.BuildRecordRequest(recordType, name, value))
	}

	_, err = helpers.BulkCreateDNSRecords(ctx, zoneID, records)
	if err != nil {
		return ctx, fmt.Errorf("failed to bulk create DNS records: %w", err)
	}

	// Store record type in context so bulk delete can use it
	ctx = helpers.SetDNSRecordType(ctx, recordType)

	return ctx, nil
}
