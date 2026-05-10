package helpers

import (
	"context"
	"net/http"

	"go.sia.tech/core/types"
	"go.sia.tech/siastorage"
)

// SetSiaAppKey stores the Sia app key in context
func SetSiaAppKey(ctx context.Context, key string) context.Context {
	return SetContextValue(ctx, SiaAppKeyKey, key)
}

// GetSiaAppKey retrieves the Sia app key from context
func GetSiaAppKey(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, SiaAppKeyKey)
}

// SetSiaMnemonic stores the Sia mnemonic in context
func SetSiaMnemonic(ctx context.Context, mnemonic string) context.Context {
	return SetContextValue(ctx, SiaMnemonicKey, mnemonic)
}

// GetSiaMnemonic retrieves the Sia mnemonic from context
func GetSiaMnemonic(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, SiaMnemonicKey)
}

// SetSiaConnectRequestID stores the Sia connect request ID in context
func SetSiaConnectRequestID(ctx context.Context, id string) context.Context {
	return SetContextValue(ctx, SiaConnectRequestIDKey, id)
}

// GetSiaConnectRequestID retrieves the Sia connect request ID from context
func GetSiaConnectRequestID(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, SiaConnectRequestIDKey)
}

// SetSiaObjectKey stores the Sia object key in context
func SetSiaObjectKey(ctx context.Context, key types.Hash256) context.Context {
	return SetContextValue(ctx, SiaObjectKeyKey, key)
}

// GetSiaObjectKey retrieves the Sia object key from context
func GetSiaObjectKey(ctx context.Context) (types.Hash256, bool) {
	return GetContextValue[types.Hash256](ctx, SiaObjectKeyKey)
}

// SetSiaSlabID stores the Sia slab ID in context
func SetSiaSlabID(ctx context.Context, id string) context.Context {
	return SetContextValue(ctx, SiaSlabIDKey, id)
}

// GetSiaSlabID retrieves the Sia slab ID from context
func GetSiaSlabID(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, SiaSlabIDKey)
}

// SetSiaSDK stores the Sia SDK instance in context
func SetSiaSDK(ctx context.Context, sdk *siastorage.SDK) context.Context {
	return SetContextValue(ctx, SiaSDKKey, sdk)
}

// GetSiaSDK retrieves the Sia SDK instance from context
func GetSiaSDK(ctx context.Context) (*siastorage.SDK, bool) {
	return GetContextValue[*siastorage.SDK](ctx, SiaSDKKey)
}

// AddSiaObjectCleanup adds a Sia object key to cleanup list
func AddSiaObjectCleanup(ctx context.Context, key string) context.Context {
	return AddToCleanupList(ctx, SiaObjectsCleanupKey, key)
}

// GetSiaObjectsCleanup retrieves Sia objects cleanup list
func GetSiaObjectsCleanup(ctx context.Context) []string {
	return GetCleanupList[string](ctx, SiaObjectsCleanupKey)
}

// AddSiaSlabCleanup adds a Sia slab ID to cleanup list
func AddSiaSlabCleanup(ctx context.Context, id string) context.Context {
	return AddToCleanupList(ctx, SiaSlabsCleanupKey, id)
}

// GetSiaSlabsCleanup retrieves Sia slabs cleanup list
func GetSiaSlabsCleanup(ctx context.Context) []string {
	return GetCleanupList[string](ctx, SiaSlabsCleanupKey)
}

func SetSiaConnectApproved(ctx context.Context) context.Context {
	return SetContextValue(ctx, SiaConnectApprovedKey, true)
}

func GetSiaConnectApproved(ctx context.Context) bool {
	v, ok := GetContextValue[bool](ctx, SiaConnectApprovedKey)
	return ok && v
}

func SetSiaAppID(ctx context.Context, id string) context.Context {
	return SetContextValue(ctx, SiaAppIDKey, id)
}

func GetSiaAppID(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, SiaAppIDKey)
}

// SetSiaConnectUIResponse stores the Sia connect UI HTTP response in context
func SetSiaConnectUIResponse(ctx context.Context, resp *http.Response) context.Context {
	return context.WithValue(ctx, SiaConnectUIResponseKey, resp)
}

// GetSiaConnectUIResponse retrieves the Sia connect UI HTTP response from context
func GetSiaConnectUIResponse(ctx context.Context) (*http.Response, bool) {
	resp, ok := ctx.Value(SiaConnectUIResponseKey).(*http.Response)
	return resp, ok
}
