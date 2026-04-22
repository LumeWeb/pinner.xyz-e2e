package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	account "go.lumeweb.com/portal-sdk"
	"pinner.xyz-e2e/helpers"
)

// SubscriptionCommonSteps holds step definitions for common subscription operations
// These steps are gateway-agnostic and work with any configured payment gateway
type SubscriptionCommonSteps struct{}

// NewSubscriptionCommonSteps creates a new SubscriptionCommonSteps instance
func NewSubscriptionCommonSteps() *SubscriptionCommonSteps {
	return &SubscriptionCommonSteps{}
}

// InitializeScenario registers all step definitions with godog
func (s *SubscriptionCommonSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Legacy step - redirect to stripe-specific
	ctx.Step(`^the stripe mock is reset$`, s.theStripeMockIsResetLegacy)

	// Gateway-agnostic billing infrastructure
	ctx.Step(`^the billing infrastructure is set up$`, s.theBillingInfrastructureIsSetUp)

	// Test user with active subscription (gateway-agnostic via mocks)
	ctx.Step(`^the test user has an active subscription$`, s.theTestUserHasAnActiveSubscription)

	// Legacy step - redirect to gateway-agnostic
	ctx.Step(`^the Stripe subscription becomes active$`, s.theStripeSubscriptionBecomesActiveLegacy)
}

// theStripeMockIsResetLegacy provides backward compatibility
// dispatches to gateway-agnostic payment mock reset
func (s *SubscriptionCommonSteps) theStripeMockIsResetLegacy(ctx context.Context) (context.Context, error) {
	return s.resetActiveGateway(ctx)
}

func (s *SubscriptionCommonSteps) resetActiveGateway(ctx context.Context) (context.Context, error) {
	if err := helpers.ResetPaymentGateway(ctx); err != nil {
		return ctx, fmt.Errorf("failed to reset payment gateway: %w", err)
	}
	return ctx, nil
}

// theBillingInfrastructureIsSetUp sets up billing infrastructure via portal admin API
// This is gateway-agnostic as it uses the portal API, not gateway-specific APIs
func (s *SubscriptionCommonSteps) theBillingInfrastructureIsSetUp(ctx context.Context) (context.Context, error) {
	ctx, err := helpers.SetupBillingInfrastructure(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to setup billing infrastructure: %w", err)
	}
	return ctx, nil
}

// theTestUserHasAnActiveSubscription creates an active subscription for the test user
// Uses gateway-agnostic mock completion
func (s *SubscriptionCommonSteps) theTestUserHasAnActiveSubscription(ctx context.Context) (context.Context, error) {
	// Get billing infrastructure
	infra, _, ok := helpers.GetBillingInfrastructure(ctx)
	if !ok || infra == nil {
		return ctx, fmt.Errorf("no billing infrastructure available")
	}

	// Get pricing plan ID and period ID from infrastructure
	planID := infra.FirstPlanID()
	periodID := infra.FirstPeriodID()
	if planID == 0 {
		return ctx, fmt.Errorf("no pricing plan found in billing infrastructure")
	}
	if periodID == 0 {
		return ctx, fmt.Errorf("no pricing plan period found in billing infrastructure")
	}

	// Get authenticated user client
	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	// Create checkout session using plan ID and period ID, specifying the active gateway
	checkoutUI, err := userClient.GetCheckoutUI(ctx, fmt.Sprint(planID),
		account.WithPeriodID(fmt.Sprint(periodID)),
		account.WithGateway(helpers.GetActiveGatewaySDKName(ctx)),
	)
	if err != nil {
		return ctx, fmt.Errorf("failed to get checkout UI: %w", err)
	}

	ctx = helpers.SetGatewayCheckoutUI(ctx, checkoutUI)

	if checkoutUI.SessionId == nil || *checkoutUI.SessionId == "" {
		return ctx, fmt.Errorf("checkout UI did not return a session ID")
	}

	sessionID := *checkoutUI.SessionId
	ctx = helpers.SetGatewayCheckoutSessionID(ctx, sessionID)

	// Use gateway-agnostic completion
	subscriptionID, err := helpers.CompleteGatewayCheckout(ctx, sessionID)
	if err != nil {
		return ctx, fmt.Errorf("failed to complete checkout: %w", err)
	}

	if subscriptionID != "" {
		ctx = helpers.SetGatewaySubscriptionID(ctx, subscriptionID)
	}

	// Wait for subscription to be active
	err = helpers.PollUserSubscriptionStatus(ctx, true, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("subscription did not become active: %w", err)
	}

	return ctx, nil
}

// theStripeSubscriptionBecomesActiveLegacy ensures backward compatibility
// Internally delegates to gateway-agnostic subscription status check
func (s *SubscriptionCommonSteps) theStripeSubscriptionBecomesActiveLegacy(ctx context.Context) (context.Context, error) {
	return s.theSubscriptionBecomesActive(ctx)
}

// theSubscriptionBecomesActive polls portal API for active subscription status
func (s *SubscriptionCommonSteps) theSubscriptionBecomesActive(ctx context.Context) (context.Context, error) {
	err := helpers.PollUserSubscriptionStatus(ctx, true, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("subscription is not active: %w", err)
	}
	return ctx, nil
}
