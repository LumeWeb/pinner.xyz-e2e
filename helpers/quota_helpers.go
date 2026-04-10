package helpers

import (
	"context"
	"fmt"

	account "go.lumeweb.com/portal-sdk"
)

// Context keys for quota management (minimal - the rest are in test_helpers.go)
const (
	QuotaStatusKey contextKey = "quota_status"
)

// SetQuotaStatus stores the account quota status in context
func SetQuotaStatus(ctx context.Context, quota *account.QuotaStatus) context.Context {
	return SetContextValue(ctx, QuotaStatusKey, quota)
}

// GetQuotaStatus retrieves the account quota status from context
func GetQuotaStatus(ctx context.Context) (*account.QuotaStatus, bool) {
	return GetContextValue[*account.QuotaStatus](ctx, QuotaStatusKey)
}

// RequireQuotaStatus retrieves the account quota status or returns an error
func RequireQuotaStatus(ctx context.Context) (*account.QuotaStatus, error) {
	quota, ok := GetQuotaStatus(ctx)
	if !ok {
		return nil, fmt.Errorf("quota status not available in context")
	}
	return quota, nil
}

// QuotaTypeStatus represents a quota type status (upload, download, bandwidth)
// This is an alias to account.QuotaTypeStatus for convenience
type QuotaTypeStatus = account.QuotaTypeStatus
