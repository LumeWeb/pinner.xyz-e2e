package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

// GatewaySteps holds step definitions for gateway configuration
type GatewaySteps struct{}

// NewGatewaySteps creates a new GatewaySteps instance
func NewGatewaySteps() *GatewaySteps {
	return &GatewaySteps{}
}

// InitializeScenario registers all step definitions with godog
func (s *GatewaySteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the payment mock is reset$`, s.thePaymentMockIsReset)
	ctx.Step(`^the ([^"]+) gateway is active$`, s.theGatewayIsActive)
	ctx.Step(`^the default gateway is set$`, s.theDefaultGatewayIsSet)
}

// thePaymentMockIsReset resets the active gateway mock
// This is the gateway-agnostic version of stripe mock reset
func (s *GatewaySteps) thePaymentMockIsReset(ctx context.Context) (context.Context, error) {
	if err := helpers.ResetPaymentGateway(ctx); err != nil {
		return ctx, fmt.Errorf("failed to reset payment gateway: %w", err)
	}
	return ctx, nil
}

// theGatewayIsActive sets the active payment gateway
func (s *GatewaySteps) theGatewayIsActive(ctx context.Context, gatewayName string) (context.Context, error) {
	gateway := helpers.PaymentGateway(gatewayName)

	switch gateway {
	case helpers.GatewayStripe:
		ctx = helpers.SetActiveGateway(ctx, gateway)
	case helpers.GatewayAtlos:
		ctx = helpers.SetActiveGateway(ctx, gateway)
	default:
		return ctx, fmt.Errorf("unsupported gateway: %s", gatewayName)
	}

	return ctx, nil
}

// theDefaultGatewayIsSet ensures the default gateway is active
func (s *GatewaySteps) theDefaultGatewayIsSet(ctx context.Context) (context.Context, error) {
	ctx = helpers.SetActiveGatewayAsDefault(ctx)
	return ctx, nil
}
