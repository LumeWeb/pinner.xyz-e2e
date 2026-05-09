package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

// AtlosManualControlSteps holds Atlos-specific manual control step definitions
// These steps provide direct manipulation of Atlos payment lifecycle for testing.
// Note: Atlos doesn't support native pause/resume operations.
// Renewals are handled by creating new payments.
// Expirations are managed by the portal when no payment is received.
type AtlosManualControlSteps struct{}

// NewAtlosManualControlSteps creates a new AtlosManualControlSteps instance
func NewAtlosManualControlSteps() *AtlosManualControlSteps {
	return &AtlosManualControlSteps{}
}

// InitializeScenario registers all Atlos manual control step definitions
func (s *AtlosManualControlSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Atlos-prefixed step names (consistent with Stripe pattern)
	ctx.Step(`^the Atlos subscription is renewed$`, s.theAtlosSubscriptionIsRenewed)
	ctx.Step(`^the Atlos subscription is expired$`, s.theAtlosSubscriptionIsExpired)

}

// theAtlosSubscriptionIsRenewed triggers renewal by creating a new payment
func (s *AtlosManualControlSteps) theAtlosSubscriptionIsRenewed(ctx context.Context) (context.Context, error) {
	subscriptionID, err := helpers.RequireGatewaySubscriptionID(ctx)
	if err != nil {
		return ctx, err
	}

	orderID, _ := helpers.ParseAtlosOrderID(subscriptionID)

	var amount float64
	if storedAmount, ok := helpers.GetGatewayCheckoutAmount(ctx); ok && storedAmount > 0 {
		amount = storedAmount
	} else {
		amount = helpers.ResolveAtlosCheckoutAmount(ctx, orderID.Raw)
	}

	_, err = helpers.SimulateAtlosCheckout(ctx, orderID.Raw, amount, "USD")
	if err != nil {
		return ctx, fmt.Errorf("failed to simulate Atlos renewal: %w", err)
	}

	return ctx, nil
}

// theAtlosSubscriptionIsExpired marks the subscription as expired
// Since Atlos doesn't have native expiration, we rely on the portal's
// subscription management to handle this
func (s *AtlosManualControlSteps) theAtlosSubscriptionIsExpired(ctx context.Context) (context.Context, error) {
	subscriptionID, err := helpers.RequireGatewaySubscriptionID(ctx)
	if err != nil {
		return ctx, err
	}

	// Use gateway-agnostic expire - the AtlosMock implementation
	// handles the Atlos-specific behavior (no-op, portal manages expiration)
	if err := helpers.ExpireGatewaySubscription(ctx, subscriptionID); err != nil {
		return ctx, fmt.Errorf("failed to expire Atlos subscription: %w", err)
	}

	return ctx, nil
}


