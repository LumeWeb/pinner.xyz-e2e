package helpers

import (
	"context"
	"fmt"
	"strings"

	// Side effects import to ensure admin subpackage is vendored
	admin "go.lumeweb.com/portal-sdk/admin"
)

// Context keys for admin quota management
const (
	AdminQuotaPlansKey       contextKey = "admin_quota_plans"
	AdminCurrentPlanKey      contextKey = "admin_current_plan"
	AdminQuotaAllowancesKey  contextKey = "admin_quota_allowances"
	AdminCurrentAllowanceKey contextKey = "admin_current_allowance"
	AdminSystemStatsKey      contextKey = "admin_system_stats"
	AdminTargetUserIDKey     contextKey = "admin_target_user_id"
)

// SetAdminQuotaPlans stores admin quota plans in context
func SetAdminQuotaPlans(ctx context.Context, plans []*admin.QuotaPlan) context.Context {
	return SetContextValue(ctx, AdminQuotaPlansKey, plans)
}

// GetAdminQuotaPlans retrieves admin quota plans from context
func GetAdminQuotaPlans(ctx context.Context) ([]*admin.QuotaPlan, bool) {
	return GetContextValue[[]*admin.QuotaPlan](ctx, AdminQuotaPlansKey)
}

// SetAdminCurrentPlan stores the current admin quota plan in context
func SetAdminCurrentPlan(ctx context.Context, plan *admin.QuotaPlan) context.Context {
	return SetContextValue(ctx, AdminCurrentPlanKey, plan)
}

// GetAdminCurrentPlan retrieves the current admin quota plan from context
func GetAdminCurrentPlan(ctx context.Context) (*admin.QuotaPlan, bool) {
	return GetContextValue[*admin.QuotaPlan](ctx, AdminCurrentPlanKey)
}

// RequireAdminCurrentPlan retrieves the current admin quota plan or returns an error
func RequireAdminCurrentPlan(ctx context.Context) (*admin.QuotaPlan, error) {
	plan, ok := GetAdminCurrentPlan(ctx)
	if !ok {
		return nil, fmt.Errorf("admin quota plan not available in context")
	}
	return plan, nil
}

// SetAdminQuotaAllowances stores admin quota allowances in context
func SetAdminQuotaAllowances(ctx context.Context, allowances []*admin.QuotaAllowance) context.Context {
	return SetContextValue(ctx, AdminQuotaAllowancesKey, allowances)
}

// GetAdminQuotaAllowances retrieves admin quota allowances from context
func GetAdminQuotaAllowances(ctx context.Context) ([]*admin.QuotaAllowance, bool) {
	return GetContextValue[[]*admin.QuotaAllowance](ctx, AdminQuotaAllowancesKey)
}

// SetAdminCurrentAllowance stores the current admin quota allowance in context
func SetAdminCurrentAllowance(ctx context.Context, allowance *admin.QuotaAllowance) context.Context {
	return SetContextValue(ctx, AdminCurrentAllowanceKey, allowance)
}

// GetAdminCurrentAllowance retrieves the current admin quota allowance from context
func GetAdminCurrentAllowance(ctx context.Context) (*admin.QuotaAllowance, bool) {
	return GetContextValue[*admin.QuotaAllowance](ctx, AdminCurrentAllowanceKey)
}

// RequireAdminCurrentAllowance retrieves the current admin quota allowance or returns an error
func RequireAdminCurrentAllowance(ctx context.Context) (*admin.QuotaAllowance, error) {
	allowance, ok := GetAdminCurrentAllowance(ctx)
	if !ok {
		return nil, fmt.Errorf("admin quota allowance not available in context")
	}
	return allowance, nil
}

// SetAdminSystemStats stores admin system stats in context
func SetAdminSystemStats(ctx context.Context, stats *admin.SystemStats) context.Context {
	return SetContextValue(ctx, AdminSystemStatsKey, stats)
}

// GetAdminSystemStats retrieves admin system stats from context
func GetAdminSystemStats(ctx context.Context) (*admin.SystemStats, bool) {
	return GetContextValue[*admin.SystemStats](ctx, AdminSystemStatsKey)
}

// RequireAdminSystemStats retrieves admin system stats or returns an error
func RequireAdminSystemStats(ctx context.Context) (*admin.SystemStats, error) {
	stats, ok := GetAdminSystemStats(ctx)
	if !ok {
		return nil, fmt.Errorf("admin system stats not available in context")
	}
	return stats, nil
}

// SetAdminTargetUserID stores target user ID in context
func SetAdminTargetUserID(ctx context.Context, userID int) context.Context {
	return SetContextValue(ctx, AdminTargetUserIDKey, userID)
}

// GetAdminTargetUserID retrieves target user ID from context
func GetAdminTargetUserID(ctx context.Context) (int, bool) {
	return GetContextValue[int](ctx, AdminTargetUserIDKey)
}

// RequireAdminTargetUserID retrieves target user ID or returns an error
func RequireAdminTargetUserID(ctx context.Context) (int, error) {
	userID, ok := GetAdminTargetUserID(ctx)
	if !ok {
		return 0, fmt.Errorf("target user ID not available in context")
	}
	return userID, nil
}

// CleanupAdminQuota cleans up quota plans and allowances created during testing
func CleanupAdminQuota(ctx context.Context) error {
	// Get the cleanup lists - they will be empty for non-admin scenarios
	planIDs := GetQuotaPlansCleanup(ctx)
	allowanceIDs := GetQuotaAllowancesCleanup(ctx)

	// If nothing to clean up, return early (non-admin scenarios or empty)
	if len(planIDs) == 0 && len(allowanceIDs) == 0 {
		return nil
	}

	// Only try to use admin client if we have something to clean up
	adminClient, err := RequireAdminClient(ctx)
	if err != nil {
		return fmt.Errorf("admin client not available for cleanup: %w", err)
	}

	// Check if the Quota interface is available before proceeding
	quotaAdmin := adminClient.Quota()
	if quotaAdmin == nil {
		return fmt.Errorf("admin client's Quota() returned nil")
	}

	// Defensive cleanup: Before deleting plans, reset any users assigned to them
	for _, planID := range planIDs {
		_ = ResetUsersForPlan(ctx, quotaAdmin, planID)
	}

	for _, planID := range planIDs {
		// Before attempting deletion, check if this is the default plan
		// Default plans cannot be deleted directly, so we need to unset the default flag first
		plan, err := getPlanSafely(ctx, quotaAdmin, planID)
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
			continue
		}

		// If the plan is default, clear the default flag before deletion
		if err == nil && plan.IsDefault {
			// Update the plan to remove default flag
			plan.IsDefault = false
			_, updateErr := quotaAdmin.UpdatePlan(ctx, fmt.Sprint(plan.Id), plan)
			if updateErr != nil {
				// Plan will remain as default and persist for reuse
				continue
			}
		}

		// Try to delete the plan
		err = quotaAdmin.DeletePlan(ctx, fmt.Sprint(planID))

		// "plan not found" is not an error - it was already deleted
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "not found") {
			continue
		}

		// If deletion fails due to plan being in use, log warning and continue
		// The plan will remain, but that's not fatal for cleanup
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "in use") {
			// Plan will persist for reuse
			continue
		}
	}

	// Clean up quota allowances
	for _, allowanceID := range allowanceIDs {
		err := quotaAdmin.DeleteAllowance(ctx, fmt.Sprint(allowanceID))

		// "not found" is not an error - it was already deleted or never existed
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "not found") {
			continue
		}

		// "grant not found" is also not an error - portal-specific error message
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "grant not found") {
			continue
		}

		// Other errors are logged but ignored during cleanup to allow other resources to be cleaned up
		// This matches the pattern used for plan cleanup
		if err != nil {
			// Log warning but continue with cleanup
			continue
		}
	}

	return nil
}

// getPlanSafely attempts to retrieve a quota plan, returning the plan or an error if not found.
// This helper is used to check plan status before operations like deletion.
func getPlanSafely(ctx context.Context, quotaAdmin *admin.QuotaService, planID int64) (*admin.QuotaPlan, error) {
	plan, err := quotaAdmin.GetPlan(ctx, fmt.Sprint(planID))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil, fmt.Errorf("plan %d not found (may have been deleted)", planID)
		}
		return nil, fmt.Errorf("failed to get plan %d: %w", planID, err)
	}
	return plan, nil
}

// ResetUsersForPlan finds all users assigned to a quota plan and resets them to default
func ResetUsersForPlan(ctx context.Context, quotaAdmin *admin.QuotaService, planID int64) error {
	// List all user configs to find those assigned to this plan
	userConfigs, _, err := quotaAdmin.ListUserConfigs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list user quota configs: %w", err)
	}

	for _, config := range userConfigs {
		// Check if this user is assigned to the plan we're deleting
		if config.QuotaPlanId != nil && int64(*config.QuotaPlanId) == planID {
			// Reset the user's plan assignment to NULL
			_ = quotaAdmin.ResetUserPlan(ctx, config.UserId)
		}
	}

	return nil
}

// EnsureDefaultQuotaPlan ensures a default quota plan exists and returns its ID
// This is used by billing infrastructure to ensure quota plans exist before creating pricing periods
// Returns the modified context so cleanup registration is visible to the caller.
func EnsureDefaultQuotaPlan(ctx context.Context) (context.Context, int64, error) {
	adminClient, err := RequireAdminClient(ctx)
	if err != nil {
		return ctx, 0, fmt.Errorf("failed to get admin client: %w", err)
	}

	quotaAdmin := adminClient.Quota()
	if quotaAdmin == nil {
		return ctx, 0, fmt.Errorf("admin client's Quota() returned nil")
	}

	// Check if a default plan already exists
	plans, _, err := quotaAdmin.ListPlans(ctx)
	if err != nil {
		return ctx, 0, fmt.Errorf("failed to list quota plans: %w", err)
	}

	// Look for a default plan
	for _, plan := range plans {
		if plan.IsDefault {
			// Plan exists and is default - ensure it's active
			if !plan.IsActive {
				plan.IsActive = true
				_, err := quotaAdmin.UpdatePlan(ctx, fmt.Sprint(plan.Id), plan)
				if err != nil {
					return ctx, 0, fmt.Errorf("failed to activate default quota plan %d: %w", plan.Id, err)
				}
			}
			return ctx, int64(plan.Id), nil
		}
	}

	// No default plan exists, create one
	newPlan := admin.NewQuotaPlan("Default Test Plan", "Default quota plan for billing testing", admin.QuotaLimits{
		UploadLimitBytes:   10 * 1024 * 1024 * 1024, // 10GB
		DownloadLimitBytes: 50 * 1024 * 1024 * 1024, // 50GB
		StorageLimitBytes:  1 * 1024 * 1024 * 1024,  // 1GB
		WindowDuration:     0,
		WindowStartHour:    0,
		WindowTimezone:     "",
		WindowType:         "LIFETIME",
	})

	createdPlan, err := quotaAdmin.CreatePlan(ctx, newPlan)
	if err != nil {
		return ctx, 0, fmt.Errorf("failed to create default quota plan: %w", err)
	}

	// Activate it
	createdPlan.IsActive = true
	_, err = quotaAdmin.UpdatePlan(ctx, fmt.Sprint(createdPlan.Id), createdPlan)
	if err != nil {
		return ctx, 0, fmt.Errorf("failed to activate default quota plan: %w", err)
	}

	// Set it as the default plan
	err = quotaAdmin.SetDefaultPlan(ctx, fmt.Sprint(createdPlan.Id))
	if err != nil {
		return ctx, 0, fmt.Errorf("failed to set default quota plan: %w", err)
	}

	// Add to cleanup list
	ctx = AddQuotaPlanCleanup(ctx, int64(createdPlan.Id))

	return ctx, int64(createdPlan.Id), nil
}
