package helpers

import (
	"context"
	"fmt"

	account "go.lumeweb.com/portal-sdk"
)

// PaymentGateway represents a supported payment gateway
// Currently supports 'stripe' with extensibility for future gateways
// (paypal, crypto, etc.)
type PaymentGateway string

const (
	// GatewayStripe is the Stripe payment gateway
	GatewayStripe PaymentGateway = "stripe"
	// GatewayAtlos is the Atlos crypto payment gateway
	GatewayAtlos PaymentGateway = "atlos"
	// GatewayDefault is the active default gateway
	GatewayDefault = GatewayStripe
)

type gatewayCtxKey int

const (
	gatewayKeyActiveGateway gatewayCtxKey = iota
	gatewayKeyCheckoutSessionID
	gatewayKeySubscriptionID
	gatewayKeyCheckoutAmount
	gatewayKeyCheckoutCurrency
	gatewayKeyCheckoutUI
)

// SetActiveGateway sets the active payment gateway in context
// Defaults to GatewayDefault if not set
func SetActiveGateway(ctx context.Context, gateway PaymentGateway) context.Context {
	return context.WithValue(ctx, gatewayKeyActiveGateway, string(gateway))
}

// GetActiveGateway retrieves the active payment gateway from context
// Returns GatewayDefault if not explicitly set
func GetActiveGateway(ctx context.Context) PaymentGateway {
	v, ok := ctx.Value(gatewayKeyActiveGateway).(string)
	if !ok || v == "" {
		return GatewayDefault
	}
	return PaymentGateway(v)
}

// RequireActiveGateway retrieves the active gateway or returns an error
func RequireActiveGateway(ctx context.Context) (PaymentGateway, error) {
	gateway := GetActiveGateway(ctx)
	if gateway == "" {
		return "", fmt.Errorf("no active payment gateway configured")
	}
	return gateway, nil
}

// SetActiveGatewayAsDefault ensures the active gateway is set to the default
// Call this in Before hooks to establish the default gateway
func SetActiveGatewayAsDefault(ctx context.Context) context.Context {
	return SetActiveGateway(ctx, GatewayDefault)
}

// GatewayCheckoutSessionID stores checkout session ID for the active gateway
// This is gateway-agnostic - the gateway implementation handles the specific format
func SetGatewayCheckoutSessionID(ctx context.Context, id string) context.Context {
	gateway := GetActiveGateway(ctx)
	// Store in gateway-specific context as well for direct access
	ctx = context.WithValue(ctx, gatewayKeyCheckoutSessionID, id)
	// Also store in Stripe context for backward compatibility during transition
	ctx = SetStripeCheckoutSessionID(ctx, id)
	return SetContextValue(ctx, contextKey(fmt.Sprintf("%s_checkout_session_id", gateway)), id)
}

// GetGatewayCheckoutSessionID retrieves checkout session ID for the active gateway
func GetGatewayCheckoutSessionID(ctx context.Context) (string, bool) {
	gateway := GetActiveGateway(ctx)
	key := contextKey(fmt.Sprintf("%s_checkout_session_id", gateway))
	id, ok := GetContextValue[string](ctx, key)
	if !ok {
		// Fallback to generic key
		id, ok = ctx.Value(gatewayKeyCheckoutSessionID).(string)
	}
	return id, ok
}

// SetGatewaySubscriptionID stores subscription ID for the active gateway
func SetGatewaySubscriptionID(ctx context.Context, id string) context.Context {
	gateway := GetActiveGateway(ctx)
	// Store in gateway-specific context
	ctx = context.WithValue(ctx, gatewayKeySubscriptionID, id)
	// Also store in Stripe context for backward compatibility
	ctx = SetStripeSubscriptionID(ctx, id)
	return SetContextValue(ctx, contextKey(fmt.Sprintf("%s_subscription_id", gateway)), id)
}

// GetGatewaySubscriptionID retrieves subscription ID for the active gateway
func GetGatewaySubscriptionID(ctx context.Context) (string, bool) {
	gateway := GetActiveGateway(ctx)
	key := contextKey(fmt.Sprintf("%s_subscription_id", gateway))
	id, ok := GetContextValue[string](ctx, key)
	if !ok {
		// Fallback to generic key
		id, ok = ctx.Value(gatewayKeySubscriptionID).(string)
	}
	return id, ok
}

// RequireGatewaySubscriptionID gets subscription ID or returns error
func RequireGatewaySubscriptionID(ctx context.Context) (string, error) {
	id, ok := GetGatewaySubscriptionID(ctx)
	if !ok || id == "" {
		return "", fmt.Errorf("no subscription ID available for active gateway")
	}
	return id, nil
}

// GetGatewayMock returns the appropriate gateway mock for the active gateway
func GetGatewayMock(ctx context.Context) (GatewayMock, error) {
	gateway := GetActiveGateway(ctx)
	switch gateway {
	case GatewayStripe:
		return NewStripeMock(), nil
	case GatewayAtlos:
		return NewAtlosMock(), nil
	default:
		return nil, fmt.Errorf("unsupported gateway: %s", gateway)
	}
}

// GatewayIsAtlos returns true if the active gateway is Atlos
func GatewayIsAtlos(ctx context.Context) bool {
	return GetActiveGateway(ctx) == GatewayAtlos
}

// GatewayIsStripe returns true if the active gateway is Stripe
func GatewayIsStripe(ctx context.Context) bool {
	return GetActiveGateway(ctx) == GatewayStripe
}

// GetActiveGatewaySDKName returns the gateway identifier string used by the
// portal SDK's WithGateway option. This must match the gateway constant
// registered in the portal (e.g. "stripe", "atlos").
func GetActiveGatewaySDKName(ctx context.Context) string {
	return string(GetActiveGateway(ctx))
}

// SetGatewayCheckoutAmount stores the checkout amount for the active gateway.
// This is used by gateways like Atlos that need the order amount to create an invoice.
func SetGatewayCheckoutAmount(ctx context.Context, amount float64) context.Context {
	return context.WithValue(ctx, gatewayKeyCheckoutAmount, amount)
}

// GetGatewayCheckoutAmount retrieves the checkout amount for the active gateway.
func GetGatewayCheckoutAmount(ctx context.Context) (float64, bool) {
	v, ok := ctx.Value(gatewayKeyCheckoutAmount).(float64)
	return v, ok
}

// SetGatewayCheckoutCurrency stores the checkout currency for the active gateway.
func SetGatewayCheckoutCurrency(ctx context.Context, currency string) context.Context {
	return context.WithValue(ctx, gatewayKeyCheckoutCurrency, currency)
}

// GetGatewayCheckoutCurrency retrieves the checkout currency for the active gateway.
func GetGatewayCheckoutCurrency(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(gatewayKeyCheckoutCurrency).(string)
	return v, ok
}

// SetGatewayCheckoutUI stores the checkout UI response in context.
// Each gateway mock extracts the fields it needs in CompleteCheckout.
func SetGatewayCheckoutUI(ctx context.Context, resp *account.CheckoutUI) context.Context {
	return context.WithValue(ctx, gatewayKeyCheckoutUI, resp)
}

// GetGatewayCheckoutUI retrieves the checkout UI response from context.
func GetGatewayCheckoutUI(ctx context.Context) (*account.CheckoutUI, bool) {
	v, ok := ctx.Value(gatewayKeyCheckoutUI).(*account.CheckoutUI)
	return v, ok
}
