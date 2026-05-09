package helpers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"go.lumeweb.com/atlos-sdk"
	account "go.lumeweb.com/portal-sdk"
)

var (
	atlosClientOnce sync.Once
	atlosMockClient *atlos.Client
	atlosMockURL    string
	atlosAPISecret  string
)

// initAtlosClient configures the Atlos client
func initAtlosClient() {
	atlosClientOnce.Do(func() {
		atlosMockURL = os.Getenv("ATLOS_MOCK_URL")
		if atlosMockURL == "" {
			port := os.Getenv("ATLOS_MOCK_PORT")
			if port == "" {
				port = "8085"
			}
			atlosMockURL = "http://localhost:" + port
		}

		atlosAPISecret = os.Getenv("ATLOS_API_SECRET")
		if atlosAPISecret == "" {
			atlosAPISecret = "test-secret"
		}

		var err error
		atlosMockClient, err = atlos.NewClient(atlosAPISecret, atlos.WithEndpoint(atlosMockURL))
		if err != nil {
			// The client is initialized lazily, so we can't fail here.
			// The error will surface when the client is used.
		}
	})
}

// getAtlosMockClient returns the configured Atlos SDK client
func getAtlosMockClient() *atlos.Client {
	initAtlosClient()
	return atlosMockClient
}

// AtlosOrderID represents a parsed Atlos order ID.
// Supports three formats produced by the billing plugin:
//   - Legacy:    {userID}-period{periodID}
//   - Regular:   sub-{userID}-{newPeriodID}-{timestamp}-{hmac}
//   - Prorated:  sub-{userID}-{oldPeriodID}-{newPeriodID}-prorated-{timestamp}-{hmac}
type AtlosOrderID struct {
	UserID      string
	OldPeriodID string // Only set for prorated plan changes
	PeriodID    string // NewPeriodID for prorated; the billing period for regular/legacy
	IsProrated  bool
	Raw         string
}

// ParseAtlosOrderID parses an order ID into its structured fields.
// If the format doesn't match any known pattern, returns the raw value with empty fields.
func ParseAtlosOrderID(orderID string) (*AtlosOrderID, error) {
	if orderID == "" {
		return &AtlosOrderID{}, nil
	}

	if strings.HasPrefix(orderID, "sub-") {
		return parseNewFormatOrderID(orderID)
	}

	sep := "-period"
	idx := strings.Index(orderID, sep)
	if idx != -1 {
		return &AtlosOrderID{
			UserID:   orderID[:idx],
			PeriodID: orderID[idx+len(sep):],
			Raw:      orderID,
		}, nil
	}

	return &AtlosOrderID{
		Raw: orderID,
	}, nil
}

func parseNewFormatOrderID(orderID string) (*AtlosOrderID, error) {
	parts := strings.Split(orderID, "-")
	if len(parts) < 5 {
		return &AtlosOrderID{Raw: orderID}, nil
	}

	userID := parts[1]

	isProrated := false
	for _, p := range parts {
		if p == "prorated" {
			isProrated = true
			break
		}
	}

	if isProrated {
		if len(parts) < 7 {
			return &AtlosOrderID{Raw: orderID}, nil
		}
		return &AtlosOrderID{
			UserID:      userID,
			OldPeriodID: parts[2],
			PeriodID:    parts[3],
			IsProrated:  true,
			Raw:         orderID,
		}, nil
	}

	// Regular: sub-{userID}-{newPeriodID}-{timestamp}-{hmac}
	return &AtlosOrderID{
		UserID:   userID,
		PeriodID: parts[2],
		Raw:      orderID,
	}, nil
}

// Context key for Atlos-specific IDs
type atlosCtxKey int

const (
	atlosKeyOrderID atlosCtxKey = iota
	atlosKeyPaymentID
	atlosKeyInvoiceID
	atlosKeyTransactionHash
)

// Context accessors for Atlos-specific IDs

func SetAtlosOrderID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, atlosKeyOrderID, id)
}

func GetAtlosOrderID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(atlosKeyOrderID).(string)
	return v, ok
}

func SetAtlosPaymentID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, atlosKeyPaymentID, id)
}

func GetAtlosPaymentID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(atlosKeyPaymentID).(string)
	return v, ok
}

func SetAtlosInvoiceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, atlosKeyInvoiceID, id)
}

func GetAtlosInvoiceID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(atlosKeyInvoiceID).(string)
	return v, ok
}

func SetAtlosTransactionHash(ctx context.Context, hash string) context.Context {
	return context.WithValue(ctx, atlosKeyTransactionHash, hash)
}

func GetAtlosTransactionHash(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(atlosKeyTransactionHash).(string)
	return v, ok
}

// atlosHTTPClient is reused for raw requests (e.g., Reset endpoint).
// This is a fallback for custom endpoints not in the SDK.
var atlosHTTPClient = &http.Client{}

// ResetAtlosMock resets the atlos-mock-server state via the custom Reset endpoint.
// This endpoint is not a real Atlos API—it's a mock server-specific utility.
func ResetAtlosMock(ctx context.Context) error {
	initAtlosClient()

	req, err := http.NewRequestWithContext(ctx, "POST", atlosMockURL+"/Reset", nil)
	if err != nil {
		return fmt.Errorf("failed to create reset request: %w", err)
	}

	req.Header.Set("ApiSecret", atlosAPISecret)

	resp, err := atlosHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("atlos-mock reset failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("atlos-mock reset returned status %d", resp.StatusCode)
	}

	return nil
}

// SimulateAtlosCheckout simulates the full Atlos widget payment flow via the SDK:
// 1. Create invoice (SDK Client.InvoiceCreate)
// 2. List assets and select one (SDK Client.AssetList)
// 3. Create payment (SDK Client.CreatePayment)
// 4. Complete payment via mock server endpoint (triggers postback in immediate mode)
// Returns the order ID.
func SimulateAtlosCheckout(ctx context.Context, orderID string, amount float64, currency string) (string, error) {
	client := getAtlosMockClient()
	if client == nil {
		return "", fmt.Errorf("failed to initialize Atlos client")
	}

	// Step 1: Create invoice — use SDK types
	merchantID := "test-merchant"
	orderCurrency := currency

	invoiceReq := atlos.InvoiceCreatePostRequest{
		MerchantId:    merchantID,
		OrderAmount:   float32(amount),
		OrderCurrency: &orderCurrency,
		OrderId:       &orderID,
	}

	invoice, err := client.InvoiceCreate(ctx, invoiceReq)
	if err != nil {
		return "", fmt.Errorf("failed to create invoice: %w", err)
	}

	if invoice.Id == nil || *invoice.Id == "" {
		return "", fmt.Errorf("invoice ID is nil or empty")
	}

	// Step 2: List assets and select USDC on Ethereum
	assetListReq := atlos.AssetListPostRequest{
		MerchantId:  merchantID,
		OrderAmount: float32(amount),
		OrderCurrency: &orderCurrency,
	}

	assets, err := client.AssetList(ctx, assetListReq)
	if err != nil {
		return "", fmt.Errorf("failed to list assets: %w", err)
	}

	if len(assets) == 0 {
		return "", fmt.Errorf("no assets available")
	}

	// Find USDC on Ethereum — asset codes are lowercase in the mock data
	var selectedAssetCode string
	var selectedChainId float32
	for _, asset := range assets {
		if asset.Code == nil || !strings.EqualFold(*asset.Code, "usdc") {
			continue
		}
		if asset.Blockchains == nil {
			continue
		}
		for _, chain := range *asset.Blockchains {
			if chain.Code == nil || !strings.EqualFold(*chain.Code, "eth") {
				continue
			}
			if chain.ChainId != nil {
				selectedAssetCode = "USDC"
				selectedChainId = *chain.ChainId
				break
			}
		}
		if selectedAssetCode != "" {
			break
		}
	}

	if selectedAssetCode == "" {
		// Fallback to first available asset/blockchain
		firstAsset := assets[0]
		if firstAsset.Code != nil {
			selectedAssetCode = *firstAsset.Code
		}
		if firstAsset.Blockchains != nil && len(*firstAsset.Blockchains) > 0 {
			firstChain := (*firstAsset.Blockchains)[0]
			if firstChain.ChainId != nil {
				selectedChainId = *firstChain.ChainId
			}
		}
	}

	// Step 3: Create payment — use SDK types
	paymentReq := atlos.CreatePaymentPostRequest{
		InvoiceId:      *invoice.Id,
		AssetCode:      selectedAssetCode,
		BlockchainCode: selectedChainId,
		IsEvm:          "true",
	}

	payment, err := client.CreatePayment(ctx, paymentReq)
	if err != nil {
		return "", fmt.Errorf("failed to create payment: %w", err)
	}

	if payment.Id == nil || *payment.Id == "" {
		return "", fmt.Errorf("payment ID is nil or empty")
	}

	// Step 4: Complete payment via mock server endpoint
	// This uses the raw HTTP client because /Payment/Complete is a mock-only endpoint
	if err := completePaymentRaw(*payment.Id); err != nil {
		return "", fmt.Errorf("failed to complete payment: %w", err)
	}

	return orderID, nil
}

// completePaymentRaw calls the mock server's /Payment/Complete endpoint via raw HTTP.
// This endpoint is not a real Atlos API—it's a mock-only test endpoint.
func completePaymentRaw(paymentID string) error {
	req, err := http.NewRequest("POST", atlosMockURL+"/Payment/Complete", strings.NewReader(
		`{"PaymentId":"`+paymentID+`"}`,
	))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("ApiSecret", atlosAPISecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := atlosHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}

// GetAtlosPayment retrieves payment details by ID using the SDK.
func GetAtlosPayment(ctx context.Context, paymentID string) (*atlos.Payment, error) {
	client := getAtlosMockClient()
	if client == nil {
		return nil, fmt.Errorf("failed to initialize Atlos client")
	}

	req := atlos.PaymentGetPostRequest{
		PaymentId: paymentID,
	}

	return client.PaymentGet(ctx, req)
}

// AtlosCheckoutData holds parsed checkout parameters extracted from
// the payment button fragment generated by the portal.
type AtlosCheckoutData struct {
	OrderID        string
	Amount         float64
	Currency       string
	RecurringUnit  string
	RecurringInterval int
}

// atlosCheckoutDataCtxKey is the context key for parsed Atlos checkout data.
type atlosCheckoutDataCtxKey int

const atlosKeyCheckoutData atlosCheckoutDataCtxKey = iota

// SetAtlosCheckoutData stores parsed Atlos checkout data in context.
func SetAtlosCheckoutData(ctx context.Context, data *AtlosCheckoutData) context.Context {
	return context.WithValue(ctx, atlosKeyCheckoutData, data)
}

// GetAtlosCheckoutData retrieves parsed Atlos checkout data from context.
func GetAtlosCheckoutData(ctx context.Context) (*AtlosCheckoutData, bool) {
	v, ok := ctx.Value(atlosKeyCheckoutData).(*AtlosCheckoutData)
	return v, ok
}

var (
	atlosReOrderAmount      = regexp.MustCompile(`orderAmount:\s*([0-9]+\.?[0-9]*)`)
	atlosReOrderId          = regexp.MustCompile(`orderId:\s*"([^"]+)"`)
	atlosReOrderCurrency    = regexp.MustCompile(`orderCurrency:\s*"([^"]+)"`)
	atlosReRecurringAmount  = regexp.MustCompile(`recurringAmount:\s*([0-9]+\.?[0-9]*)`)
	atlosReRecurringUnit    = regexp.MustCompile(`recurringUnit:\s*"([^"]+)"`)
	atlosReRecurringInterval = regexp.MustCompile(`recurringInterval:\s*([0-9]+)`)
)

// StoreAtlosCheckoutFromUI parses Atlos checkout data from the GetCheckoutUI
// response fragments and stores it in context. This should be called after every
// GetCheckoutUI call when the Atlos gateway is active, so that CompleteCheckout
// uses the fragment data as the single source of truth instead of re-querying
// the admin API.
func StoreAtlosCheckoutFromUI(ctx context.Context, checkoutUI *account.CheckoutUI) (context.Context, error) {
	if checkoutUI == nil || !GatewayIsAtlos(ctx) {
		return ctx, nil
	}

	data, err := ParseAtlosCheckoutFromFragments(checkoutUI.Fragments)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse Atlos checkout from fragments: %w", err)
	}

	ctx = SetAtlosCheckoutData(ctx, data)

	// Also store amount/currency in gateway-agnostic context keys
	ctx = SetGatewayCheckoutAmount(ctx, data.Amount)
	ctx = SetGatewayCheckoutCurrency(ctx, data.Currency)

	return ctx, nil
}

// ParseAtlosCheckoutFromFragments extracts checkout data from the
// payment button fragment script in a GetCheckoutUI response.
// The button script contains a paymentConfig JS object with orderId,
// orderAmount, orderCurrency, recurringAmount, recurringUnit, recurringInterval.
func ParseAtlosCheckoutFromFragments(fragments []account.CheckoutUIFragment) (*AtlosCheckoutData, error) {
	for _, frag := range fragments {
		if frag.Script == nil || *frag.Script == "" {
			continue
		}

		script := *frag.Script

		amountMatch := atlosReOrderAmount.FindStringSubmatch(script)
		idMatch := atlosReOrderId.FindStringSubmatch(script)
		currencyMatch := atlosReOrderCurrency.FindStringSubmatch(script)

		if idMatch == nil || amountMatch == nil {
			continue
		}

		amount, err := strconv.ParseFloat(amountMatch[1], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse orderAmount %q: %w", amountMatch[1], err)
		}

		data := &AtlosCheckoutData{
			OrderID:  idMatch[1],
			Amount:   amount,
			Currency: "USD",
		}

		if currencyMatch != nil {
			data.Currency = currencyMatch[1]
		}

		if m := atlosReRecurringUnit.FindStringSubmatch(script); m != nil {
			data.RecurringUnit = m[1]
		}
		if m := atlosReRecurringInterval.FindStringSubmatch(script); m != nil {
			if v, err := strconv.Atoi(m[1]); err == nil {
				data.RecurringInterval = v
			}
		}

		return data, nil
	}

	return nil, fmt.Errorf("no payment button fragment found with orderId and orderAmount")
}

// ResolveAtlosCheckoutAmount resolves the payment amount from the order ID.
// For new-format IDs (sub-...), PeriodID is the newPeriodID from the order ID.
// For prorated orders, returns period.PriceUSD of the new period (the fragment
// path should be used for the actual prorated amount).
// Falls back to 10.0 if lookup fails.
func ResolveAtlosCheckoutAmount(ctx context.Context, orderID string) float64 {
	parsed, _ := ParseAtlosOrderID(orderID)
	if parsed.PeriodID == "" {
		return 10.0
	}

	periodID, err := strconv.Atoi(parsed.PeriodID)
	if err != nil {
		return 10.0
	}

	adminClient, err := RequireAdminClient(ctx)
	if err != nil {
		return 10.0
	}

	billingService := adminClient.Billing()
	period, err := billingService.GetPricingPlanPeriod(ctx, fmt.Sprint(periodID))
	if err != nil || period == nil {
		return 10.0
	}

	return float64(period.PriceUsd)
}

// generateTestTxHash generates a fake transaction hash for testing (64-char hex).
func generateTestTxHash() string {
	return strings.Repeat("01234567abcdef00", 4)
}
