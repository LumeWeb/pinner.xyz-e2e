package helpers

import (
	"context"
	"fmt"
	"os"

	"go.lumeweb.com/ipfs-sdk"
	account "go.lumeweb.com/portal-sdk"
)

// GenerateTestFile creates a temporary file with the specified name and content.
// Returns the file path for use in tests. The caller is responsible for cleaning up.
func GenerateTestFile(name string, content string) (string, error) {
	// Create temporary file with specified name
	tmpFile, err := os.CreateTemp("", name)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	// Write content to file
	if _, err := tmpFile.WriteString(content); err != nil {
		os.Remove(tmpFile.Name())

		return "", fmt.Errorf("failed to write content to temp file: %w", err)
	}

	return tmpFile.Name(), nil
}

// GetIPFSClient creates an IPFS SDK client using JWT token from context and portal endpoint
// The client includes a download rate limiter that checks quota before allowing downloads.
// Returns an error if JWT token is not available or if the endpoint is not configured
func GetIPFSClient(ctx context.Context) (*ipfs.Client, error) {
	token, ok := GetJWTToken(ctx)
	if !ok {
		return nil, fmt.Errorf("no JWT token found in context")
	}

	if token == "" {
		return nil, fmt.Errorf("JWT token is empty")
	}

	ipfsEndpoint := GetIPFSEndpoint()
	if ipfsEndpoint == "" {
		return nil, fmt.Errorf("ipfs endpoint not configured")
	}

	// Get portal account client for rate limiter
	api, err := RequireAuthenticatedClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated client: %w", err)
	}

	// Create download rate limiter that checks quota before allowing downloads
	// Blocks downloads when quota usage reaches 98% to provide margin before hard limit
	rateLimiter := account.CreateDownloadPercentLimitedRateLimiter(api, 98)

	// Get gateway secret for internal API authentication
	gatewaySecret, err := GetGatewaySecret(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway secret: %w", err)
	}

	// Use host override for vhost routing (similar to account API),
	// gateway secret for internal API authentication,
	// and download rate limiter for quota enforcement
	client, err := ipfs.NewClient(ipfsEndpoint, token,
		ipfs.WithHostOverride(GetIPFSHost(), GetPortalTarget()),
		ipfs.WithGatewaySecret(gatewaySecret),
		ipfs.WithDownloadOption(ipfs.WithDownloadRateLimiter(rateLimiter)))
	if err != nil {
		return nil, fmt.Errorf("failed to create IPFS client: %w", err)
	}

	return client, nil
}
