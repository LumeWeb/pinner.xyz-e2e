package helpers

import (
	"context"
	"fmt"

	"go.lumeweb.com/ipfs-sdk"
)

// GetIPNSClient creates an IPFS SDK client for IPNS operations using JWT token from context and portal endpoint
// Returns an error if JWT token is not available or if the endpoint is not configured
func GetIPNSClient(ctx context.Context) (*ipfs.Client, error) {
	return GetIPFSClient(ctx)
}

// GetIPNSService retrieves the IPNS service from an IPFS SDK client
// This provides access to IPNS key management, publishing, and resolution operations
func GetIPNSService(ctx context.Context) (ipfs.IPNSService, error) {
	client, err := GetIPFSClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPFS client: %w", err)
	}

	return client.IPNS(), nil
}

// RequireIPNSService retrieves the IPNS service from context and returns an error if not available
func RequireIPNSService(ctx context.Context) (ipfs.IPNSService, error) {
	return GetIPNSService(ctx)
}

// EnsureIPNSClientAccess verifies that IPNS client can be created and used
// Returns an error if client creation fails or if the service is not available
func EnsureIPNSClientAccess(ctx context.Context) error {
	if _, err := GetIPNSClient(ctx); err != nil {
		return fmt.Errorf("IPNS client access check failed: %w", err)
	}

	if _, err := GetIPNSService(ctx); err != nil {
		return fmt.Errorf("IPNS service access check failed: %w", err)
	}

	return nil
}
