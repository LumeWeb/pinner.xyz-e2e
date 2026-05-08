package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	account "go.lumeweb.com/portal-sdk"
	"pinner.xyz-e2e/helpers"
)

// AtlosSubscriptionSteps holds Atlos-specific step definitions
// These steps simulate the Atlos widget checkout flow via server-side API calls
// since we cannot test the JavaScript widget directly in E2E tests.
type AtlosSubscriptionSteps struct{}

// NewAtlosSubscriptionSteps creates a new AtlosSubscriptionSteps instance
func NewAtlosSubscriptionSteps() *AtlosSubscriptionSteps {
	return &AtlosSubscriptionSteps{}
}

// InitializeScenario registers all Atlos-specific step definitions
func (s *AtlosSubscriptionSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Checkout lifecycle - Atlos specific (simulates widget API calls)
	ctx.Step(`^the Atlos checkout session completes$`, s.theAtlosCheckoutSessionCompletes)
	ctx.Step(`^the Atlos subscription becomes active$`, s.theAtlosSubscriptionBecomesActive)

	// Payment simulation - Atlos specific
	ctx.Step(`^the user pays for the next billing period via Atlos$`, s.theUserPaysForTheNextBillingPeriodViaAtlos)

	// Atlos-specific subscription states
	ctx.Step(`^the user has an active Atlos subscription$`, s.theUserHasAnActiveAtlosSubscription)
	ctx.Step(`^the user has a subscription scheduled for cancellation via Atlos$`, s.theUserHasASubscriptionScheduledForCancellationViaAtlos)

	// Atlos-specific cancellation abort (equivalent to Stripe's resume after cancel)
	ctx.Step(`^the user aborts the subscription cancellation$`, s.theUserAbortsTheSubscriptionCancellation)
}

// theAtlosCheckoutSessionCompletes simulates the full Atlos widget checkout flow
// This executes the same API calls the widget would make:
// 1. InvoiceCreate → 2. AssetList → 3. PaymentCreate → 4. CompletePayment (mock)
func (s *AtlosSubscriptionSteps) theAtlosCheckoutSessionCompletes(ctx context.Context) (context.Context, error) {
	sessionID, ok := helpers.GetGatewayCheckoutSessionID(ctx)
	if !ok || sessionID == "" {
		return ctx, fmt.Errorf("no checkout session ID available")
	}

	var orderID string
	var amount float64
	var currency string

	// Parse checkout data from the stored GetCheckoutUI response fragments
	if checkoutUI, ok := helpers.GetGatewayCheckoutUI(ctx); ok {
		if data, err := helpers.ParseAtlosCheckoutFromFragments(checkoutUI.Fragments); err == nil && data.OrderID != "" {
			orderID = data.OrderID
			amount = data.Amount
			currency = data.Currency
		}
	}

	// Fallback: parse session ID and resolve amount from admin API
	if orderID == "" {
		parsed, _ := helpers.ParseAtlosOrderID(sessionID)
		orderID = parsed.Raw
		if storedAmount, ok := helpers.GetGatewayCheckoutAmount(ctx); ok && storedAmount > 0 {
			amount = storedAmount
		} else {
			amount = helpers.ResolveAtlosCheckoutAmount(ctx, orderID)
		}
		currency = "USD"
		if storedCurrency, ok := helpers.GetGatewayCheckoutCurrency(ctx); ok && storedCurrency != "" {
			currency = storedCurrency
		}
	}

	// Simulate the full Atlos checkout via API calls (not browser widget)
	// This creates invoice, selects asset, creates payment, and completes it
	subscriptionID, err := helpers.SimulateAtlosCheckout(ctx, orderID, amount, currency)
	if err != nil {
		return ctx, fmt.Errorf("atlos checkout simulation failed: %w", err)
	}

	// Store the subscription ID
	ctx = helpers.SetAtlosOrderID(ctx, subscriptionID)
	ctx = helpers.SetGatewaySubscriptionID(ctx, subscriptionID)

	// Poll for subscription to become active (postback should have been sent)
	err = helpers.PollUserSubscriptionStatus(ctx, true, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("subscription did not become active after postback: %w", err)
	}

	return ctx, nil
}

// theAtlosSubscriptionBecomesActive polls for Atlos subscription active status
// Note: For Atlos, "subscription" is managed by the portal based on payment postbacks
func (s *AtlosSubscriptionSteps) theAtlosSubscriptionBecomesActive(ctx context.Context) (context.Context, error) {
	err := helpers.PollUserSubscriptionStatus(ctx, true, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("Atlos subscription is not active: %w", err)
	}
	return ctx, nil
}

// theUserPaysForTheNextBillingPeriodViaAtlos simulates a new payment for renewal
// Since Atlos has no native subscription renewals, we create a new payment
func (s *AtlosSubscriptionSteps) theUserPaysForTheNextBillingPeriodViaAtlos(ctx context.Context) (context.Context, error) {
	subscriptionID, ok := helpers.GetGatewaySubscriptionID(ctx)
	if !ok || subscriptionID == "" {
		return ctx, fmt.Errorf("no subscription ID available")
	}

	orderID, _ := helpers.ParseAtlosOrderID(subscriptionID)

	var amount float64
	if storedAmount, ok := helpers.GetGatewayCheckoutAmount(ctx); ok && storedAmount > 0 {
		amount = storedAmount
	} else {
		amount = helpers.ResolveAtlosCheckoutAmount(ctx, orderID.Raw)
	}

	_, err := helpers.SimulateAtlosCheckout(ctx, orderID.Raw, amount, "USD")
	if err != nil {
		return ctx, fmt.Errorf("failed to simulate Atlos renewal payment: %w", err)
	}

	return ctx, nil
}

// theUserHasAnActiveAtlosSubscription ensures user has active Atlos subscription
// This sets up the full flow: create checkout → simulate widget → verify postback → active
func (s *AtlosSubscriptionSteps) theUserHasAnActiveAtlosSubscription(ctx context.Context) (context.Context, error) {
	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	// Check if already subscribed
	subscriptionStatus, err := userClient.GetSubscriptionStatus(ctx)
	if err == nil && subscriptionStatus.IsSubscribed {
		return ctx, nil
	}

	// Get billing infrastructure
	infra, _, ok := helpers.GetBillingInfrastructure(ctx)
	if !ok || infra == nil {
		return ctx, fmt.Errorf("no billing infrastructure available")
	}

	planID := infra.FirstPlanID()
	periodID := infra.FirstPeriodID()

	// Get checkout UI - this returns the order ID as session ID
	checkoutUI, err := userClient.GetCheckoutUI(ctx, fmt.Sprint(planID),
		account.WithPeriodID(fmt.Sprint(periodID)),
		account.WithGateway(helpers.GetActiveGatewaySDKName(ctx)),
	)
	if err != nil {
		return ctx, fmt.Errorf("failed to get checkout UI: %w", err)
	}

	ctx = helpers.SetGatewayCheckoutUI(ctx, checkoutUI)

	if checkoutUI.SessionId != nil {
		// The SessionId for Atlos is the order ID
		ctx = helpers.SetGatewayCheckoutSessionID(ctx, *checkoutUI.SessionId)
	}

	// Complete the checkout (simulates widget API calls)
	return s.theAtlosCheckoutSessionCompletes(ctx)
}

// theUserHasASubscriptionScheduledForCancellationViaAtlos cancels at period end via portal API
// Atlos doesn't have native cancel_at_period_end — the portal tracks this internally
func (s *AtlosSubscriptionSteps) theUserHasASubscriptionScheduledForCancellationViaAtlos(ctx context.Context) (context.Context, error) {
	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	_, err = userClient.CancelSubscription(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	subscriptionID, err := helpers.RequireGatewaySubscriptionID(ctx)
	if err != nil {
		return ctx, err
	}

	if err := helpers.SetGatewayCancelAtPeriodEnd(ctx, subscriptionID); err != nil {
		return ctx, fmt.Errorf("failed to schedule cancellation: %w", err)
	}

	return ctx, nil
}

// theUserAbortsTheSubscriptionCancellation reverses a scheduled cancellation via portal API
// For Atlos, this calls AbortSubscriptionCancellation which clears WillCancelAt
func (s *AtlosSubscriptionSteps) theUserAbortsTheSubscriptionCancellation(ctx context.Context) (context.Context, error) {
	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	_, err = userClient.AbortSubscriptionCancellation(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to abort subscription cancellation: %w", err)
	}

	return ctx, nil
}
