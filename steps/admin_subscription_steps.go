package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	admin "go.lumeweb.com/portal-sdk/admin"
	"pinner.xyz-e2e/helpers"
)

// AdminSubscriptionSteps holds step definitions for admin subscription management
type AdminSubscriptionSteps struct{}

// NewAdminSubscriptionSteps creates a new AdminSubscriptionSteps instance
func NewAdminSubscriptionSteps() *AdminSubscriptionSteps {
	return &AdminSubscriptionSteps{}
}

// InitializeScenario registers all step definitions with godog
func (s *AdminSubscriptionSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the admin cancels the subscription immediately$`, s.theAdminCancelsTheSubscriptionImmediately)
	ctx.Step(`^the admin cancels the subscription at end of period$`, s.theAdminCancelsTheSubscriptionAtEndOfPeriod)
	ctx.Step(`^the admin changes the user's plan$`, s.theAdminChangesTheUserPlan)

	// Abort cancellation steps
	ctx.Step(`^there is a subscription scheduled for cancellation at period end$`, s.thereIsASubscriptionScheduledForCancellationAtPeriodEnd)
	ctx.Step(`^the admin aborts the subscription cancellation$`, s.theAdminAbortsTheSubscriptionCancellation)
	ctx.Step(`^the scheduled cancellation is aborted successfully$`, s.theScheduledCancellationIsAbortedSuccessfully)

	// Pause subscription steps
	ctx.Step(`^the admin pauses the user's subscription$`, s.theAdminPausesTheUsersSubscription)
	ctx.Step(`^the pause operation is successful$`, s.thePauseOperationIsSuccessful)

	// Resume subscription steps
	ctx.Step(`^the user's subscription is paused$`, s.theUsersSubscriptionIsPaused)
	ctx.Step(`^the admin resumes the user's subscription$`, s.theAdminResumesTheUsersSubscription)
	ctx.Step(`^the resume operation is successful$`, s.theResumeOperationIsSuccessful)

	// Common verification
	ctx.Step(`^the subscription status is "([^"]*)"$`, s.theSubscriptionStatusIs)
}

// theAdminCancelsTheSubscriptionImmediately cancels the user's subscription immediately
func (s *AdminSubscriptionSteps) theAdminCancelsTheSubscriptionImmediately(ctx context.Context) (context.Context, error) {
	// Get user ID from context
	userID, err := helpers.RequireAdminTargetUserID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no target user ID available: %w", err)
	}

	// Get admin client
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get admin client: %w", err)
	}

	// Cancel user's subscription immediately via admin API
	immediate := true
	result, err := adminClient.Billing().CancelUserSubscription(ctx, fmt.Sprint(userID), &admin.CancelSubscriptionRequest{
		Immediate: &immediate,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	ctx = helpers.SetAdminManagementResult(ctx, result)

	// Poll for subscription status - should be inactive
	err = helpers.PollSubscriptionStatus(ctx, userID, false, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("subscription was not canceled: %w", err)
	}

	return ctx, nil
}

// theAdminCancelsTheSubscriptionAtEndOfPeriod cancels the user's subscription at end of period
func (s *AdminSubscriptionSteps) theAdminCancelsTheSubscriptionAtEndOfPeriod(ctx context.Context) (context.Context, error) {
	// Get user ID from context
	userID, err := helpers.RequireAdminTargetUserID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no target user ID available: %w", err)
	}

	// Get admin client
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get admin client: %w", err)
	}

	// Cancel user's subscription at period end via admin API
	result, err := adminClient.Billing().CancelUserSubscription(ctx, fmt.Sprint(userID), &admin.CancelSubscriptionRequest{})
	if err != nil {
		return ctx, fmt.Errorf("failed to schedule subscription cancellation: %w", err)
	}

	ctx = helpers.SetAdminManagementResult(ctx, result)

	// Poll for subscription status - should still be active
	err = helpers.PollSubscriptionStatus(ctx, userID, true, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("failed to verify scheduled cancellation: %w", err)
	}

	return ctx, nil
}

// theAdminChangesTheUserPlan changes the user's subscription plan
func (s *AdminSubscriptionSteps) theAdminChangesTheUserPlan(ctx context.Context) (context.Context, error) {
	// Get user ID from context
	userID, err := helpers.RequireAdminTargetUserID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no target user ID available: %w", err)
	}

	// Get billing infrastructure for new plan
	infra, _, ok := helpers.GetBillingInfrastructure(ctx)
	if !ok || infra == nil {
		return ctx, fmt.Errorf("no billing infrastructure available")
	}

	// Get admin client
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get admin client: %w", err)
	}

	// Update user's subscription with new pricing plan via admin API
	_, err = adminClient.Billing().ChangeUserPlan(ctx, fmt.Sprint(userID), &admin.ChangePlanRequest{
		PeriodId: int(infra.FirstPeriodID()),
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to change user's plan: %w", err)
	}

	// Poll for updated subscription status - should be active
	err = helpers.PollSubscriptionStatus(ctx, userID, true, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("failed to verify plan change: %w", err)
	}

	return ctx, nil
}

// =============================================================================
// Abort Cancellation Steps
// =============================================================================

// thereIsASubscriptionScheduledForCancellationAtPeriodEnd sets up a scheduled cancellation
func (s *AdminSubscriptionSteps) thereIsASubscriptionScheduledForCancellationAtPeriodEnd(ctx context.Context) (context.Context, error) {
	// Get user ID from context
	userID, err := helpers.RequireAdminTargetUserID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no target user ID available: %w", err)
	}

	// Get admin client
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get admin client: %w", err)
	}

	// Schedule cancellation at period end
	_, err = adminClient.Billing().CancelUserSubscription(ctx, fmt.Sprint(userID), &admin.CancelSubscriptionRequest{})
	if err != nil {
		return ctx, fmt.Errorf("failed to schedule subscription cancellation: %w", err)
	}

	// Give webhook time to process the scheduled cancellation before attempting to abort
	time.Sleep(5 * time.Second)

	// Poll for subscription status - should still be active
	err = helpers.PollSubscriptionStatus(ctx, userID, true, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("failed to verify scheduled cancellation: %w", err)
	}

	return ctx, nil
}

// theAdminAbortsTheSubscriptionCancellation aborts a scheduled cancellation
func (s *AdminSubscriptionSteps) theAdminAbortsTheSubscriptionCancellation(ctx context.Context) (context.Context, error) {
	// Get user ID from context
	userID, err := helpers.RequireAdminTargetUserID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no target user ID available: %w", err)
	}

	// Get admin client
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get admin client: %w", err)
	}

	// Abort the scheduled cancellation
	result, err := adminClient.Billing().AbortUserSubscriptionCancellation(ctx, fmt.Sprint(userID))
	if err != nil {
		return ctx, fmt.Errorf("failed to abort subscription cancellation: %w", err)
	}

	ctx = helpers.SetAdminManagementResult(ctx, result)
	return ctx, nil
}

// theScheduledCancellationIsAbortedSuccessfully verifies the abort was successful
func (s *AdminSubscriptionSteps) theScheduledCancellationIsAbortedSuccessfully(ctx context.Context) (context.Context, error) {
	_, err := helpers.RequireAdminManagementResult(ctx)
	if err != nil {
		return ctx, fmt.Errorf("abort operation result not found: %w", err)
	}
	return ctx, nil
}

// theSubscriptionStatusIs verifies the subscription status matches expected
func (s *AdminSubscriptionSteps) theSubscriptionStatusIs(ctx context.Context, status string) (context.Context, error) {
	return ctx, helpers.VerifySubscriptionStatus(ctx, status)
}

// =============================================================================
// Pause Subscription Steps
// =============================================================================

// theAdminPausesTheUsersSubscription pauses the user's subscription
func (s *AdminSubscriptionSteps) theAdminPausesTheUsersSubscription(ctx context.Context) (context.Context, error) {
	// Get user ID from context
	userID, err := helpers.RequireAdminTargetUserID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no target user ID available: %w", err)
	}

	// Get admin client
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get admin client: %w", err)
	}

	// Pause the subscription
	result, err := adminClient.Billing().PauseUserSubscription(ctx, fmt.Sprint(userID))
	if err != nil {
		return ctx, fmt.Errorf("failed to pause subscription: %w", err)
	}

	ctx = helpers.SetAdminManagementResult(ctx, result)

	// Poll for subscription status - should be paused (is_active=false, paused_at set)
	err = helpers.PollSubscriptionStatus(ctx, userID, false, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("subscription was not paused: %w", err)
	}

	return ctx, nil
}

// thePauseOperationIsSuccessful verifies the pause operation was successful
func (s *AdminSubscriptionSteps) thePauseOperationIsSuccessful(ctx context.Context) (context.Context, error) {
	_, err := helpers.RequireAdminManagementResult(ctx)
	if err != nil {
		return ctx, fmt.Errorf("pause operation result not found: %w", err)
	}
	return ctx, nil
}

// =============================================================================
// Resume Subscription Steps
// =============================================================================

// theUsersSubscriptionIsPaused sets up a paused subscription
func (s *AdminSubscriptionSteps) theUsersSubscriptionIsPaused(ctx context.Context) (context.Context, error) {
	return s.theAdminPausesTheUsersSubscription(ctx)
}

// theAdminResumesTheUsersSubscription resumes the user's subscription
func (s *AdminSubscriptionSteps) theAdminResumesTheUsersSubscription(ctx context.Context) (context.Context, error) {
	// Get user ID from context
	userID, err := helpers.RequireAdminTargetUserID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no target user ID available: %w", err)
	}

	// Get admin client
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get admin client: %w", err)
	}

	// Resume the subscription
	result, err := adminClient.Billing().ResumeUserSubscription(ctx, fmt.Sprint(userID))
	if err != nil {
		return ctx, fmt.Errorf("failed to resume subscription: %w", err)
	}

	ctx = helpers.SetAdminManagementResult(ctx, result)

	// Poll for subscription status - should be active
	err = helpers.PollSubscriptionStatus(ctx, userID, true, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("subscription was not resumed: %w", err)
	}

	return ctx, nil
}

// theResumeOperationIsSuccessful verifies the resume operation was successful
func (s *AdminSubscriptionSteps) theResumeOperationIsSuccessful(ctx context.Context) (context.Context, error) {
	_, err := helpers.RequireAdminManagementResult(ctx)
	if err != nil {
		return ctx, fmt.Errorf("resume operation result not found: %w", err)
	}
	return ctx, nil
}
