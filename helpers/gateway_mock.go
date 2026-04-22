package helpers

import (
	"context"
	"fmt"
)

// GatewayMock defines the interface for payment gateway mock implementations
// This abstraction allows tests to run against any payment gateway
// with gateway-specific implementations handling the unique quirks of each.
type GatewayMock interface {
	// Reset resets the mock gateway state (clears all data)
	Reset(ctx context.Context) error

	// CompleteCheckout completes a checkout session and creates a subscription
	// Returns the subscription ID
	CompleteCheckout(ctx context.Context, sessionID string) (string, error)

	// GetCheckoutSession retrieves checkout session details
	GetCheckoutSession(ctx context.Context, sessionID string) (*CheckoutSession, error)

	// SimulatePayment simulates a successful payment for a subscription
	SimulatePayment(ctx context.Context, subscriptionID string) error

	// SetCancelAtPeriodEnd schedules cancellation at period end
	SetCancelAtPeriodEnd(ctx context.Context, subscriptionID string) error

	// PauseSubscription pauses an active subscription
	PauseSubscription(ctx context.Context, subscriptionID string) error

	// ResumeSubscription resumes a paused subscription
	ResumeSubscription(ctx context.Context, subscriptionID string) error

	// RenewSubscription renews/extends a subscription
	RenewSubscription(ctx context.Context, subscriptionID string) error

	// ExpireSubscription expires/cancels a subscription immediately
	ExpireSubscription(ctx context.Context, subscriptionID string) error
}

// CheckoutSession represents a gateway-agnostic checkout session
// Gateway-specific implementations map their internal types to this
type CheckoutSession struct {
	ID             string
	SubscriptionID string
	Status         string
	PaymentStatus  string
}

// StripeMock implements GatewayMock for Stripe via stripe-mock
type StripeMock struct{}

// NewStripeMock creates a new Stripe mock instance
func NewStripeMock() *StripeMock {
	return &StripeMock{}
}

// Reset resets stripe-mock state
func (m *StripeMock) Reset(ctx context.Context) error {
	return ResetStripeMock(ctx)
}

// CompleteCheckout completes a Stripe checkout session
func (m *StripeMock) CompleteCheckout(ctx context.Context, sessionID string) (string, error) {
	if err := CompleteStripeCheckoutSession(ctx, sessionID); err != nil {
		return "", err
	}

	session, err := GetStripeCheckoutSession(ctx, sessionID)
	if err != nil {
		return "", err
	}

	if session.Subscription != nil {
		return session.Subscription.ID, nil
	}
	return "", nil
}

// GetCheckoutSession retrieves Stripe checkout session details
func (m *StripeMock) GetCheckoutSession(ctx context.Context, sessionID string) (*CheckoutSession, error) {
	session, err := GetStripeCheckoutSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	result := &CheckoutSession{
		ID:            session.ID,
		Status:        string(session.Status),
		PaymentStatus: string(session.PaymentStatus),
	}

	if session.Subscription != nil {
		result.SubscriptionID = session.Subscription.ID
	}

	return result, nil
}

// SimulatePayment simulates a successful Stripe payment by renewing the subscription
func (m *StripeMock) SimulatePayment(ctx context.Context, subscriptionID string) error {
	return SimulateStripePaymentSuccess(ctx, subscriptionID)
}

// SetCancelAtPeriodEnd sets cancel_at_period_end via Stripe API
func (m *StripeMock) SetCancelAtPeriodEnd(ctx context.Context, subscriptionID string) error {
	return SetStripeCancelAtPeriodEnd(ctx, subscriptionID)
}

// PauseSubscription pauses via Stripe subscription action
func (m *StripeMock) PauseSubscription(ctx context.Context, subscriptionID string) error {
	return PauseStripeSubscription(ctx, subscriptionID)
}

// ResumeSubscription resumes via Stripe subscription action
func (m *StripeMock) ResumeSubscription(ctx context.Context, subscriptionID string) error {
	return ResumeStripeSubscription(ctx, subscriptionID)
}

// RenewSubscription renews via Stripe renew action
func (m *StripeMock) RenewSubscription(ctx context.Context, subscriptionID string) error {
	return RenewStripeSubscription(ctx, subscriptionID)
}

// ExpireSubscription expires via Stripe expire action
func (m *StripeMock) ExpireSubscription(ctx context.Context, subscriptionID string) error {
	return ExpireStripeSubscription(ctx, subscriptionID)
}

// AtlosMock implements GatewayMock for Atlos via atlos-sdk mock server
type AtlosMock struct{}

// NewAtlosMock creates a new Atlos mock instance
func NewAtlosMock() *AtlosMock {
	return &AtlosMock{}
}

// Reset resets atlos-mock state via API
func (m *AtlosMock) Reset(ctx context.Context) error {
	return ResetAtlosMock(ctx)
}

// CompleteCheckout simulates the Atlos widget checkout flow
// For Atlos, the sessionID is the order ID (format "{userID}-period{periodID}" or opaque).
// It extracts order/amount from the checkout UI fragments (authoritative source),
// falling back to resolving the amount from the order ID via the admin API.
func (m *AtlosMock) CompleteCheckout(ctx context.Context, sessionID string) (string, error) {
	var orderID string
	var amount float64
	var currency string

	// Parse checkout data from the stored GetCheckoutUI response fragments
	if checkoutUI, ok := GetGatewayCheckoutUI(ctx); ok {
		if data, err := ParseAtlosCheckoutFromFragments(checkoutUI.Fragments); err == nil && data.OrderID != "" {
			orderID = data.OrderID
			amount = data.Amount
			currency = data.Currency
		}
	}

	// Fallback: parse session ID and resolve amount from admin API
	if orderID == "" {
		parsed, _ := ParseAtlosOrderID(sessionID)
		orderID = parsed.Raw
		amount = ResolveAtlosCheckoutAmount(ctx, orderID)
		currency = "USD"
	}

	// Simulate the full checkout flow with a test payment
	// This creates invoice, payment, and completes it (triggering postback)
	_, err := SimulateAtlosCheckout(ctx, orderID, amount, currency)
	if err != nil {
		return "", fmt.Errorf("atlos checkout simulation failed: %w", err)
	}

	// Return the order ID as the subscription identifier
	return orderID, nil
}

// GetCheckoutSession retrieves Atlos checkout session details
// For Atlos, this retrieves the payment status
func (m *AtlosMock) GetCheckoutSession(ctx context.Context, sessionID string) (*CheckoutSession, error) {
	orderID, _ := ParseAtlosOrderID(sessionID)

	// For Atlos, we return the order as the session
	return &CheckoutSession{
		ID:             orderID.Raw,
		SubscriptionID: orderID.Raw,
		Status:         "complete",
		PaymentStatus:  "paid",
	}, nil
}

// SimulatePayment simulates a successful payment via Atlos
// Since Atlos has no native subscription renewal, we trigger the postback
func (m *AtlosMock) SimulatePayment(ctx context.Context, subscriptionID string) error {
	orderID, _ := ParseAtlosOrderID(subscriptionID)

	amount := ResolveAtlosCheckoutAmount(ctx, orderID.Raw)

	_, err := SimulateAtlosCheckout(ctx, orderID.Raw, amount, "USD")
	return err
}

// SetCancelAtPeriodEnd is a no-op for Atlos
// Atlos doesn't support native subscription cancellation at period end
// The portal manages subscription lifecycle independently
func (m *AtlosMock) SetCancelAtPeriodEnd(ctx context.Context, subscriptionID string) error {
	// Atlos doesn't have this concept - subscriptions are managed by the portal
	// Return nil to allow tests to continue
	return nil
}

// PauseSubscription is a no-op for Atlos
// Atlos doesn't support pausing subscriptions natively
func (m *AtlosMock) PauseSubscription(ctx context.Context, subscriptionID string) error {
	// Atlos doesn't support pausing - return error to indicate this
	return fmt.Errorf("atlos does not support subscription pausing")
}

// ResumeSubscription is a no-op for Atlos
// Atlos doesn't support resuming subscriptions natively
func (m *AtlosMock) ResumeSubscription(ctx context.Context, subscriptionID string) error {
	// Atlos doesn't support resuming - return error to indicate this
	return fmt.Errorf("atlos does not support subscription resuming")
}

// RenewSubscription simulates a renewal by creating a new payment
// Atlos doesn't have native subscriptions, so we create a new payment
func (m *AtlosMock) RenewSubscription(ctx context.Context, subscriptionID string) error {
	return m.SimulatePayment(ctx, subscriptionID)
}

// ExpireSubscription is a no-op for Atlos
// Atlos doesn't have native subscription expiration
// The portal handles this via its internal subscription management
func (m *AtlosMock) ExpireSubscription(ctx context.Context, subscriptionID string) error {
	// Atlos doesn't track subscription expiration - return nil
	// The portal's subscription will be marked expired by the test expectations
	return nil
}

// ResetPaymentGateway resets the active gateway mock
// This is the agnostic version that dispatches to the appropriate gateway
func ResetPaymentGateway(ctx context.Context) error {
	mock, err := GetGatewayMock(ctx)
	if err != nil {
		return err
	}
	return mock.Reset(ctx)
}

// CompleteGatewayCheckout completes checkout via the active gateway
func CompleteGatewayCheckout(ctx context.Context, sessionID string) (string, error) {
	mock, err := GetGatewayMock(ctx)
	if err != nil {
		return "", err
	}
	return mock.CompleteCheckout(ctx, sessionID)
}

// SimulateGatewayPayment simulates payment via the active gateway
func SimulateGatewayPayment(ctx context.Context, subscriptionID string) error {
	mock, err := GetGatewayMock(ctx)
	if err != nil {
		return err
	}
	return mock.SimulatePayment(ctx, subscriptionID)
}

// SetGatewayCancelAtPeriodEnd schedules cancellation via active gateway
func SetGatewayCancelAtPeriodEnd(ctx context.Context, subscriptionID string) error {
	mock, err := GetGatewayMock(ctx)
	if err != nil {
		return err
	}
	return mock.SetCancelAtPeriodEnd(ctx, subscriptionID)
}

// PauseGatewaySubscription pauses via active gateway mock
func PauseGatewaySubscription(ctx context.Context, subscriptionID string) error {
	mock, err := GetGatewayMock(ctx)
	if err != nil {
		return err
	}
	return mock.PauseSubscription(ctx, subscriptionID)
}

// ResumeGatewaySubscription resumes via active gateway mock
func ResumeGatewaySubscription(ctx context.Context, subscriptionID string) error {
	mock, err := GetGatewayMock(ctx)
	if err != nil {
		return err
	}
	return mock.ResumeSubscription(ctx, subscriptionID)
}

// RenewGatewaySubscription renews via active gateway mock
func RenewGatewaySubscription(ctx context.Context, subscriptionID string) error {
	mock, err := GetGatewayMock(ctx)
	if err != nil {
		return err
	}
	return mock.RenewSubscription(ctx, subscriptionID)
}

// ExpireGatewaySubscription expires/cancels via active gateway mock
func ExpireGatewaySubscription(ctx context.Context, subscriptionID string) error {
	mock, err := GetGatewayMock(ctx)
	if err != nil {
		return err
	}
	return mock.ExpireSubscription(ctx, subscriptionID)
}
