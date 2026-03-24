package helpers

import (
	"context"
	"fmt"
	"os"

	"go.lumeweb.com/ipfs-sdk"
)

// GetWebsiteService retrieves the website service from an IPFS SDK client
// This provides access to website management operations
func GetWebsiteService(ctx context.Context) (ipfs.WebsitesService, error) {
	client, err := GetIPFSClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPFS client: %w", err)
	}

	return client.Websites(), nil
}

// RequireWebsiteService retrieves the website service from context and returns an error if not available
func RequireWebsiteService(ctx context.Context) (ipfs.WebsitesService, error) {
	return GetWebsiteService(ctx)
}

// EnsureWebsiteClientAccess verifies that website client can be created and used
// Returns an error if client creation fails or if the service is not available
func EnsureWebsiteClientAccess(ctx context.Context) error {
	if _, err := GetIPFSClient(ctx); err != nil {
		return fmt.Errorf("website client access check failed: %w", err)
	}

	if _, err := GetWebsiteService(ctx); err != nil {
		return fmt.Errorf("website service access check failed: %w", err)
	}

	return nil
}

// GetGatewaySecret retrieves the gateway secret from environment or configuration
// This is required for authenticating internal API calls from edge nodes
func GetGatewaySecret(ctx context.Context) (string, error) {
	// Get gateway secret from environment variable PORTAL__PLUGIN__IPFS__API__GATEWAY_SECRET
	secret := os.Getenv("PORTAL__PLUGIN__IPFS__API__GATEWAY_SECRET")
	if secret == "" {
		return "", fmt.Errorf("PORTAL__PLUGIN__IPFS__API__GATEWAY_SECRET environment variable not set")
	}
	return secret, nil
}
