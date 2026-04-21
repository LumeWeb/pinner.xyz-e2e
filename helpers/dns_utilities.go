package helpers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	ipfs_sdk "go.lumeweb.com/ipfs-sdk"
)

// CreateDNSZone creates a DNS zone and tracks it for cleanup
// Returns the updated context, zone response, or an error
func CreateDNSZone(ctx context.Context, domain string, nameservers []string) (context.Context, *ipfs_sdk.ZoneResponse, error) {
	dnsService, err := RequireDNSService(ctx)
	if err != nil {
		return ctx, nil, err
	}

	zone, err := dnsService.CreateZone(ctx, domain, nameservers)
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to create DNS zone: %w", err)
	}

	// Track zone for cleanup and context
	// Use database integer ID for API calls
	zoneDBID := strconv.Itoa(zone.Id)

	ctx = AddDNSZoneCleanup(ctx, zoneDBID)
	ctx = SetDNSZoneID(ctx, zoneDBID)
	ctx = SetDNSZoneDomain(ctx, zone.Domain)
	return ctx, zone, nil
}

// DeleteDNSZone deletes a DNS zone by ID
func DeleteDNSZone(ctx context.Context, zoneID string) error {
	dnsService, err := RequireDNSService(ctx)
	if err != nil {
		return err
	}

	err = dnsService.DeleteZone(ctx, zoneID)
	if err != nil {
		return fmt.Errorf("failed to delete DNS zone %s: %w", zoneID, err)
	}

	return nil
}

// ListDNSZones retrieves all DNS zones for the authenticated user
func ListDNSZones(ctx context.Context) ([]ipfs_sdk.ZoneListResponse, error) {
	dnsService, err := RequireDNSService(ctx)
	if err != nil {
		return nil, err
	}

	zones, err := dnsService.ListZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS zones: %w", err)
	}

	return zones, nil
}

// GetDNSZone retrieves a specific DNS zone by ID
func GetDNSZone(ctx context.Context, zoneID string) (*ipfs_sdk.ZoneResponse, error) {
	dnsService, err := RequireDNSService(ctx)
	if err != nil {
		return nil, err
	}

	zone, err := dnsService.GetZone(ctx, zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS zone %s: %w", zoneID, err)
	}

	return zone, nil
}

// CreateDNSRecord creates a DNS record in a zone
// Returns the updated context, record response, or an error
func CreateDNSRecord(ctx context.Context, zoneID string, record ipfs_sdk.RecordRequest) (context.Context, *ipfs_sdk.RecordResponse, error) {
	dnsService, err := RequireDNSService(ctx)
	if err != nil {
		return ctx, nil, err
	}

	createdRecord, err := dnsService.CreateRecord(ctx, zoneID, record)
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to create DNS record: %w", err)
	}
	// DNS records are identified by name+type combination, not by integer ID
	// Store the server-returned FQDN for API operations (get/delete/update)
	newCtx := SetContextValue(ctx, DNSRecordFQDNKey, createdRecord.Name)
	newCtx = SetDNSRecordType(newCtx, createdRecord.Type)
	newCtx = SetContextValue(newCtx, DNSRecordValueKey, createdRecord.Content)

	return newCtx, createdRecord, nil
}

// DeleteDNSRecord deletes a DNS record by zone ID, name, and type
func DeleteDNSRecord(ctx context.Context, zoneID string, name string, recordType string) error {
	dnsService, err := RequireDNSService(ctx)
	if err != nil {
		return err
	}

	err = dnsService.DeleteRecord(ctx, zoneID, name, recordType)
	if err != nil {
		return fmt.Errorf("failed to delete DNS record %s %s: %w", name, recordType, err)
	}

	return nil
}

// ListDNSRecords retrieves all DNS records for a zone
func ListDNSRecords(ctx context.Context, zoneID string) ([]ipfs_sdk.RecordResponse, error) {
	dnsService, err := RequireDNSService(ctx)
	if err != nil {
		return nil, err
	}

	records, err := dnsService.ListRecords(ctx, zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}

	return records, nil
}

// GetDNSRecord retrieves a specific DNS record
func GetDNSRecord(ctx context.Context, zoneID string, name string, recordType string) (*ipfs_sdk.RecordResponse, error) {
	dnsService, err := RequireDNSService(ctx)
	if err != nil {
		return nil, err
	}

	record, err := dnsService.GetRecord(ctx, zoneID, name, recordType)
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS record %s %s: %w", name, recordType, err)
	}

	return record, nil
}

// UpdateDNSRecord updates a DNS record
// Returns the updated context, record response, or an error
func UpdateDNSRecord(ctx context.Context, zoneID string, name string, recordType string, record ipfs_sdk.RecordRequest) (context.Context, *ipfs_sdk.RecordResponse, error) {
	dnsService, err := RequireDNSService(ctx)
	if err != nil {
		return ctx, nil, err
	}

	updatedRecord, err := dnsService.UpdateRecord(ctx, zoneID, name, recordType, record)
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to update DNS record %s %s: %w", name, recordType, err)
	}

	// Store updated record details in context
	// DNS records are identified by name+type combination, not by integer ID
	ctx = SetContextValue(ctx, DNSRecordFQDNKey, updatedRecord.Name)
	ctx = SetDNSRecordType(ctx, updatedRecord.Type)
	ctx = SetContextValue(ctx, DNSRecordValueKey, updatedRecord.Content)

	return ctx, updatedRecord, nil
}

// BulkCreateDNSRecords creates multiple DNS records in a zone
func BulkCreateDNSRecords(ctx context.Context, zoneID string, records []ipfs_sdk.RecordRequest) ([]ipfs_sdk.RecordResponse, error) {
	dnsService, err := RequireDNSService(ctx)
	if err != nil {
		return nil, err
	}

	createdRecords, err := dnsService.BulkCreateRecords(ctx, zoneID, records)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk create DNS records: %w", err)
	}

	return createdRecords, nil
}

// BulkDeleteDNSRecords deletes multiple DNS records
func BulkDeleteDNSRecords(ctx context.Context, zoneID string, identifiers []ipfs_sdk.RecordIdentifier) ([]ipfs_sdk.RecordResult, error) {
	dnsService, err := RequireDNSService(ctx)
	if err != nil {
		return nil, err
	}

	results, err := dnsService.BulkDeleteRecords(ctx, zoneID, identifiers)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk delete DNS records: %w", err)
	}

	return results, nil
}

// CleanupDNSZones removes all DNS zones tracked in the cleanup list
// Returns an error only for actual failures, not "zone not found" (which indicates successful prior deletion)
func CleanupDNSZones(ctx context.Context) error {
	zoneIDs := GetDNSZonesCleanup(ctx)
	if len(zoneIDs) == 0 {
		return nil
	}

	dnsService, err := GetDNSService(ctx)
	if err != nil {
		return err
	}

	for _, zoneID := range zoneIDs {
		err := dnsService.DeleteZone(ctx, zoneID)
		// "zone not found" is not an error - it means the zone was already deleted (successful cleanup)
		if err != nil && err.Error() != "zone not found" {
			return fmt.Errorf("failed to delete DNS zone %s: %w", zoneID, err)
		}
	}

	return nil
}

// GenerateTestDomain generates a unique test domain for DNS testing
func GenerateTestDomain() string {
	return fmt.Sprintf("test-%s.example.com", uuid.New().String()[:8])
}

// GenerateTestRecordName generates a unique test record name for DNS testing
func GenerateTestRecordName() string {
	return fmt.Sprintf("test-%s", uuid.New().String()[:8])
}
