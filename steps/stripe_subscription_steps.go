package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	account "go.lumeweb.com/portal-sdk"
	"pinner.xyz-e2e/helpers"
)

// StripeSubscriptionSteps holds Stripe-specific step definitions
// These steps explicitly depend on Stripe-specific functionality and
// are kept separate from gateway-agnostic steps to allow testing of
// Stripe-specific behaviors like payment simulation, cancel_at_period_end, etc.
type StripeSubscriptionSteps struct{}

// NewStripeSubscriptionSteps creates a new StripeSubscriptionSteps instance
func NewStripeSubscriptionSteps() *StripeSubscriptionSteps {
	return &StripeSubscriptionSteps{}
}

// InitializeScenario registers all Stripe-specific step definitions
func (s *StripeSubscriptionSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Checkout lifecycle - Stripe specific
	ctx.Step(`^the Stripe checkout session completes$`, s.theStripeCheckoutSessionCompletes)
	ctx.Step(`^the Stripe subscription becomes active$`, s.theStripeSubscriptionBecomesActive)

	// Payment simulation - Stripe specific
	ctx.Step(`^the user pays for the next billing period via Stripe$`, s.theUserPaysForTheNextBillingPeriodViaStripe)

	// Stripe-specific subscription states
	ctx.Step(`^the user has an active Stripe subscription$`, s.theUserHasAnActiveStripeSubscription)
	ctx.Step(`^the user has a paused Stripe subscription$`, s.theUserHasAPausedStripeSubscription)
	ctx.Step(`^the user has a subscription scheduled for cancellation via Stripe$`, s.theUserHasASubscriptionScheduledForCancellationViaStripe)
}

// theStripeCheckoutSessionCompletes completes checkout specifically via Stripe mock
func (s *StripeSubscriptionSteps) theStripeCheckoutSessionCompletes(ctx context.Context) (context.Context, error) {
	sessionID, ok := helpers.GetStripeCheckoutSessionID(ctx)
	if !ok || sessionID == "" {
		return ctx, fmt.Errorf("no checkout session ID available")
	}

	if err := helpers.CompleteStripeCheckoutSession(ctx, sessionID); err != nil {
		return ctx, fmt.Errorf("failed to complete checkout session: %w", err)
	}

	checkoutSession, err := helpers.GetStripeCheckoutSession(ctx, sessionID)
	if err != nil {
		return ctx, fmt.Errorf("failed to get checkout session after completion: %w", err)
	}

	if checkoutSession.Subscription != nil && checkoutSession.Subscription.ID != "" {
		ctx = helpers.SetStripeSubscriptionID(ctx, checkoutSession.Subscription.ID)
		ctx = helpers.SetGatewaySubscriptionID(ctx, checkoutSession.Subscription.ID)
	}

	err = helpers.PollUserSubscriptionStatus(ctx, true, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("subscription did not become active: %w", err)
	}

	return ctx, nil
}

// theStripeSubscriptionBecomesActive polls for Stripe subscription active status
func (s *StripeSubscriptionSteps) theStripeSubscriptionBecomesActive(ctx context.Context) (context.Context, error) {
	err := helpers.PollUserSubscriptionStatus(ctx, true, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("Stripe subscription is not active: %w", err)
	}
	return ctx, nil
}

// theUserPaysForTheNextBillingPeriodViaStripe simulates Stripe payment success
func (s *StripeSubscriptionSteps) theUserPaysForTheNextBillingPeriodViaStripe(ctx context.Context) (context.Context, error) {
	subscriptionID, ok := helpers.GetStripeSubscriptionID(ctx)
	if !ok || subscriptionID == "" {
		return ctx, fmt.Errorf("no subscription ID available")
	}

	if err := helpers.SimulateStripePaymentSuccess(ctx, subscriptionID); err != nil {
		return ctx, fmt.Errorf("failed to simulate Stripe payment: %w", err)
	}

	return ctx, nil
}

// theUserHasAnActiveStripeSubscription ensures user has active Stripe subscription
func (s *StripeSubscriptionSteps) theUserHasAnActiveStripeSubscription(ctx context.Context) (context.Context, error) {
	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	subscriptionStatus, err := userClient.GetSubscriptionStatus(ctx)
	if err == nil && subscriptionStatus.IsSubscribed {
		return ctx, nil
	}

	infra, _, ok := helpers.GetBillingInfrastructure(ctx)
	if !ok || infra == nil {
		return ctx, fmt.Errorf("no billing infrastructure available")
	}

	planID := infra.FirstPlanID()
	periodID := infra.FirstPeriodID()

	checkoutUI, err := userClient.GetCheckoutUI(ctx, fmt.Sprint(planID),
		account.WithPeriodID(fmt.Sprint(periodID)),
		account.WithGateway(helpers.GetActiveGatewaySDKName(ctx)),
	)
	if err != nil {
		return ctx, fmt.Errorf("failed to get checkout UI: %w", err)
	}

	ctx = helpers.SetGatewayCheckoutUI(ctx, checkoutUI)

	if checkoutUI.SessionId != nil {
		ctx = helpers.SetStripeCheckoutSessionID(ctx, *checkoutUI.SessionId)
		ctx = helpers.SetGatewayCheckoutSessionID(ctx, *checkoutUI.SessionId)
	}

	return s.theStripeCheckoutSessionCompletes(ctx)
}

// theUserHasAPausedStripeSubscription creates a paused subscription via Stripe
func (s *StripeSubscriptionSteps) theUserHasAPausedStripeSubscription(ctx context.Context) (context.Context, error) {
	ctx, err := s.theUserHasAnActiveStripeSubscription(ctx)
	if err != nil {
		return ctx, err
	}

	subscriptionID, ok := helpers.GetStripeSubscriptionID(ctx)
	if !ok || subscriptionID == "" {
		return ctx, fmt.Errorf("no subscription ID available")
	}

	if err := helpers.PauseStripeSubscription(ctx, subscriptionID); err != nil {
		return ctx, fmt.Errorf("failed to pause subscription: %w", err)
	}

	return ctx, nil
}

// theUserHasASubscriptionScheduledForCancellationViaStripe cancels at period end via Stripe
func (s *StripeSubscriptionSteps) theUserHasASubscriptionScheduledForCancellationViaStripe(ctx context.Context) (context.Context, error) {
	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	_, err = userClient.CancelSubscription(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	subscriptionID, ok := helpers.GetStripeSubscriptionID(ctx)
	if !ok || subscriptionID == "" {
		return ctx, fmt.Errorf("no subscription ID available")
	}

	if err := helpers.SetStripeCancelAtPeriodEnd(ctx, subscriptionID); err != nil {
		return ctx, fmt.Errorf("failed to schedule cancellation: %w", err)
	}

	return ctx, nil
}
