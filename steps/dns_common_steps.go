package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"pinner.xyz-e2e/helpers"
)

// DNSCommonSteps provides shared verification steps for DNS testing
type DNSCommonSteps struct{}

// NewDNSCommonSteps creates a new DNSCommonSteps instance
func NewDNSCommonSteps() *DNSCommonSteps {
	return &DNSCommonSteps{}
}

// InitializeScenario registers DNS common steps with godog
func (s *DNSCommonSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Zone verification steps
	ctx.Step(`^the DNS zone is retrieved successfully`, s.theDNSZoneIsRetrievedSuccessfully)
	ctx.Step(`^the zone domain is stored in context`, s.theZoneDomainIsVerified)
	ctx.Step(`^the zone cannot be retrieved`, s.theZoneCannotBeRetrieved)

	// Record verification steps
	ctx.Step(`^the DNS record is retrieved successfully`, s.theDNSRecordIsRetrievedSuccessfully)
	ctx.Step(`^the record name is "([^"]*)"$`, s.theRecordNameIs)
	ctx.Step(`^the record type is "([^"]*)"$`, s.theRecordTypeIs)
	ctx.Step(`^the record value is "([^"]*)"$`, s.theRecordValueIs)
	ctx.Step(`^the DNS record is created successfully`, s.theDNSRecordIsCreatedSuccessfully)
	ctx.Step(`^the DNS record is updated successfully`, s.theDNSRecordIsUpdatedSuccessfully)
	ctx.Step(`^the record cannot be retrieved`, s.theRecordCannotBeRetrieved)

	// List verification steps
	ctx.Step(`^the user receives a zone list`, s.theUserReceivesAZoneList)
	ctx.Step(`^the zone list contains the zone count`, s.theZoneListContainsTheZoneCount)
	ctx.Step(`^the user receives a record list`, s.theUserReceivesARecordList)
	ctx.Step(`^the record list contains the record count`, s.theRecordListContainsTheRecordCount)

	// Record count verification
	ctx.Step(`^the DNS record count is (\d+)$`, s.theDNSRecordCountIs)

	// Bulk delete verification (bulk create validation is handled by count check)
	ctx.Step(`^all DNS records are deleted successfully`, s.allDNSRecordsAreDeletedSuccessfully)

	// General verification steps
	ctx.Step(`^the DNS zone is created successfully`, s.theDNSZoneIsCreatedSuccessfully)
}

// theDNSZoneIsRetrievedSuccessfully verifies that a DNS zone can be retrieved
func (s *DNSCommonSteps) theDNSZoneIsRetrievedSuccessfully(ctx context.Context) (context.Context, error) {
	zoneID, ok := helpers.GetDNSZoneID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no DNS zone ID found in context")
	}

	zone, err := helpers.GetDNSZone(ctx, zoneID)
	if err != nil {
		return ctx, fmt.Errorf("failed to retrieve DNS zone: %w", err)
	}

	if zone == nil {
		return ctx, fmt.Errorf("DNS zone is nil")
	}

	return ctx, nil
}

// theZoneDomainIsVerified verifies that a zone domain is stored in context
func (s *DNSCommonSteps) theZoneDomainIsVerified(ctx context.Context) (context.Context, error) {
	domain, ok := helpers.GetDNSZoneDomain(ctx)
	if !ok {
		return ctx, fmt.Errorf("no DNS zone domain found in context")
	}

	if domain == "" {
		return ctx, fmt.Errorf("DNS zone domain is empty")
	}

	return ctx, nil
}

// theZoneCannotBeRetrieved verifies that a DNS zone cannot be retrieved after deletion
func (s *DNSCommonSteps) theZoneCannotBeRetrieved(ctx context.Context) (context.Context, error) {
	zoneID, ok := helpers.GetDNSZoneID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no DNS zone ID found in context")
	}

	zone, err := helpers.GetDNSZone(ctx, zoneID)
	if err == nil && zone != nil {
		return ctx, fmt.Errorf("DNS zone %s can still be retrieved after deletion", zoneID)
	}

	return ctx, nil
}

// theDNSRecordIsRetrievedSuccessfully verifies that a DNS record can be retrieved
func (s *DNSCommonSteps) theDNSRecordIsRetrievedSuccessfully(ctx context.Context) (context.Context, error) {
	zoneID, ok := helpers.GetDNSZoneID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no DNS zone ID found in context")
	}

	recordFQDN, ok := helpers.GetContextValue[string](ctx, helpers.DNSRecordFQDNKey)
	if !ok {
		return ctx, fmt.Errorf("no DNS record FQDN found in context")
	}

	recordType, ok := helpers.GetDNSRecordType(ctx)
	if !ok {
		return ctx, fmt.Errorf("no DNS record type found in context")
	}

	record, err := helpers.GetDNSRecord(ctx, zoneID, recordFQDN, recordType)
	if err != nil {
		return ctx, fmt.Errorf("failed to retrieve DNS record: %w", err)
	}

	if record == nil {
		return ctx, fmt.Errorf("DNS record is nil")
	}

	return ctx, nil
}

// theRecordNameIs verifies that the record name matches the expected value
func (s *DNSCommonSteps) theRecordNameIs(ctx context.Context, expectedName string) (context.Context, error) {
	name, ok := helpers.GetDNSRecordName(ctx)
	if !ok {
		return ctx, fmt.Errorf("no DNS record name found in context")
	}

	if name != expectedName {
		return ctx, fmt.Errorf("DNS record name is %s, expected %s", name, expectedName)
	}

	return ctx, nil
}

// theRecordTypeIs verifies that the record type matches the expected value
func (s *DNSCommonSteps) theRecordTypeIs(ctx context.Context, expectedType string) (context.Context, error) {
	recordType, ok := helpers.GetDNSRecordType(ctx)
	if !ok {
		return ctx, fmt.Errorf("no DNS record type found in context")
	}

	if recordType != expectedType {
		return ctx, fmt.Errorf("DNS record type is %s, expected %s", recordType, expectedType)
	}

	return ctx, nil
}

// theRecordValueIs verifies that the record value matches the expected value
func (s *DNSCommonSteps) theRecordValueIs(ctx context.Context, expectedValue string) (context.Context, error) {
	recordValue, ok := helpers.GetContextValue[string](ctx, helpers.DNSRecordValueKey)
	if !ok {
		return ctx, fmt.Errorf("no DNS record value found in context")
	}

	if recordValue != expectedValue {
		return ctx, fmt.Errorf("DNS record value is %s, expected %s", recordValue, expectedValue)
	}

	return ctx, nil
}

// theDNSRecordIsCreatedSuccessfully verifies that a DNS record creation succeeds
func (s *DNSCommonSteps) theDNSRecordIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	recordFQDN, ok := helpers.GetContextValue[string](ctx, helpers.DNSRecordFQDNKey)
	if !ok {
		return ctx, fmt.Errorf("no DNS record FQDN found in context")
	}

	recordType, ok := helpers.GetDNSRecordType(ctx)
	if !ok {
		return ctx, fmt.Errorf("no DNS record type found in context")
	}

	if recordFQDN == "" || recordType == "" {
		return ctx, fmt.Errorf("DNS record FQDN or type is empty")
	}

	return ctx, nil
}

// theDNSRecordIsUpdatedSuccessfully verifies that a DNS record update succeeds
func (s *DNSCommonSteps) theDNSRecordIsUpdatedSuccessfully(ctx context.Context) (context.Context, error) {
	recordFQDN, ok := helpers.GetContextValue[string](ctx, helpers.DNSRecordFQDNKey)
	if !ok {
		return ctx, fmt.Errorf("no DNS record FQDN found in context")
	}

	recordType, ok := helpers.GetDNSRecordType(ctx)
	if !ok {
		return ctx, fmt.Errorf("no DNS record type found in context")
	}

	if recordFQDN == "" || recordType == "" {
		return ctx, fmt.Errorf("DNS record FQDN or type is empty after update")
	}

	return ctx, nil
}

// theRecordCannotBeRetrieved verifies that a DNS record cannot be retrieved after deletion
func (s *DNSCommonSteps) theRecordCannotBeRetrieved(ctx context.Context) (context.Context, error) {
	zoneID, ok := helpers.GetDNSZoneID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no DNS zone ID found in context")
	}

	recordFQDN, ok := helpers.GetContextValue[string](ctx, helpers.DNSRecordFQDNKey)
	if !ok {
		return ctx, fmt.Errorf("no DNS record FQDN found in context")
	}

	recordType, ok := helpers.GetDNSRecordType(ctx)
	if !ok {
		return ctx, fmt.Errorf("no DNS record type found in context")
	}

	record, err := helpers.GetDNSRecord(ctx, zoneID, recordFQDN, recordType)
	if err == nil && record != nil {
		return ctx, fmt.Errorf("DNS record %s %s can still be retrieved after deletion", recordFQDN, recordType)
	}

	return ctx, nil
}

// theUserReceivesAZoneList verifies that a zone list can be returned
func (s *DNSCommonSteps) theUserReceivesAZoneList(ctx context.Context) (context.Context, error) {
	zones, err := helpers.ListDNSZones(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list DNS zones: %w", err)
	}

	if zones == nil {
		return ctx, fmt.Errorf("DNS zone list is nil")
	}

	return ctx, nil
}

// theZoneListContainsTheZoneCount verifies that the zone list has an expected count
// This expects the count to have been stored in context by the list operation
func (s *DNSCommonSteps) theZoneListContainsTheZoneCount(ctx context.Context) (context.Context, error) {
	count, ok := helpers.GetContextValue[int](ctx, helpers.DNSZoneListCountKey)
	if !ok {
		return ctx, fmt.Errorf("zone count not found in context (key: %s)", helpers.DNSZoneListCountKey)
	}
	if count < 0 {
		return ctx, fmt.Errorf("zone count is negative: %d", count)
	}
	return ctx, nil
}

// theUserReceivesARecordList verifies that a record list can be returned
func (s *DNSCommonSteps) theUserReceivesARecordList(ctx context.Context) (context.Context, error) {
	zoneID, ok := helpers.GetDNSZoneID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no DNS zone ID found in context")
	}

	records, err := helpers.ListDNSRecords(ctx, zoneID)
	if err != nil {
		return ctx, fmt.Errorf("failed to list DNS records: %w", err)
	}

	if records == nil {
		return ctx, fmt.Errorf("DNS record list is nil")
	}

	return ctx, nil
}

// theRecordListContainsTheRecordCount verifies that the record list has an expected count
// This expects the count to have been stored in context by the list operation
func (s *DNSCommonSteps) theRecordListContainsTheRecordCount(ctx context.Context) (context.Context, error) {
	count, ok := helpers.GetContextValue[int](ctx, helpers.DNSRecordListCountKey)
	if !ok {
		return ctx, fmt.Errorf("record count not found in context (key: %s)", helpers.DNSRecordListCountKey)
	}
	if count < 0 {
		return ctx, fmt.Errorf("record count is negative: %d", count)
	}
	return ctx, nil
}

// theDNSRecordCountIs verifies the DNS record count matches expected
// Uses the count stored in context from list operation instead of making another API call
func (s *DNSCommonSteps) theDNSRecordCountIs(ctx context.Context, expectedCount int) (context.Context, error) {
	count, ok := helpers.GetContextValue[int](ctx, helpers.DNSRecordListCountKey)
	if !ok {
		return ctx, fmt.Errorf("record count not found in context (key: %s)", helpers.DNSRecordListCountKey)
	}

	if count != expectedCount {
		return ctx, fmt.Errorf("DNS record count is %d, expected %d", count, expectedCount)
	}

	return ctx, nil
}


// allDNSRecordsAreDeletedSuccessfully verifies that all DNS records were deleted
// by confirming the zone's record list is empty
func (s *DNSCommonSteps) allDNSRecordsAreDeletedSuccessfully(ctx context.Context) (context.Context, error) {
	zoneID, ok := helpers.GetDNSZoneID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no DNS zone ID found in context")
	}

	records, err := helpers.ListDNSRecords(ctx, zoneID)
	if err != nil {
		return ctx, fmt.Errorf("failed to list DNS records for verification: %w", err)
	}

	if len(records) != 0 {
		return ctx, fmt.Errorf("DNS zone still has %d records after bulk delete, expected 0", len(records))
	}

	return ctx, nil
}


// theDNSZoneIsCreatedSuccessfully verifies that a DNS zone was created
func (s *DNSCommonSteps) theDNSZoneIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	zoneID, ok := helpers.GetDNSZoneID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no DNS zone ID found in context")
	}

	if zoneID == "" {
		return ctx, fmt.Errorf("DNS zone ID is empty")
	}

	// Verify the zone can be retrieved
	zone, err := helpers.GetDNSZone(ctx, zoneID)
	if err != nil {
		return ctx, fmt.Errorf("failed to verify DNS zone creation: %w", err)
	}

	if zone == nil {
		return ctx, fmt.Errorf("DNS zone is nil after creation")
	}

	return ctx, nil
}
