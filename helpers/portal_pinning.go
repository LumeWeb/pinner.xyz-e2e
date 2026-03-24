package helpers

import (
	"context"
	"fmt"
	"strings"
	"time"

	goCid "github.com/ipfs/go-cid"
	"go.lumeweb.com/ipfs-sdk"
	account "go.lumeweb.com/portal-sdk"
	"github.com/samber/lo"
)

// DefaultOperationTimeout is the default timeout for IPFS operation completion.
// This gives operations sufficient time to complete before timing out.
const DefaultOperationTimeout = 30 * time.Minute



// PortalPinning wraps Portal SDK pinning operations
type PortalPinning struct {
	client *ipfs.Client
}

// NewPortalPinning creates a new PortalPinning instance
func NewPortalPinning(ctx context.Context) (*PortalPinning, error) {
	client, err := GetIPFSClient(ctx)
	if err != nil {
		return nil, err
	}

	return &PortalPinning{client: client}, nil
}

// WaitForPinStatus polls the pin status until it reaches the desired status
// Returns the final pin status and an error if the timeout is reached
func WaitForPinStatus(ctx context.Context, requestID string, desiredStatus ipfs.PinStatusEnum, timeout time.Duration) (*ipfs.PinStatus, error) {
	pp, err := NewPortalPinning(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create pinning client: %w", err)
	}

	deadline := time.Now().Add(timeout)
	pollInterval := 500 * time.Millisecond

	for time.Now().Before(deadline) {
		pinStatus, err := pp.client.Pinning().GetPin(ctx, requestID)
		if err != nil {
			return nil, fmt.Errorf("failed to get pin status: %w", err)
		}

		currentStatus := pinStatus.PinStatusEnum
		if currentStatus == desiredStatus {
			return pinStatus, nil
		}

		if currentStatus == ipfs.StatusFailed {
			return nil, fmt.Errorf("pin operation failed for request ID %s", requestID)
		}

		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("timeout waiting for pin to reach status %v (request ID: %s, timeout: %s)", desiredStatus, requestID, timeout)
}

// AddPin pins the given CID using Portal SDK
// Returns the PinStatus which contains the request ID for cleanup
func (pp *PortalPinning) AddPin(ctx context.Context, cidString string) (*ipfs.PinStatus, error) {
	cid, err := goCid.Decode(cidString)
	if err != nil {
		return nil, fmt.Errorf("failed to decode CID %s: %w", cidString, err)
	}

	pinStatus, err := pp.client.Pinning().AddPin(ctx, cid)
	if err != nil {
		return nil, fmt.Errorf("failed to pin CID %s: %w", cidString, err)
	}

	return pinStatus, nil
}

// ListPins lists all pinned CIDs using Portal SDK
func (pp *PortalPinning) ListPins(ctx context.Context) ([]string, error) {
	pins, err := pp.client.Pinning().ListPins(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list pins: %w", err)
	}

	// Extract CID strings from pins
	pinCids := lo.Map(pins, func(pin ipfs.PinStatus, _ int) string {
		if pin.Pin.Cid == "" {
			return ""
		}
		return pin.Pin.Cid
	})

	return lo.Filter(pinCids, func(cid string, _ int) bool {
		return cid != ""
	}), nil
}

// getPinRequestID extracts the request ID from a PinStatus
func getPinRequestID(pin *ipfs.PinStatus) (string, error) {
	if pin == nil {
		return "", fmt.Errorf("pin is nil")
	}

	return pin.Requestid, nil
}

// RemovePin removes a pin for the given CID using Portal SDK
// Note: The pinning API's RemovePin method takes a requestID (UUID), not a CID.
// We must first find the requestID by listing pins filtered by CID.
func (pp *PortalPinning) RemovePin(ctx context.Context, cidString string) error {
	// List pins filtering by this CID to find the requestID
	pins, err := pp.client.Pinning().ListPins(ctx, ipfs.WithFilterCIDs(cidString))
	if err != nil {
		return fmt.Errorf("failed to list pins: %w", err)
	}

	if len(pins) == 0 {
		return fmt.Errorf("no pin found for CID %s", cidString)
	}

	// Get the first matching pin's request ID
	requestID, err := getPinRequestID(&pins[0])
	if err != nil {
		return fmt.Errorf("failed to get request ID for CID %s: %w", cidString, err)
	}

	err = pp.client.Pinning().RemovePin(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to unpin CID %s: %w", cidString, err)
	}

	return nil
}

// GetPin checks if a CID is pinned using Portal SDK
// Note: The pinning API's GetPin method takes a requestID (UUID), not a CID.
// To check if a CID is pinned, we must use ListPins with filtering.
// Returns true only if the CID exists and has StatusPinned
func (pp *PortalPinning) GetPin(ctx context.Context, cidString string) (bool, error) {
	// List pins filtering by this CID
	pins, err := pp.client.Pinning().ListPins(ctx, ipfs.WithFilterCIDs(cidString))
	if err != nil {
		return false, fmt.Errorf("failed to list pins: %w", err)
	}

	// No pins found
	if len(pins) == 0 {
		return false, nil
	}

	// Check if the pin status is StatusPinned (not just Queue or Pinning)
	for _, pin := range pins {
		if pin.PinStatusEnum == ipfs.StatusPinned {
			return true, nil
		}
	}

	// CID is being pinned but not yet pinned
	return false, nil
}

// WaitForOperationCompleteByCID waits for the account operation for a CID to reach StatusCompleted
// Uses the portal SDK's WaitForOperation method for polling
func WaitForOperationCompleteByCID(ctx context.Context, cid string, timeout time.Duration) error {
	
	// Get authenticated account client from context
	api, err := RequireAuthenticatedClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authenticated client: %w", err)
	}

	deadline := time.Now().Add(timeout)
	pollInterval := 500 * time.Millisecond
	pollCount := 0

	for time.Now().Before(deadline) {
		pollCount++
		
		// List operations filtered by CID
		operations, err := api.ListOperations(ctx)
		if err != nil {
			return fmt.Errorf("failed to list operations: %w", err)
		}

		// Find the operation for this CID
		var targetOp *account.Operation
		for _, op := range operations {
			if op.Cid != nil {
				equal, err := CIDsEqual(*op.Cid, cid)
				if err != nil {
					continue
				}
				if equal {
					targetOp = op
					break
				}
			}
		}

			if targetOp == nil {
			// No operation found yet - wait and retry
			time.Sleep(pollInterval)
			continue
		}

		// Use the SDK's built-in WaitForOperation with the operation ID
		remainingTime := time.Until(deadline)
		_, err = api.WaitForOperation(ctx, int64(targetOp.Id),
			account.WithPollTimeout(remainingTime),
			account.WithPollInterval(pollInterval),
			account.WithPollSettledStates(account.OperationStatusCompleted),
		)
		if err != nil {
			return fmt.Errorf("operation for CID %s failed: %w", cid, err)
		}

		return nil
	}

	return fmt.Errorf("timeout waiting for operation for CID %s (timeout: %s)", cid, timeout)
}

// IPFSPinAdd pins the given CID using Portal SDK
// Returns the created pin status and updated context with CID and request ID tracked for cleanup
// Note: This creates the pin without waiting for completion. Use WaitForPinCreation and WaitForOperation for full lifecycle control.
func IPFSPinAdd(ctx context.Context, cidString string) (*ipfs.PinStatus, context.Context, error) {
	pp, err := NewPortalPinning(ctx)
	if err != nil {
		return nil, ctx, err
	}

	pinStatus, err := pp.AddPin(ctx, cidString)
	if err != nil {
		return nil, ctx, err
	}

	// Track request ID for cleanup (do this early so cleanup works even if waiting fails)
	if pinStatus != nil {
		ctx = AddPinRequestIDCleanup(ctx, pinStatus.Requestid)
	}

	// Track CID for cleanup
	ctx = AddCIDCleanup(ctx, cidString)

	return pinStatus, ctx, nil
}

// WaitForPinCreation waits for a pin to reach StatusPinned
// Call after IPFSPinAdd to wait for pinning completion
func WaitForPinCreation(ctx context.Context, cid string) error {
	pp, err := NewPortalPinning(ctx)
	if err != nil {
		return fmt.Errorf("failed to create pinning client: %w", err)
	}

	pins, err := pp.client.Pinning().ListPins(ctx, ipfs.WithFilterCIDs(cid))
	if err != nil {
		return fmt.Errorf("failed to list pins: %w", err)
	}

	if len(pins) == 0 {
		return fmt.Errorf("no pin found for CID %s", cid)
	}

	requestID := pins[0].Requestid
	_, err = WaitForPinStatus(ctx, requestID, ipfs.StatusPinned, DefaultOperationTimeout)
	if err != nil {
		return fmt.Errorf("pin create succeeded but wait for StatusPinned failed: %w", err)
	}

	return nil
}

// WaitForOperation waits for the account operation for a CID to reach StatusCompleted
// Call after WaitForPinCreation to verify operation completion
func WaitForOperation(ctx context.Context, cid string) error {
	return WaitForOperationCompleteByCID(ctx, cid, DefaultOperationTimeout)
}

// IPFSPinLs lists all pinned CIDs using Portal SDK (legacy wrapper)
func IPFSPinLs(ctx context.Context) ([]string, error) {
	pp, err := NewPortalPinning(ctx)
	if err != nil {
		return nil, err
	}
	return pp.ListPins(ctx)
}

// IPFSPinRm removes a pin for the given CID using Portal SDK (legacy wrapper)
func IPFSPinRm(ctx context.Context, cidString string) error {
	pp, err := NewPortalPinning(ctx)
	if err != nil {
		return err
	}
	return pp.RemovePin(ctx, cidString)
}

// IPFSIsPinned checks if a CID is pinned using Portal SDK (legacy wrapper)
func IPFSIsPinned(ctx context.Context, cidString string) (bool, error) {
	pp, err := NewPortalPinning(ctx)
	if err != nil {
		return false, err
	}
	return pp.GetPin(ctx, cidString)
}

// GetPinCount returns the number of pinned items with the given prefix (for testing)
func GetPinCount(ctx context.Context, prefix string) (int, error) {
	pins, err := IPFSPinLs(ctx)
	if err != nil {
		return 0, err
	}

	matching := lo.Filter(pins, func(item string, _ int) bool {
		return strings.Contains(item, prefix)
	})

	return len(matching), nil
}

// VerifyPinned checks if a CID is pinned by listing pins (GetPin takes requestID, not CID)
// Note: This is used for verification after upload; it returns the pin status if pinned
func VerifyPinned(ctx context.Context, cidString string) (bool, error) {
	pinned, err := IPFSIsPinned(ctx, cidString)
	if err != nil {
		return false, err
	}
	return pinned, nil
}

// CleanupAllPinsForUser removes all pins for the current authenticated user
// Used for test isolation before scenarios that require a clean slate
// This function will silently succeed if authentication is not complete yet
func CleanupAllPinsForUser(ctx context.Context) error {
	pp, err := NewPortalPinning(ctx)
	if err != nil {
		// Silently fail if not authenticated yet - JWT might not be available in BeforeScenario
		return nil
	}

	// List all pins
	pins, err := pp.client.Pinning().ListPins(ctx)
	if err != nil {
		// Silently fail if not authenticated or if pins service is unavailable
		return nil
	}

	if len(pins) == 0 {
		return nil
	}

	// Delete all pins by request ID
	deleteCount := 0
	for i, pin := range pins {
		// PinStatus is a struct, not a pointer, so we check if Requestid is empty
		if pin.Requestid == "" {
			continue
		}
		requestID := pin.Requestid
		if requestID != "" {
			if err := pp.client.Pinning().RemovePin(ctx, requestID); err != nil {
				// Log but continue - some pins may have already been deleted
				fmt.Printf("[WARN] Failed to delete pin %d (requestID: %s): %v\n", i, requestID, err)
			} else {
				deleteCount++
			}
		}
	}

	fmt.Printf("[Cleanup] Deleted %d pins\n", deleteCount)
	return nil
}
