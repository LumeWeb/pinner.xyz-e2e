package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/checkout/session"
	"github.com/stripe/stripe-go/v85/customer"
	"github.com/stripe/stripe-go/v85/price"
	"github.com/stripe/stripe-go/v85/product"
	"github.com/stripe/stripe-go/v85/subscription"
	"github.com/stripe/stripe-go/v85/webhook"
)

var (
	stripeBackendOnce sync.Once
	stripeMockURL     string
	stripeAPIKey      string
	stripeHTTPClient  = &http.Client{}
)

// initStripeBackend configures the Stripe SDK to use the stripe-mock server
func initStripeBackend() {
	stripeBackendOnce.Do(func() {
		stripeMockURL = os.Getenv("STRIPE_MOCK_URL")
		if stripeMockURL == "" {
			port := os.Getenv("STRIPE_MOCK_PORT")
			if port == "" {
				port = "80"
			}
			stripeMockURL = "http://localhost:" + port
		}

		stripeAPIKey = os.Getenv("PORTAL__BILLING__STRIPE__API_KEY")
		if stripeAPIKey == "" {
			stripeAPIKey = "sk_test_mock"
		}

		stripe.Key = stripeAPIKey

		existingBackend := stripe.GetBackend(stripe.APIBackend)
		if existingImpl, ok := existingBackend.(*stripe.BackendImplementation); ok {
			stripe.SetBackend(stripe.APIBackend, &stripe.BackendImplementation{
				Type:              stripe.APIBackend,
				URL:               stripeMockURL,
				HTTPClient:        existingImpl.HTTPClient,
				LeveledLogger:     existingImpl.LeveledLogger,
				MaxNetworkRetries: existingImpl.MaxNetworkRetries,
			})
		}
	})
}

// Context key and value helpers
type stripeCtxKey int

const (
	stripeKeyCustomerID stripeCtxKey = iota
	stripeKeySubscriptionID
	stripeKeyCheckoutSessionID
	stripeKeyProductID
	stripeKeyPriceID
	stripeKeyWebhookSecret
)

func stripeCtxGet(ctx context.Context, key stripeCtxKey) (string, bool) {
	v, ok := ctx.Value(key).(string)
	return v, ok
}

func stripeCtxSet(ctx context.Context, key stripeCtxKey, val string) context.Context {
	return context.WithValue(ctx, key, val)
}

// Context accessors for Stripe-specific IDs
func SetStripeCustomerID(ctx context.Context, id string) context.Context    { return stripeCtxSet(ctx, stripeKeyCustomerID, id) }
func GetStripeCustomerID(ctx context.Context) (string, bool)                { return stripeCtxGet(ctx, stripeKeyCustomerID) }
func SetStripeSubscriptionID(ctx context.Context, id string) context.Context { return stripeCtxSet(ctx, stripeKeySubscriptionID, id) }
func GetStripeSubscriptionID(ctx context.Context) (string, bool)             { return stripeCtxGet(ctx, stripeKeySubscriptionID) }
func SetStripeCheckoutSessionID(ctx context.Context, id string) context.Context {
	return stripeCtxSet(ctx, stripeKeyCheckoutSessionID, id)
}
func GetStripeCheckoutSessionID(ctx context.Context) (string, bool) { return stripeCtxGet(ctx, stripeKeyCheckoutSessionID) }
func SetStripeProductID(ctx context.Context, id string) context.Context {
	return stripeCtxSet(ctx, stripeKeyProductID, id)
}
func GetStripeProductID(ctx context.Context) (string, bool) { return stripeCtxGet(ctx, stripeKeyProductID) }

// stripeMockPost performs a POST to a custom stripe-mock endpoint
func stripeMockPost(path string) error {
	initStripeBackend()
	req, err := http.NewRequest("POST", stripeMockURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+stripeAPIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := stripeHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// stripeMockGet performs a GET to a custom stripe-mock endpoint
func stripeMockGet(path string) ([]byte, error) {
	initStripeBackend()
	req, err := http.NewRequest("GET", stripeMockURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+stripeAPIKey)

	resp, err := stripeHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}

// ResetStripeMock resets the stripe-mock-server state
func ResetStripeMock(ctx context.Context) error {
	if err := stripeMockPost("/v1/reset"); err != nil {
		return fmt.Errorf("stripe-mock reset failed: %w", err)
	}
	return nil
}

// ListStripeProducts lists all products from Stripe
func ListStripeProducts(ctx context.Context) ([]*stripe.Product, error) {
	initStripeBackend()
	iter := product.List(&stripe.ProductListParams{
		Expand: []*string{stripe.String("data.default_price")},
	})
	var products []*stripe.Product
	for iter.Next() {
		products = append(products, iter.Product())
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}
	return products, nil
}

// ListStripePrices lists all prices from Stripe
func ListStripePrices(ctx context.Context) ([]*stripe.Price, error) {
	initStripeBackend()
	iter := price.List(&stripe.PriceListParams{
		Expand: []*string{stripe.String("data.product")},
	})
	var prices []*stripe.Price
	for iter.Next() {
		prices = append(prices, iter.Price())
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prices: %w", err)
	}
	return prices, nil
}

// GetStripeProduct retrieves a Stripe product by ID
func GetStripeProduct(ctx context.Context, id string) (*stripe.Product, error) {
	initStripeBackend()
	prod, err := product.Get(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return prod, nil
}

// GetStripePrice retrieves a Stripe price by ID
func GetStripePrice(ctx context.Context, id string) (*stripe.Price, error) {
	initStripeBackend()
	p, err := price.Get(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}
	return p, nil
}

// GetStripeCheckoutSession retrieves a Stripe checkout session by ID
func GetStripeCheckoutSession(ctx context.Context, id string) (*stripe.CheckoutSession, error) {
	initStripeBackend()
	sess, err := session.Get(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkout session: %w", err)
	}
	return sess, nil
}

// GetStripeSubscription retrieves a Stripe subscription by ID
func GetStripeSubscription(ctx context.Context, id string) (*stripe.Subscription, error) {
	initStripeBackend()
	sub, err := subscription.Get(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return sub, nil
}

// GetStripeCustomer retrieves a Stripe customer by ID
func GetStripeCustomer(ctx context.Context, id string) (*stripe.Customer, error) {
	initStripeBackend()
	cust, err := customer.Get(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	return cust, nil
}

// GetStripeWebhookSecret retrieves the webhook secret from stripe-mock
func GetStripeWebhookSecret(ctx context.Context) (string, error) {
	body, err := stripeMockGet("/v1/webhook_secret")
	if err != nil {
		return "", fmt.Errorf("failed to get webhook secret: %w", err)
	}
	var result struct {
		Secret string `json:"secret"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to decode webhook secret: %w", err)
	}
	return result.Secret, nil
}

// CompleteStripeCheckoutSession completes a checkout session via stripe-mock custom endpoint
func CompleteStripeCheckoutSession(ctx context.Context, id string) error {
	if err := stripeMockPost(fmt.Sprintf("/v1/checkout/sessions/%s/complete", id)); err != nil {
		return fmt.Errorf("checkout complete failed: %w", err)
	}
	return nil
}

// stripeMockPostWithResponse performs POST and returns response body
func stripeMockPostWithResponse(path string) ([]byte, error) {
	initStripeBackend()
	req, err := http.NewRequest("POST", stripeMockURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+stripeAPIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := stripeHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// stripeSubscriptionAction performs a subscription action (renew/expire/pause/resume) via stripe-mock
func stripeSubscriptionAction(action, id string) (*stripe.Subscription, error) {
	path := fmt.Sprintf("/v1/subscriptions/%s/%s", id, action)
	body, err := stripeMockPostWithResponse(path)
	if err != nil {
		return nil, fmt.Errorf("subscription %s failed: %w", action, err)
	}
	var sub stripe.Subscription
	if err := json.Unmarshal(body, &sub); err != nil {
		return nil, fmt.Errorf("failed to decode subscription: %w", err)
	}
	return &sub, nil
}

// RenewStripeSubscriptionAction renews a subscription via stripe-mock custom endpoint
func RenewStripeSubscriptionAction(ctx context.Context, id string) (*stripe.Subscription, error) {
	return stripeSubscriptionAction("renew", id)
}

// ExpireStripeSubscriptionAction expires a subscription via stripe-mock custom endpoint
func ExpireStripeSubscriptionAction(ctx context.Context, id string) (*stripe.Subscription, error) {
	return stripeSubscriptionAction("expire", id)
}

// PauseStripeSubscriptionAction pauses a subscription via stripe-mock custom endpoint
func PauseStripeSubscriptionAction(ctx context.Context, id string) (*stripe.Subscription, error) {
	return stripeSubscriptionAction("pause", id)
}

// ResumeStripeSubscriptionAction resumes a subscription via stripe-mock custom endpoint
func ResumeStripeSubscriptionAction(ctx context.Context, id string) (*stripe.Subscription, error) {
	return stripeSubscriptionAction("resume", id)
}

// VerifyStripeWebhookSignature verifies webhook signature using Stripe library
func VerifyStripeWebhookSignature(payload []byte, signature, secret string) (stripe.Event, error) {
	initStripeBackend()
	return webhook.ConstructEvent(payload, signature, secret)
}

// RenewStripeSubscription is a convenience wrapper that discards the return value
func RenewStripeSubscription(ctx context.Context, id string) error {
	_, err := RenewStripeSubscriptionAction(ctx, id)
	return err
}

// ExpireStripeSubscription is a convenience wrapper that discards the return value
func ExpireStripeSubscription(ctx context.Context, id string) error {
	_, err := ExpireStripeSubscriptionAction(ctx, id)
	return err
}

// PauseStripeSubscription is a convenience wrapper that discards the return value
func PauseStripeSubscription(ctx context.Context, id string) error {
	_, err := PauseStripeSubscriptionAction(ctx, id)
	return err
}

// ResumeStripeSubscription is a convenience wrapper that discards the return value
func ResumeStripeSubscription(ctx context.Context, id string) error {
	_, err := ResumeStripeSubscriptionAction(ctx, id)
	return err
}

// SimulateStripePaymentSuccess simulates a successful payment for a subscription
// In stripe-mock, this is done by renewing the subscription
func SimulateStripePaymentSuccess(ctx context.Context, subscriptionID string) error {
	return RenewStripeSubscription(ctx, subscriptionID)
}

// SetStripeCancelAtPeriodEnd sets a subscription for cancellation at period end via Stripe API
// This simulates what the Stripe Customer Portal does when a user cancels a subscription
func SetStripeCancelAtPeriodEnd(ctx context.Context, subscriptionID string) error {
	initStripeBackend()

	sub, err := subscription.Update(subscriptionID, &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to set cancel_at_period_end: %w", err)
	}

	if sub == nil || !sub.CancelAtPeriodEnd {
		return fmt.Errorf("subscription was not updated with cancel_at_period_end")
	}

	return nil
}


