package helpers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.sia.tech/core/types"
	"go.sia.tech/indexd/api/app"
	"go.sia.tech/indexd/slabs"
	"go.sia.tech/siastorage"
)

func SiaTestAppMetadata() siastorage.AppMetadata {
	return siastorage.AppMetadata{
		ID:          siastorage.GenerateAppID(),
		Name:        "e2e-test-app",
		Description: "E2E Test Application",
		ServiceURL:  GetSiaEndpoint(),
	}
}

// SiaSDKConnect performs the full Sia SDK connect flow:
// generate app ID, generate mnemonic, request connection, approve through portal API,
// wait for approval, register, and store SDK + mnemonic + app key in context.
func SiaSDKConnect(ctx context.Context) (context.Context, error) {
	indexerURL := GetSiaEndpoint()
	if indexerURL == "" {
		return ctx, fmt.Errorf("sia endpoint not configured")
	}

	metadata := SiaTestAppMetadata()
	mnemonic := siastorage.NewSeedPhrase()

	builder := siastorage.NewBuilder(indexerURL, metadata)

	responseURL, err := builder.RequestConnection(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to request sia connection: %w", err)
	}

	ctx = SetSiaConnectRequestID(ctx, responseURL)

	requestID, err := ParseRequestIDFromConnectURL(responseURL)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse SIA connect request ID: %w", err)
	}

	if err := ApproveSiaConnect(ctx, requestID); err != nil {
		return ctx, fmt.Errorf("failed to approve SIA connection: %w", err)
	}

	approved, err := builder.WaitForApproval(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to wait for sia approval: %w", err)
	}
	if !approved {
		return ctx, fmt.Errorf("sia connection request was denied")
	}

	ctx = SetSiaConnectApproved(ctx)

	sdk, err := builder.Register(ctx, mnemonic)
	if err != nil {
		return ctx, fmt.Errorf("failed to register sia app: %w", err)
	}

	appKey := sdk.AppKey()
	appKeyStr := fmt.Sprintf("%x", appKey[:])

	ctx = SetSiaSDK(ctx, sdk)
	ctx = SetSiaMnemonic(ctx, mnemonic)
	ctx = SetSiaAppKey(ctx, appKeyStr)
	ctx = SetSiaAppID(ctx, metadata.ID.String())

	return ctx, nil
}

// SiaSDKFromAppKey reconnects to the Sia indexer using a stored app key.
func SiaSDKFromAppKey(ctx context.Context) (context.Context, error) {
	appKeyStr, ok := GetSiaAppKey(ctx)
	if !ok {
		return ctx, fmt.Errorf("no sia app key found in context")
	}

	appKeyBytes, err := hex.DecodeString(appKeyStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to decode sia app key: %w", err)
	}
	appKey := types.PrivateKey(appKeyBytes)

	indexerURL := GetSiaEndpoint()
	if indexerURL == "" {
		return ctx, fmt.Errorf("sia endpoint not configured")
	}

	var appID types.Hash256
	appIDStr, hasAppID := GetSiaAppID(ctx)
	if hasAppID {
		if err := appID.UnmarshalText([]byte(appIDStr)); err != nil {
			return ctx, fmt.Errorf("failed to decode sia app ID: %w", err)
		}
	}

	metadata := siastorage.AppMetadata{
		ID:          appID,
		Name:        "e2e-test-app",
		Description: "E2E Test Application",
		ServiceURL:  indexerURL,
	}

	builder := siastorage.NewBuilder(indexerURL, metadata)

	sdk, err := builder.SDK(appKey)
	if err != nil {
		return ctx, fmt.Errorf("failed to create sia SDK from app key: %w", err)
	}

	ctx = SetSiaSDK(ctx, sdk)
	return ctx, nil
}

// SiaUploadData uploads data using the Sia SDK with specified redundancy,
// pins the object, and stores the object key in context.
func SiaUploadData(ctx context.Context, data io.Reader, size int64, redundancyData, redundancyParity uint8) (context.Context, error) {
	return SiaUploadDataWithMetadata(ctx, data, size, redundancyData, redundancyParity, nil)
}

// SiaUploadDataWithMetadata uploads data with optional custom metadata.
// If metadata is non-nil, it is attached to the object before pinning.
func SiaUploadDataWithMetadata(ctx context.Context, data io.Reader, size int64, redundancyData, redundancyParity uint8, metadata json.RawMessage) (context.Context, error) {
	sdk, ok := GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return ctx, fmt.Errorf("no sia SDK available in context")
	}

	obj := siastorage.NewEmptyObject()

	if metadata != nil {
		obj.UpdateMetadata(metadata)
	}

	if err := sdk.Upload(ctx, &obj, data, siastorage.WithRedundancy(redundancyData, redundancyParity)); err != nil {
		return ctx, fmt.Errorf("failed to upload data to sia: %w", err)
	}

	if err := sdk.PinObject(ctx, obj); err != nil {
		return ctx, fmt.Errorf("failed to pin sia object: %w", err)
	}

	objectKey := obj.ID()
	ctx = SetSiaObjectKey(ctx, objectKey)
	ctx = AddSiaObjectCleanup(ctx, objectKey.String())

	return ctx, nil
}

// SiaUploadRandomData generates random data of the given size and uploads it to Sia.
func SiaUploadRandomData(ctx context.Context, size int64) (context.Context, error) {
	content := make([]byte, size)
	if _, err := rand.Read(content); err != nil {
		return ctx, fmt.Errorf("failed to generate random data: %w", err)
	}

	ctx = SetKnownContent(ctx, string(content))

	return SiaUploadData(ctx, bytes.NewReader(content), size, 1, 1)
}

// SiaDownloadObject downloads the current object from context using the Sia SDK.
func SiaDownloadObject(ctx context.Context, w io.Writer) error {
	sdk, ok := GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return fmt.Errorf("no sia SDK available in context")
	}

	objectKey, ok := GetSiaObjectKey(ctx)
	if !ok {
		return fmt.Errorf("no sia object key found in context")
	}

	obj, err := sdk.Object(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("failed to get sia object: %w", err)
	}

	if err := sdk.Download(ctx, w, obj); err != nil {
		return fmt.Errorf("failed to download sia object: %w", err)
	}

	return nil
}

// SiaDeleteObject deletes a Sia object by its key.
func SiaDeleteObject(ctx context.Context, key types.Hash256) error {
	sdk, ok := GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return fmt.Errorf("no sia SDK available in context")
	}

	if err := sdk.DeleteObject(ctx, key); err != nil {
		return fmt.Errorf("failed to delete sia object: %w", err)
	}

	return nil
}

// SiaListObjectEvents lists object events from the Sia indexer.
func SiaListObjectEvents(ctx context.Context, cursor slabs.Cursor, limit int) ([]siastorage.Object, error) {
	sdk, ok := GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return nil, fmt.Errorf("no sia SDK available in context")
	}

	objects, err := sdk.ListObjects(ctx, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list sia object events: %w", err)
	}

	return objects, nil
}

// SiaPruneSlabs prunes orphaned slabs from the Sia indexer.
func SiaPruneSlabs(ctx context.Context) error {
	sdk, ok := GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return fmt.Errorf("no sia SDK available in context")
	}

	if err := sdk.PruneSlabs(ctx); err != nil {
		return fmt.Errorf("failed to prune sia slabs: %w", err)
	}

	return nil
}

// SiaGetAccount retrieves account information from the Sia indexer.
func SiaGetAccount(ctx context.Context) (app.AccountResponse, error) {
	sdk, ok := GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return app.AccountResponse{}, fmt.Errorf("no sia SDK available in context")
	}

	account, err := sdk.Account(ctx)
	if err != nil {
		return app.AccountResponse{}, fmt.Errorf("failed to get sia account: %w", err)
	}

	return account, nil
}

func WaitForSiaReady(ctx context.Context) error {
	sdk, ok := GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return fmt.Errorf("no sia SDK available in context")
	}

	timeout := time.After(15 * time.Minute)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		account, err := sdk.Account(ctx)
		if err == nil && account.Ready {
			return nil
		}

		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for SIA account to be ready")
		case <-ticker.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// ParseRequestIDFromConnectURL extracts the request ID from a Sia connect response URL.
// The URL format is: {portalBaseURL}/auth/connect/{requestID}
func ParseRequestIDFromConnectURL(connectURL string) (string, error) {
	u, err := url.Parse(connectURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse connect URL: %w", err)
	}
	// Path should be /auth/connect/{requestID}
	parts := strings.Split(strings.TrimSuffix(u.Path, "/"), "/")
	if len(parts) < 4 || parts[len(parts)-2] != "connect" {
		return "", fmt.Errorf("unexpected connect URL path format: %s", u.Path)
	}
	requestID := parts[len(parts)-1]
	if requestID == "" {
		return "", fmt.Errorf("empty request ID in connect URL: %s", connectURL)
	}
	return requestID, nil
}

// ApproveSiaConnect approves a Sia connect request through the portal API.
// The portal injects the connect key server-side, so the client only needs
// to POST {"approve": true} with JWT authentication.
func ApproveSiaConnect(ctx context.Context, requestID string) error {
	token, ok := GetJWTToken(ctx)
	if !ok || token == "" {
		return fmt.Errorf("no JWT token available for Sia connect approval")
	}

	approveURL := fmt.Sprintf("%s/auth/connect/%s", GetSiaEndpoint(), requestID)

	body := bytes.NewReader([]byte(`{"approve": true}`))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, approveURL, body)
	if err != nil {
		return fmt.Errorf("failed to create approval request: %w", err)
	}

	req.Host = GetSiaHost()
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send approval request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SIA connect approval failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// RejectSiaConnect rejects a Sia connect request through the portal API.
func RejectSiaConnect(ctx context.Context, requestID string) error {
	token, ok := GetJWTToken(ctx)
	if !ok || token == "" {
		return fmt.Errorf("no JWT token available for Sia connect rejection")
	}

	rejectURL := fmt.Sprintf("%s/auth/connect/%s", GetSiaEndpoint(), requestID)

	body := bytes.NewReader([]byte(`{"approve": false}`))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rejectURL, body)
	if err != nil {
		return fmt.Errorf("failed to create rejection request: %w", err)
	}

	req.Host = GetSiaHost()
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send rejection request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SIA connect rejection failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
