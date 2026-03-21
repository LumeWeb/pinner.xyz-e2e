package steps

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"

	ipfs_sdk "go.lumeweb.com/ipfs-sdk"
)

// IPNSKeyManagementSteps holds step definitions for IPNS key management operations
type IPNSKeyManagementSteps struct{}

// NewIPNSKeyManagementSteps creates a new IPNSKeyManagementSteps instance
func NewIPNSKeyManagementSteps() *IPNSKeyManagementSteps {
	return &IPNSKeyManagementSteps{}
}

// InitializeScenario registers all IPNS key management step definitions with godog
func (s *IPNSKeyManagementSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the user creates an IPNS key named "([^"]*)"$`, s.theUserCreatesAnIPNSKey)
	ctx.Step(`^an IPNS key with name "([^"]*)" is created$`, s.anIPNSKeyWithNameIsCreated)
	ctx.Step(`^the IPNS key has a valid peer ID$`, s.theIPNSKeyHasValidPeerIDFromKey)
	ctx.Step(`^the user creates (\d+) IPNS keys$`, s.theUserCreatesMultipleIPNSKeys)
	ctx.Step(`^the user lists their IPNS keys$`, s.theUserListsTheirIPNSKeys)
	ctx.Step(`^all (\d+) IPNS keys are returned$`, s.allIPNSKeysAreReturned)
	ctx.Step(`^each key has a unique peer ID$`, s.eachKeyHasUniquePeerID)
	ctx.Step(`^the user has (\d+) IPNS keys$`, s.theUserHasMultipleIPNSKeys)
	ctx.Step(`^the user deletes the IPNS key$`, s.theUserDeletesTheIPNSKey)
	ctx.Step(`^the IPNS key is no longer in the list$`, s.theIPNSKeyIsNoLongerInTheList)
}

// theUserCreatesAnIPNSKey creates a new IPNS key with the specified name
func (s *IPNSKeyManagementSteps) theUserCreatesAnIPNSKey(ctx context.Context, name string) (context.Context, error) {
	var key *ipfs_sdk.IPNSKeyResponse
	var err error
	ctx, key, err = helpers.CreateIPNSKey(ctx, name)
	if err != nil {
		return ctx, err
	}

	if key == nil {
		return ctx, fmt.Errorf("IPNS key creation returned nil response")
	}

	return ctx, nil
}

// anIPNSKeyWithNameIsCreated verifies that an IPNS key with the specified name exists
func (s *IPNSKeyManagementSteps) anIPNSKeyWithNameIsCreated(ctx context.Context, expectedName string) (context.Context, error) {
	keyID, ok := helpers.GetIPNSKeyID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS key ID found in context")
	}

	key, err := helpers.GetIPNSKey(ctx, strconv.Itoa(keyID))
	if err != nil {
		return ctx, err
	}

	if key == nil {
		return ctx, fmt.Errorf("IPNS key %d not found", keyID)
	}

	if key.Name != expectedName {
		return ctx, fmt.Errorf("IPNS key has name %s, expected %s", key.Name, expectedName)
	}
	// Store key details in context for verification
	ctx = helpers.SetIPNSKeyName(ctx, key.Name)
	if key.PeerId != "" {
		ctx = helpers.SetIPNSPeerID(ctx, key.PeerId)
	}
	if key.IpnsName != "" {
		ctx = helpers.SetIPNSIPNSName(ctx, key.IpnsName)
	}

	return ctx, nil
}

// theIPNSKeyHasValidPeerIDFromKey verifies that the IPNS key has a valid peer ID
// This is a delegated helper that uses the common step
func (s *IPNSKeyManagementSteps) theIPNSKeyHasValidPeerIDFromKey(ctx context.Context) (context.Context, error) {
	commonSteps := NewIPNSCommonSteps()
	return commonSteps.theIPNSKeyHasValidPeerID(ctx)
}

// theUserCreatesMultipleIPNSKeys creates the specified number of IPNS keys
func (s *IPNSKeyManagementSteps) theUserCreatesMultipleIPNSKeys(ctx context.Context, count int) (context.Context, error) {
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("test-key-%d", i)
		var err error
		ctx, _, err = helpers.CreateIPNSKey(ctx, name)
		if err != nil {
			return ctx, fmt.Errorf("failed to create IPNS key %d: %w", i, err)
		}
	}

	return ctx, nil
}

// theUserListsTheirIPNSKeys retrieves all IPNS keys for the authenticated user
func (s *IPNSKeyManagementSteps) theUserListsTheirIPNSKeys(ctx context.Context) (context.Context, error) {
	keys, err := helpers.ListIPNSKeys(ctx)
	if err != nil {
		return ctx, err
	}

	if len(keys) == 0 {
		return ctx, fmt.Errorf("no IPNS keys found for user")
	}

	return ctx, nil
}

// allIPNSKeysAreReturned verifies that the expected number of keys were returned
func (s *IPNSKeyManagementSteps) allIPNSKeysAreReturned(ctx context.Context, expectedCount int) (context.Context, error) {
	keys, err := helpers.ListIPNSKeys(ctx)
	if err != nil {
		return ctx, err
	}

	if len(keys) != expectedCount {
		return ctx, fmt.Errorf("received %d IPNS keys, expected %d", len(keys), expectedCount)
	}

	return ctx, nil
}

// eachKeyHasUniquePeerID verifies that all keys have unique peer IDs
func (s *IPNSKeyManagementSteps) eachKeyHasUniquePeerID(ctx context.Context) (context.Context, error) {
	keys, err := helpers.ListIPNSKeys(ctx)
	if err != nil {
		return ctx, err
	}

	peerIDs := make(map[string]bool)
	for _, key := range keys {
		if key.PeerId == "" {
			return ctx, fmt.Errorf("key %s has empty peer ID", key.Name)
		}

		if peerIDs[key.PeerId] {
			return ctx, fmt.Errorf("duplicate peer ID %s found", key.PeerId)
		}

		peerIDs[key.PeerId] = true
	}

	return ctx, nil
}

// theUserHasMultipleIPNSKeys is a Given step that creates the specified number of keys
func (s *IPNSKeyManagementSteps) theUserHasMultipleIPNSKeys(ctx context.Context, count int) (context.Context, error) {
	// Create keys with unique names
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("test-key-%d", i)
		var err error
		ctx, _, err = helpers.CreateIPNSKey(ctx, name)
		if err != nil {
			return ctx, fmt.Errorf("failed to create IPNS key %d: %w", i, err)
		}
	}

	return ctx, nil
}

// theUserDeletesTheIPNSKey deletes the current IPNS key
func (s *IPNSKeyManagementSteps) theUserDeletesTheIPNSKey(ctx context.Context) (context.Context, error) {
	keyID, ok := helpers.GetIPNSKeyID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS key ID found in context")
	}

	err := helpers.DeleteIPNSKey(ctx, strconv.Itoa(keyID))
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

// theIPNSKeyIsNoLongerInTheList verifies that the IPNS key has been deleted
func (s *IPNSKeyManagementSteps) theIPNSKeyIsNoLongerInTheList(ctx context.Context) (context.Context, error) {
	keyID, ok := helpers.GetIPNSKeyID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS key ID found in context")
	}

	_, err := helpers.GetIPNSKey(ctx, strconv.Itoa(keyID))
	if err == nil {
		return ctx, fmt.Errorf("IPNS key %d still exists, should have been deleted", keyID)
	}

	return ctx, nil
}

// NewIPNSService is a helper function to get the IPNS service from the IPNS context
// This maintains consistency with the IPFS pinning pattern
func NewIPNSService() error {
	// This is a placeholder for service initialization if needed
	// The actual service is obtained via helpers.GetIPNSService(ctx)
	return nil
}
