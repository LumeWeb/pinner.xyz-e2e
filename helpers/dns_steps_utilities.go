package helpers

import (
	"context"
	"fmt"

	ipfs_sdk "go.lumeweb.com/ipfs-sdk"
)

// DefaultTTL is the default TTL for DNS records in tests
const DefaultTTL = 3600

// DNSRecordContext contains context values needed for DNS record operations
type DNSRecordContext struct {
	ZoneID    string
	RecordFQDN string
	RecordType string
}

// RequireDNSRecordContext extracts zone ID, record FQDN, and record type from context
// This consolidates repeated context extraction patterns used across DNS step implementations
func RequireDNSRecordContext(ctx context.Context) (*DNSRecordContext, error) {
	zoneID, ok := GetDNSZoneID(ctx)
	if !ok {
		return nil, fmt.Errorf("no DNS zone ID found in context")
	}

	recordFQDN, ok := GetContextValue[string](ctx, DNSRecordFQDNKey)
	if !ok {
		return nil, fmt.Errorf("no DNS record FQDN found in context")
	}

	recordType, ok := GetDNSRecordType(ctx)
	if !ok {
		return nil, fmt.Errorf("no DNS record type found in context")
	}

	return &DNSRecordContext{
		ZoneID:    zoneID,
		RecordFQDN: recordFQDN,
		RecordType: recordType,
	}, nil
}

// pointer returns a pointer to the given value
func pointer[T any](v T) *T {
	return &v
}

// RequireDNSZoneID retrieves the DNS zone ID from context or returns an error
// This DRYs up the common "get and check" pattern used across DNS step implementations
func RequireDNSZoneID(ctx context.Context) (string, error) {
	zoneID, ok := GetDNSZoneID(ctx)
	if !ok {
		return "", fmt.Errorf("no DNS zone ID found in context")
	}
	return zoneID, nil
}

// RequireDNSRecordType retrieves the DNS record type from context or returns an error
// This DRYs up the common "get and check" pattern used across DNS step implementations
func RequireDNSRecordType(ctx context.Context) (string, error) {
	recordType, ok := GetDNSRecordType(ctx)
	if !ok {
		return "", fmt.Errorf("no DNS record type found in context")
	}
	return recordType, nil
}

// BuildRecordRequest creates a DNS RecordRequest with the given parameters and default TTL
// This DRYs up the common record building pattern used across DNS step implementations
func BuildRecordRequest(recordType, name, content string) ipfs_sdk.RecordRequest {
	return ipfs_sdk.RecordRequest{
		Type:    recordType,
		Name:    name,
		Content: content,
		Ttl:     pointer(DefaultTTL),
	}
}

// StoreDNSRecordValue stores a DNS record value in context for verification
// This DRYs up the common pattern of storing record values after create/update operations
func StoreDNSRecordValue(ctx context.Context, value string) context.Context {
	return SetContextValue(ctx, DNSRecordValueKey, value)
}

// RequireDNSZone retrieves the zone ID and returns an error if not available
// This consolidates zone ID validation logic used across multiple steps
func RequireDNSZone(ctx context.Context) error {
	_, err := RequireDNSZoneID(ctx)
	return err
}
