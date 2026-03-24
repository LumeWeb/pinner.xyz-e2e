package helpers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.lumeweb.com/ipfs-sdk"
)

const (
	// DefaultTestDomain is the standard test domain used in feature files
	// Using a constant makes it easier to identify and change later
	DefaultTestWebsiteDomain = "test-website.example.com"
	// DefaultTargetType is the default target type for websites (ipfs or ipns)
	DefaultTargetType = "ipfs"
)

// WebsiteContext contains context values needed for website operations
type WebsiteContext struct {
	ID         int
	Domain     string
	TargetHash string
	TargetType string
}

// RequireWebsiteContext extracts website ID, domain, target hash, and target type from context
// This consolidates repeated context extraction patterns used across website step implementations
func RequireWebsiteContext(ctx context.Context) (*WebsiteContext, error) {
	id, ok := GetWebsiteID(ctx)
	if !ok {
		return nil, fmt.Errorf("no website ID found in context")
	}

	domain, ok := GetWebsiteDomain(ctx)
	if !ok {
		return nil, fmt.Errorf("no website domain found in context")
	}

	targetHash, ok := GetWebsiteTargetHash(ctx)
	if !ok {
		return nil, fmt.Errorf("no website target hash found in context")
	}

	targetType, ok := GetWebsiteTargetType(ctx)
	if !ok {
		return nil, fmt.Errorf("no website target type found in context")
	}

	return &WebsiteContext{
		ID:         id,
		Domain:     domain,
		TargetHash: targetHash,
		TargetType: targetType,
	}, nil
}

// RequireWebsiteID retrieves the website ID from context or returns an error
// This DRYs up the common "get and check" pattern used across website step implementations
func RequireWebsiteID(ctx context.Context) (int, error) {
	id, ok := GetWebsiteID(ctx)
	if !ok {
		return 0, fmt.Errorf("no website ID found in context")
	}
	return id, nil
}

// RequireWebsiteDomain retrieves the website domain from context or returns an error
func RequireWebsiteDomain(ctx context.Context) (string, error) {
	domain, ok := GetWebsiteDomain(ctx)
	if !ok {
		return "", fmt.Errorf("no website domain found in context")
	}
	return domain, nil
}

// GenerateTestWebsiteDomain generates a unique test domain for website testing
func GenerateTestWebsiteDomain() string {
	return fmt.Sprintf("test-website-%s.example.com", uuid.New().String()[:8])
}

// BuildWebsiteRequest creates a WebsiteRequest with the given parameters
// This DRYs up the common request building pattern used across website step implementations
func BuildWebsiteRequest(domain, targetHash, targetType string, dnsHostingEnabled bool) ipfs.WebsiteRequest {
	return ipfs.WebsiteRequest{
		Domain:            domain,
		TargetHash:       targetHash,
		TargetType:       targetType,
		DnsHostingEnabled: &dnsHostingEnabled,
	}
}

// StoreWebsiteID stores a website ID in context
func StoreWebsiteID(ctx context.Context, id int) context.Context {
	return SetContextValue(ctx, WebsiteIDKey, id)
}

// GetWebsiteID retrieves the website ID from context
func GetWebsiteID(ctx context.Context) (int, bool) {
	return GetContextValue[int](ctx, WebsiteIDKey)
}

// AddWebsiteIDCleanup adds a website ID to the cleanup list
func AddWebsiteIDCleanup(ctx context.Context, id int) context.Context {
	return AddToCleanupList(ctx, WebsitesCleanupKey, id)
}

// GetWebsitesCleanup retrieves the website IDs cleanup list
func GetWebsitesCleanup(ctx context.Context) []int {
	return GetCleanupList[int](ctx, WebsitesCleanupKey)
}

// SetWebsiteDomain stores a website domain in context
func SetWebsiteDomain(ctx context.Context, domain string) context.Context {
	return SetContextValue(ctx, WebsiteDomainKey, domain)
}

// GetWebsiteDomain retrieves the website domain from context
func GetWebsiteDomain(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, WebsiteDomainKey)
}

// SetWebsiteTargetHash stores a target hash in context
func SetWebsiteTargetHash(ctx context.Context, targetHash string) context.Context {
	return SetContextValue(ctx, WebsiteTargetHashKey, targetHash)
}

// GetWebsiteTargetHash retrieves the target hash from context
func GetWebsiteTargetHash(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, WebsiteTargetHashKey)
}

// SetWebsiteTargetType stores a target type in context
func SetWebsiteTargetType(ctx context.Context, targetType string) context.Context {
	return SetContextValue(ctx, WebsiteTargetTypeKey, targetType)
}

// GetWebsiteTargetType retrieves the target type from context
func GetWebsiteTargetType(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, WebsiteTargetTypeKey)
}

// SetWebsiteDnsHostingEnabled stores DNS hosting enabled status in context
func SetWebsiteDnsHostingEnabled(ctx context.Context, enabled bool) context.Context {
	return SetContextValue(ctx, WebsiteDnsHostingKey, enabled)
}

// GetWebsiteDnsHostingEnabled retrieves DNS hosting enabled status from context
func GetWebsiteDnsHostingEnabled(ctx context.Context) (bool, bool) {
	return GetContextValue[bool](ctx, WebsiteDnsHostingKey)
}

// SetWebsiteSslStatus stores SSL status in context
func SetWebsiteSslStatus(ctx context.Context, status string) context.Context {
	return SetContextValue(ctx, WebsiteSslStatusKey, status)
}

// GetWebsiteSslStatus retrieves SSL status from context
func GetWebsiteSslStatus(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, WebsiteSslStatusKey)
}

// SetWebsiteStatus stores website status in context
func SetWebsiteStatus(ctx context.Context, status string) context.Context {
	return SetContextValue(ctx, WebsiteStatusKey, status)
}

// GetWebsiteStatus retrieves website status from context
func GetWebsiteStatus(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, WebsiteStatusKey)
}

// GetWebsiteFixturePath returns the path to the sample website fixture
func GetWebsiteFixturePath() string {
	return "./internal/fixtures/sample-website"
}

// SetWebsiteDevDnsZoneID stores the development DNS server zone ID for a website in context
func SetWebsiteDevDnsZoneID(ctx context.Context, zoneID string) context.Context {
	return SetContextValue(ctx, WebsiteDnsZoneIDKey, zoneID)
}

// GetWebsiteDevDnsZoneID retrieves the development DNS server zone ID for a website from context
func GetWebsiteDevDnsZoneID(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, WebsiteDnsZoneIDKey)
}
