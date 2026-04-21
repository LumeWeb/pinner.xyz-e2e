package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

// StripeManualControlSteps holds step definitions for manual Stripe subscription control
// These steps provide direct manipulation of Stripe subscriptions via stripe-mock
// for testing Stripe-specific lifecycle events.
type StripeManualControlSteps struct{}

// NewStripeManualControlSteps creates a new StripeManualControlSteps instance
func NewStripeManualControlSteps() *StripeManualControlSteps {
	return &StripeManualControlSteps{}
}

// InitializeScenario registers all step definitions with godog
func (s *StripeManualControlSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Legacy step names (kept for backward compatibility)
	ctx.Step(`^the subscription is renewed$`, s.theStripeSubscriptionIsRenewed)
	ctx.Step(`^the subscription is expired$`, s.theStripeSubscriptionIsExpired)
	ctx.Step(`^the subscription is canceled at period end$`, s.theStripeSubscriptionIsCanceledAtPeriodEnd)
	ctx.Step(`^the subscription is paused$`, s.theStripeSubscriptionIsPaused)
	ctx.Step(`^the subscription is resumed$`, s.theStripeSubscriptionIsResumed)

	// Explicitly Stripe-prefixed step names (preferred)
	ctx.Step(`^the Stripe subscription is renewed$`, s.theStripeSubscriptionIsRenewed)
	ctx.Step(`^the Stripe subscription is expired$`, s.theStripeSubscriptionIsExpired)
	ctx.Step(`^the Stripe subscription is canceled at period end$`, s.theStripeSubscriptionIsCanceledAtPeriodEnd)
	ctx.Step(`^the Stripe subscription is paused$`, s.theStripeSubscriptionIsPaused)
	ctx.Step(`^the Stripe subscription is resumed$`, s.theStripeSubscriptionIsResumed)
}

// theStripeSubscriptionIsRenewed manually triggers Stripe subscription renewal
func (s *StripeManualControlSteps) theStripeSubscriptionIsRenewed(ctx context.Context) (context.Context, error) {
	subscriptionID, err := helpers.RequireGatewaySubscriptionID(ctx)
	if err != nil {
		return ctx, err
	}

	if err := helpers.RenewGatewaySubscription(ctx, subscriptionID); err != nil {
		return ctx, fmt.Errorf("failed to renew Stripe subscription: %w", err)
	}

	return ctx, nil
}

// theStripeSubscriptionIsExpired manually triggers Stripe subscription expiration
func (s *StripeManualControlSteps) theStripeSubscriptionIsExpired(ctx context.Context) (context.Context, error) {
	subscriptionID, err := helpers.RequireGatewaySubscriptionID(ctx)
	if err != nil {
		return ctx, err
	}

	if err := helpers.ExpireStripeSubscription(ctx, subscriptionID); err != nil {
		return ctx, fmt.Errorf("failed to expire Stripe subscription: %w", err)
	}

	return ctx, nil
}

// theStripeSubscriptionIsCanceledAtPeriodEnd manually triggers cancellation at period end
func (s *StripeManualControlSteps) theStripeSubscriptionIsCanceledAtPeriodEnd(ctx context.Context) (context.Context, error) {
	subscriptionID, err := helpers.RequireGatewaySubscriptionID(ctx)
	if err != nil {
		return ctx, err
	}

	sub, err := helpers.GetStripeSubscription(ctx, subscriptionID)
	if err != nil {
		return ctx, fmt.Errorf("failed to get Stripe subscription: %w", err)
	}
	if !sub.CancelAtPeriodEnd {
		return ctx, fmt.Errorf("subscription not scheduled for cancellation, cannot cancel at period end")
	}

	if err := helpers.ExpireStripeSubscription(ctx, subscriptionID); err != nil {
		return ctx, fmt.Errorf("failed to cancel Stripe subscription at period end: %w", err)
	}

	time.Sleep(5 * time.Second)

	return ctx, nil
}

// theStripeSubscriptionIsPaused pauses via Stripe mock
func (s *StripeManualControlSteps) theStripeSubscriptionIsPaused(ctx context.Context) (context.Context, error) {
	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	_, err = userClient.PauseBilling(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to initiate pause via user API: %w", err)
	}

	subscriptionID, err := helpers.RequireGatewaySubscriptionID(ctx)
	if err != nil {
		return ctx, err
	}

	if err := helpers.PauseGatewaySubscription(ctx, subscriptionID); err != nil {
		return ctx, fmt.Errorf("failed to trigger Stripe pause: %w", err)
	}

	return ctx, nil
}

// theStripeSubscriptionIsResumed resumes via Stripe mock
func (s *StripeManualControlSteps) theStripeSubscriptionIsResumed(ctx context.Context) (context.Context, error) {
	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	_, err = userClient.ResumeBilling(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to initiate resume via user API: %w", err)
	}

	subscriptionID, err := helpers.RequireGatewaySubscriptionID(ctx)
	if err != nil {
		return ctx, err
	}

	if err := helpers.ResumeGatewaySubscription(ctx, subscriptionID); err != nil {
		return ctx, fmt.Errorf("failed to trigger Stripe resume: %w", err)
	}

	return ctx, nil
}
