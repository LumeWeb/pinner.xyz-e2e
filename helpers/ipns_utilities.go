package helpers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	ipfs_sdk "go.lumeweb.com/ipfs-sdk"
)

// IPNSKeyCleanup represents an IPNS key that needs to be cleaned up
type IPNSKeyCleanup struct {
	KeyID int
	Name  string
}

// CreateIPNSKey creates an IPNS key and tracks it for cleanup
// Returns the updated context with context, key response, or an error
func CreateIPNSKey(ctx context.Context, name string) (context.Context, *ipfs_sdk.IPNSKeyResponse, error) {
	ipnsService, err := RequireIPNSService(ctx)
	if err != nil {
		return ctx, nil, err
	}

	key, err := ipnsService.CreateKey(ctx, name)
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to create IPNS key: %w", err)
	}

	// Track key for cleanup
	keyIDStr := strconv.Itoa(key.Id)
	ctx = AddIPNSKeyCleanup(ctx, keyIDStr)

	// Store key details in context for verification steps
	ctx = SetIPNSKeyID(ctx, key.Id)
	ctx = SetIPNSKeyName(ctx, key.Name)

	if key.IpnsName != "" {
		ctx = SetIPNSIPNSName(ctx, key.IpnsName)
	}

	if key.PeerId != "" {
		ctx = SetIPNSPeerID(ctx, key.PeerId)
	}

	return ctx, key, nil
}

// DeleteIPNSKey deletes an IPNS key by ID and removes it from cleanup list
// Returns the updated context or an error
func DeleteIPNSKey(ctx context.Context, keyID string) (context.Context, error) {
	ipnsService, err := RequireIPNSService(ctx)
	if err != nil {
		return ctx, err
	}

	err = ipnsService.DeleteKey(ctx, keyID)
	if err != nil {
		return ctx, fmt.Errorf("failed to delete IPNS key %s: %w", keyID, err)
	}

	// Remove from cleanup list to prevent duplicate deletion
	ctx = RemoveIPNSKeyCleanup(ctx, keyID)

	return ctx, nil
}

// CleanupIPNSKeys removes all IPNS keys tracked in the cleanup list
func CleanupIPNSKeys(ctx context.Context) error {
	keyIDs := GetIPNSKeysCleanup(ctx)
	if len(keyIDs) == 0 {
		return nil
	}

	ipnsService, err := GetIPNSService(ctx)
	if err != nil {
		fmt.Printf("Warning: failed to get IPNS service for cleanup: %v\n", err)
		return err
	}

	var lastErr error
	for _, keyID := range keyIDs {
		if err := ipnsService.DeleteKey(ctx, keyID); err != nil {
			fmt.Printf("Warning: failed to delete IPNS key %s: %v\n", keyID, err)
			lastErr = err
		}
	}

	return lastErr
}

// ListIPNSKeys retrieves all IPNS keys for the authenticated user
func ListIPNSKeys(ctx context.Context) ([]ipfs_sdk.IPNSKeyResponse, error) {
	ipnsService, err := RequireIPNSService(ctx)
	if err != nil {
		return nil, err
	}

	keys, err := ipnsService.ListKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list IPNS keys: %w", err)
	}

	return keys, nil
}

// GetIPNSKey retrieves a specific IPNS key by ID
func GetIPNSKey(ctx context.Context, keyID string) (*ipfs_sdk.IPNSKeyResponse, error) {
	ipnsService, err := RequireIPNSService(ctx)
	if err != nil {
		return nil, err
	}

	key, err := ipnsService.GetKey(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPNS key %s: %w", keyID, err)
	}

	return key, nil
}

// PublishToIPNS publishes a CID to an IPNS key
// Returns the publish response or an error
func PublishToIPNS(ctx context.Context, keyID int, cid string) (*ipfs_sdk.IPNSPublishResponse, error) {
	ipnsService, err := RequireIPNSService(ctx)
	if err != nil {
		return nil, err
	}

	publish, err := ipnsService.Publish(ctx, keyID, cid)
	if err != nil {
		return nil, fmt.Errorf("failed to publish to IPNS: %w", err)
	}

	return publish, nil
}

// RepublishIPNS republishes all IPNS entries
func RepublishIPNS(ctx context.Context) error {
	ipnsService, err := RequireIPNSService(ctx)
	if err != nil {
		return err
	}

	err = ipnsService.Republish(ctx)
	if err != nil {
		return fmt.Errorf("failed to republish IPNS: %w", err)
	}

	return nil
}

// ResolveIPNSName resolves an IPNS name to its current CID
func ResolveIPNSName(ctx context.Context, name string) (*ipfs_sdk.IPNSResolveResponse, error) {
	ipnsService, err := RequireIPNSService(ctx)
	if err != nil {
		return nil, err
	}

	resolve, err := ipnsService.Resolve(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve IPNS name %s: %w", name, err)
	}

	return resolve, nil
}

// WaitForIPNSPublish waits for an IPNS publish operation to complete
// WaitForIPNSPublish waits for an IPNS publish operation to complete
// Uses the SDK's WaitForIPNSResolution to poll until the record resolves to expected CID
func WaitForIPNSPublish(ctx context.Context, keyID int, expectedCID string) error {

	ipnsService, err := RequireIPNSService(ctx)
	if err != nil {
		return err
	}

	// Get the IPNS key to retrieve the peer ID/name
	keyIDStr := fmt.Sprintf("%d", keyID)
	key, err := ipnsService.GetKey(ctx, keyIDStr)
	if err != nil {
		return fmt.Errorf("failed to get IPNS key %d: %w", keyID, err)
	}

	// Wait for the IPNS record to resolve to the expected CID
	_, err = ipnsService.WaitForIPNSResolution(ctx, key.PeerId, expectedCID)

	if err != nil {
		return fmt.Errorf("failed waiting for IPNS resolution (key ID: %d, expected CID: %s): %w", keyID, expectedCID, err)
	}

	return nil
}

// RequireIPNSKeyID retrieves IPNS key ID from context and returns an error if not available
func RequireIPNSKeyID(ctx context.Context, contextDesc string) (int, error) {
	keyID, ok := GetIPNSKeyID(ctx)
	if !ok {
		return 0, fmt.Errorf("no IPNS key ID found in context (%s)", contextDesc)
	}

	if keyID == 0 {
		return 0, fmt.Errorf("IPNS key ID is 0 which is invalid (%s)", contextDesc)
	}

	return keyID, nil
}

// ResolveIPNSNameViaPortal resolves an IPNS name and validates the result
// Returns the resolved CID (without /ipfs/ prefix) or an error
func ResolveIPNSNameViaPortal(ctx context.Context, ipnsName string) (string, error) {
	resolve, err := ResolveIPNSName(ctx, ipnsName)
	if err != nil {
		return "", fmt.Errorf("failed to resolve IPNS name %s: %w", ipnsName, err)
	}

	if resolve == nil || resolve.Value == "" {
		return "", fmt.Errorf("IPNS name %s could not be resolved", ipnsName)
	}

	cid := resolve.Value

	// Strip /ipfs/ prefix if present for consistency
	if strings.HasPrefix(cid, "/ipfs/") {
		cid = cid[6:]
	}

	return cid, nil
}

// StoreKeyInfoInContext stores IPNS key information (ID, name, peer ID) in context
// Uses the key response to populate all relevant context values
func StoreKeyInfoInContext(ctx context.Context, key *ipfs_sdk.IPNSKeyResponse) context.Context {
	if key == nil {
		return ctx
	}

	ctx = SetIPNSKeyID(ctx, key.Id)
	if key.IpnsName != "" {
		ctx = SetIPNSIPNSName(ctx, key.IpnsName)
	}
	if key.PeerId != "" {
		ctx = SetIPNSPeerID(ctx, key.PeerId)
	}

	return ctx
}
