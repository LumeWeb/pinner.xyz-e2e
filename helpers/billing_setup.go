package helpers

import (
	"context"
	"fmt"
	"time"

	account "go.lumeweb.com/portal-sdk"
	"go.lumeweb.com/portal-sdk/admin"
)

// PricingPlanInfo holds info about a single plan with its periods
type PricingPlanInfo struct {
	PlanID          uint
	PeriodIDs       []uint // Monthly and yearly period IDs
	StripeProductID string
	StripePriceIDs  []string
}

// BillingInfrastructure holds created billing resources
type BillingInfrastructure struct {
	PriceLineID uint
	Plans       []PricingPlanInfo // Multiple plans with their periods
}

// PlanLifecycle contains plan creation and deletion helpers
type PlanLifecycle struct {
	infra     *BillingInfrastructure
	planIndex int
}

// NewPlanLifecycle creates a new plan lifecycle for creating a plan with periods
func NewPlanLifecycle(infra *BillingInfrastructure, planIndex int) *PlanLifecycle {
	return &PlanLifecycle{infra: infra, planIndex: planIndex}
}

// CreatePricingPlan creates a plan with monthly and yearly periods
func (pl *PlanLifecycle) CreatePricingPlan(ctx context.Context, billingService *admin.BillingService, planDef PlanDefinition) (*admin.PricingPlan, error) {
	createdPlan, err := billingService.CreatePricingPlan(ctx, &admin.PricingPlanCreateRequest{
		Name:           planDef.Name,
		Description:    planDef.Description,
		Currency:       "USD",
		IsActive:       true,
		IsPublic:       true, // Plans must be public for users to check out
		PricingPeriods: []admin.PricingPlanPeriod{},
		PricelineId:    new(int(pl.infra.PriceLineID)),
		Position:       &pl.planIndex,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pricing plan %s: %w", planDef.Name, err)
	}

	return createdPlan, nil
}

// CreatePeriod creates a pricing plan period with specified cadence
func (pl *PlanLifecycle) CreatePeriod(ctx context.Context, billingService *admin.BillingService, planID int, cadence string, priceUsd float32, quotaPlanID int64) (uint, error) {
	period := &admin.PricingPlanPeriodCreateRequest{
		Cadence:       cadence,
		PriceUsd:      priceUsd,
		PricingPlanId: planID,
		QuotaPlanId:   int(quotaPlanID),
	}

	createdPeriod, err := billingService.CreatePricingPlanPeriod(ctx, period)
	if err != nil {
		return 0, fmt.Errorf("failed to create %s period: %w", cadence, err)
	}

	return uint(createdPeriod.Id), nil
}

// AppendPlanInfo adds plan info with periods to infrastructure
func (pl *PlanLifecycle) AppendPlanInfo(planID uint, periodIDs []uint) {
	pl.infra.Plans = append(pl.infra.Plans, PricingPlanInfo{
		PlanID:    planID,
		PeriodIDs: periodIDs,
	})
}

// FirstPlan returns the first plan (convenience method for tests that need a single plan)
func (b *BillingInfrastructure) FirstPlan() *PricingPlanInfo {
	if len(b.Plans) == 0 {
		return nil
	}
	return &b.Plans[0]
}

// FirstPlanID returns the first plan's ID
func (b *BillingInfrastructure) FirstPlanID() uint {
	if p := b.FirstPlan(); p != nil {
		return p.PlanID
	}
	return 0
}

// FirstPeriodID returns the first plan's first period ID (monthly)
func (b *BillingInfrastructure) FirstPeriodID() uint {
	if p := b.FirstPlan(); p != nil && len(p.PeriodIDs) > 0 {
		return p.PeriodIDs[0]
	}
	return 0
}

// MonthlyPeriodID returns the monthly period ID for the first plan
func (b *BillingInfrastructure) MonthlyPeriodID() uint {
	return b.FirstPeriodID()
}

// YearlyPeriodID returns the yearly period ID for the first plan
func (b *BillingInfrastructure) YearlyPeriodID() uint {
	if p := b.FirstPlan(); p != nil && len(p.PeriodIDs) > 1 {
		return p.PeriodIDs[1]
	}
	return 0
}

// PlanByName finds a pricing plan by its name
func (b *BillingInfrastructure) PlanByName(name string) *PricingPlanInfo {
	index := PlanIndexByName(name)
	if index < 0 || index >= len(b.Plans) {
		return nil
	}
	return &b.Plans[index]
}

// PlanIndexByName returns the index of a plan by its name
func PlanIndexByName(name string) int {
	for i, planDef := range DefaultPlans {
		if planDef.Name == name {
			return i
		}
	}
	return -1
}

// PlanByIndex returns a plan by its index
func (b *BillingInfrastructure) PlanByIndex(index int) *PricingPlanInfo {
	if index < 0 || index >= len(b.Plans) {
		return nil
	}
	return &b.Plans[index]
}

// PlanIDByIndex returns the plan ID for a given index
func (b *BillingInfrastructure) PlanIDByIndex(index int) uint {
	if p := b.PlanByIndex(index); p != nil {
		return p.PlanID
	}
	return 0
}

// PlanIDByName returns the plan ID for a given plan name
func (b *BillingInfrastructure) PlanIDByName(name string) uint {
	index := PlanIndexByName(name)
	if index < 0 {
		return 0
	}
	return b.PlanIDByIndex(index)
}

// PeriodIDByPlanAndCadence returns a period ID for a given plan name and cadence
func (b *BillingInfrastructure) PeriodIDByPlanAndCadence(planName, cadence string) uint {
	index := PlanIndexByName(planName)
	if index < 0 {
		return 0
	}
	p := b.PlanByIndex(index)
	if p == nil {
		return 0
	}
	// Periods are stored as [monthly, yearly]
	if cadence == "monthly" && len(p.PeriodIDs) > 0 {
		return p.PeriodIDs[0]
	}
	if cadence == "yearly" && len(p.PeriodIDs) > 1 {
		return p.PeriodIDs[1]
	}
	return 0
}

// PlanDefinition defines a plan to create
type PlanDefinition struct {
	Name         string
	Description  string
	MonthlyPrice float32
	YearlyPrice  float32
}

// DefaultPlans defines the standard plans to create
var DefaultPlans = []PlanDefinition{
	{Name: "Basic", Description: "Basic storage plan", MonthlyPrice: 10.00, YearlyPrice: 100.00},
	{Name: "Pro", Description: "Professional storage plan", MonthlyPrice: 25.00, YearlyPrice: 250.00},
	{Name: "Enterprise", Description: "Enterprise storage plan", MonthlyPrice: 50.00, YearlyPrice: 500.00},
}

// SetupBillingInfrastructure creates billing infrastructure via portal admin API
// Creates: one price line with multiple plans, each with monthly and yearly periods
// Portal auto-triggers sync to Stripe
func SetupBillingInfrastructure(ctx context.Context) (context.Context, error) {
	return SetupBillingInfrastructureWithPlans(ctx, DefaultPlans)
}

// SetupBillingInfrastructureWithPlans creates billing infrastructure with custom plans
func SetupBillingInfrastructureWithPlans(ctx context.Context, plans []PlanDefinition) (context.Context, error) {
	// Get admin client
	adminClient, err := RequireAdminClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get admin client: %w", err)
	}

	// Ensure a default quota plan exists before creating billing periods
	// Billing infrastructure needs quota plans to reference when creating pricing plan periods
	quotaPlanID, err := EnsureDefaultQuotaPlan(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to ensure default quota plan: %w", err)
	}

	// Check if a default price line already exists
	priceLines, _, err := adminClient.Billing().ListPriceLines(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list price lines: %w", err)
	}

	var defaultPriceLineExists bool
	var existingDefaultPriceLineID int
	for _, pl := range priceLines {
		if pl.IsDefault {
			defaultPriceLineExists = true
			existingDefaultPriceLineID = int(pl.Id)
			break
		}
	}

	// Create price line only if no default exists
	var priceLineID uint
	if !defaultPriceLineExists {
		priceLine := &admin.PriceLineCreateRequest{
			Name:        "Test Price Line",
			Description: "Test price line for subscription testing",
			IsActive:    true,
			IsDefault:   true,
		}

		createdPriceLine, err := adminClient.Billing().CreatePriceLine(ctx, priceLine)
		if err != nil {
			return ctx, fmt.Errorf("failed to create price line: %w", err)
		}
		priceLineID = uint(createdPriceLine.Id)
	} else {
		// Use existing default price line ID for sync verification
		priceLineID = uint(existingDefaultPriceLineID)
	}

	infra := &BillingInfrastructure{
		PriceLineID: priceLineID,
		Plans:       make([]PricingPlanInfo, 0, len(plans)),
	}

	// Get billing service
	billingService := adminClient.Billing()
	if billingService == nil {
		return ctx, fmt.Errorf("admin client's Billing() returned nil")
	}

	// Create each plan with monthly and yearly periods
	for i, planDef := range plans {
		lifecycle := NewPlanLifecycle(infra, i)

		// Create plan
		createdPlan, err := lifecycle.CreatePricingPlan(ctx, billingService, planDef)
		if err != nil {
			return ctx, err
		}

		// Create monthly period with quota plan reference
		monthlyPeriodID, err := lifecycle.CreatePeriod(ctx, billingService, int(createdPlan.Id), "monthly", planDef.MonthlyPrice, quotaPlanID)
		if err != nil {
			return ctx, err
		}

		// Create yearly period with quota plan reference
		yearlyPeriodID, err := lifecycle.CreatePeriod(ctx, billingService, int(createdPlan.Id), "yearly", planDef.YearlyPrice, quotaPlanID)
		if err != nil {
			return ctx, err
		}

		// Append plan info with periods
		lifecycle.AppendPlanInfo(uint(createdPlan.Id), []uint{monthlyPeriodID, yearlyPeriodID})
	}

	// Store in context for cleanup
	infraID := fmt.Sprintf("billing-infra-%d", time.Now().UnixNano())
	ctx = SetBillingInfrastructure(ctx, infraID, infra)

	// Wait for all plans to sync to Stripe
	err = PollAllPlansStripeSync(ctx, infra, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("failed to verify Stripe sync: %w", err)
	}

	return ctx, nil
}

const StripeMetadataPlanIDKey = "plan_id"

// PollAllPlansStripeSync waits for all plans to sync to Stripe
func PollAllPlansStripeSync(ctx context.Context, infra *BillingInfrastructure, timeout time.Duration) error {
	// Poll each plan individually
	for i := range infra.Plans {
		err := PollStripeProductSync(ctx, &infra.Plans[i], timeout)
		if err != nil {
			return fmt.Errorf("plan %d: %w", infra.Plans[i].PlanID, err)
		}
	}
	return nil
}

// PollStripeProductSync waits for portal → Stripe sync for a single plan
// Polls stripe-mock GET /v1/products
// Matches product by metadata.plan_id = planID
func PollStripeProductSync(ctx context.Context, planInfo *PricingPlanInfo, timeout time.Duration) error {
	planID := planInfo.PlanID
	// Give portal some time to trigger the sync
	time.Sleep(5 * time.Second)

	// Poll stripe-mock for products matching the plan
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			// List all products from stripe-mock
			products, err := ListStripeProducts(ctx)
			if err != nil {
				continue
			}

			// Find product that matches our plan
			for _, product := range products {
				// Check metadata for portal_plan_id
				if product.Metadata != nil {
					if planIDStr, ok := product.Metadata[StripeMetadataPlanIDKey]; ok {
						if fmt.Sprint(planID) == planIDStr {
							planInfo.StripeProductID = product.ID

							// Collect all prices for this product
							if product.DefaultPrice != nil {
								planInfo.StripePriceIDs = append(planInfo.StripePriceIDs, product.DefaultPrice.ID)
							}

							// Verify at least one price exists
							if len(planInfo.StripePriceIDs) > 0 {
								return nil
							}
							break
						}
					}
				}
			}

		case <-timeoutCh:
			return fmt.Errorf("timed out waiting for plan %d to sync to Stripe", planID)
		}
	}
}

// checkSubscriptionStatus checks if a subscription status matches the expected value.
// Returns nil on match, or a descriptive error on mismatch.
func checkSubscriptionStatus(status *account.SubscriptionStatus, expected string) error {
	switch expected {
	case "active":
		if !status.IsSubscribed {
			return fmt.Errorf("expected subscription status 'active', got inactive")
		}
		if status.WillCancelAt != nil {
			return fmt.Errorf("expected subscription status 'active', got cancel_at_period_end (will_cancel_at=%s)", status.WillCancelAt.Format(time.RFC3339))
		}
	case "cancel_at_period_end":
		if !status.IsSubscribed {
			return fmt.Errorf("expected subscription status 'cancel_at_period_end', got inactive")
		}
		if status.WillCancelAt == nil {
			return fmt.Errorf("expected subscription status 'cancel_at_period_end', got active (will_cancel_at is nil)")
		}
	case "canceled":
		if status.IsSubscribed {
			return fmt.Errorf("expected subscription status 'canceled', got active")
		}
	case "paused":
		if !status.IsSubscribed {
			return fmt.Errorf("expected subscription status 'paused', got inactive")
		}
	default:
		return fmt.Errorf("unknown subscription status: %s", expected)
	}
	return nil
}

// VerifySubscriptionStatus polls until the user's subscription status matches expected.
// Polling accommodates asynchronous webhook processing (e.g., cancel_at_period_end
// being set by a webhook after the cancel API call returns).
func VerifySubscriptionStatus(ctx context.Context, expectedStatus string) error {
	const timeout = 30 * time.Second

	userClient, err := RequireAuthenticatedClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user client: %w", err)
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	// Check immediately before entering the poll loop
	initialStatus, err := userClient.GetSubscriptionStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get subscription status: %w", err)
	}
	if checkErr := checkSubscriptionStatus(initialStatus, expectedStatus); checkErr == nil {
		return nil
	}

	for {
		select {
		case <-ticker.C:
			latestStatus, err := userClient.GetSubscriptionStatus(ctx)
			if err != nil {
				continue
			}
			if checkErr := checkSubscriptionStatus(latestStatus, expectedStatus); checkErr == nil {
				return nil
			}
		case <-timeoutCh:
			finalStatus, _ := userClient.GetSubscriptionStatus(ctx)
			return fmt.Errorf("timed out waiting for subscription status '%s': %w", expectedStatus, checkSubscriptionStatus(finalStatus, expectedStatus))
		}
	}
}

// CleanupBillingInfrastructureImpl removes created billing resources via portal admin API
func CleanupBillingInfrastructureImpl(ctx context.Context, infra *BillingInfrastructure) error {
	if infra == nil {
		return nil
	}

	// Get admin client and billing service
	adminClient, err := RequireAdminClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get admin client: %w", err)
	}

	billingService := adminClient.Billing()
	if billingService == nil {
		return fmt.Errorf("admin client's Billing() returned nil")
	}

	// Delete plans from price line first (in reverse order)
	for i := len(infra.Plans) - 1; i >= 0; i-- {
		planInfo := infra.Plans[i]

		// Remove plan from price line before deleting plan
		if planInfo.PlanID > 0 {
			err := billingService.DeletePlanFromPriceLine(ctx, fmt.Sprint(infra.PriceLineID), fmt.Sprint(planInfo.PlanID))
			if err != nil {
				// Plan already removed, not an error
			}
		}
	}

	// Delete all periods and plans in reverse order
	for i := len(infra.Plans) - 1; i >= 0; i-- {
		planInfo := infra.Plans[i]

		// Delete periods first (in reverse order)
		for j := len(planInfo.PeriodIDs) - 1; j >= 0; j-- {
			err := billingService.DeletePricingPlanPeriod(ctx, fmt.Sprint(planInfo.PeriodIDs[j]))
			if err != nil {
				return fmt.Errorf("failed to delete pricing plan period %d: %w", planInfo.PeriodIDs[j], err)
			}
		}

		// Delete the plan
		if planInfo.PlanID > 0 {
			err := billingService.DeletePricingPlan(ctx, fmt.Sprint(planInfo.PlanID))
			if err != nil {
				return fmt.Errorf("failed to delete pricing plan %d: %w", planInfo.PlanID, err)
			}
		}
	}

	// Delete price line
	if infra.PriceLineID > 0 {
		err := billingService.DeletePriceLine(ctx, fmt.Sprint(infra.PriceLineID))
		if err != nil {
			return fmt.Errorf("failed to delete price line %d: %w", infra.PriceLineID, err)
		}
	}

	return nil
}

// PollSubscriptionStatus polls portal API for user's subscription status
// active should be true for subscribed, false for unsubscribed
func PollSubscriptionStatus(ctx context.Context, userID int, active bool, timeout time.Duration) error {
	adminClient, err := RequireAdminClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get admin client: %w", err)
	}

	// Get user's subscribers using GetUserSubscribers (returns 3 values)
	subscribers, _, err := adminClient.Billing().GetUserSubscribers(ctx, fmt.Sprint(userID))
	if err != nil {
		return fmt.Errorf("failed to get user subscribers: %w", err)
	}

	if len(subscribers) == 0 {
		return fmt.Errorf("no subscriber found for user %d", userID)
	}

	// Poll for expected status
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			// Fetch latest subscription status
			latestSubscribers, _, err := adminClient.Billing().GetUserSubscribers(ctx, fmt.Sprint(userID))
			if err != nil {
				continue
			}

			if len(latestSubscribers) > 0 {
				// Check if active state matches
				if latestSubscribers[0].IsActive == active {
					return nil
				}
			}

		case <-timeoutCh:
			return fmt.Errorf("timed out waiting for subscription active=%v, got: %v", active, subscribers[0].IsActive)
		}
	}
}

// PollUserSubscriptionStatus polls user's own subscription status.
// Delegates to VerifySubscriptionStatus for DRY polling logic.
func PollUserSubscriptionStatus(ctx context.Context, subscribed bool, timeout time.Duration) error {
	expected := "canceled"
	if subscribed {
		expected = "active"
	}
	return VerifySubscriptionStatus(ctx, expected)
}
