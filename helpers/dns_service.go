package helpers

import (
	"context"
	"fmt"

	"go.lumeweb.com/ipfs-sdk"
)

// GetDNSService retrieves the DNS service from an IPFS SDK client
// This provides access to DNS zone and record management operations
func GetDNSService(ctx context.Context) (ipfs.DNSService, error) {
	client, err := GetIPFSClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPFS client: %w", err)
	}

	return client.DNS(), nil
}

// RequireDNSService retrieves the DNS service from context and returns an error if not available
func RequireDNSService(ctx context.Context) (ipfs.DNSService, error) {
	return GetDNSService(ctx)
}

// EnsureDNSClientAccess verifies that DNS client can be created and used
// Returns an error if client creation fails or if the service is not available
func EnsureDNSClientAccess(ctx context.Context) error {
	if _, err := GetIPFSClient(ctx); err != nil {
		return fmt.Errorf("DNS client access check failed: %w", err)
	}

	if _, err := GetDNSService(ctx); err != nil {
		return fmt.Errorf("DNS service access check failed: %w", err)
	}

	return nil
}
