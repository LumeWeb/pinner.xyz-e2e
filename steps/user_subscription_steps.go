package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	account "go.lumeweb.com/portal-sdk"
	"pinner.xyz-e2e/helpers"
)

// UserSubscriptionSteps holds step definitions for user subscription flow
// These steps are gateway-agnostic and work with any configured payment gateway.
// Gateway-specific behavior is delegated to GatewayMock implementations.
type UserSubscriptionSteps struct{}

// NewUserSubscriptionSteps creates a new UserSubscriptionSteps instance
func NewUserSubscriptionSteps() *UserSubscriptionSteps {
	return &UserSubscriptionSteps{}
}

// InitializeScenario registers all step definitions with godog
func (s *UserSubscriptionSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Checkout
	ctx.Step(`^the user creates a checkout session for the plan$`, s.theUserCreatesACheckoutSessionForThePlan)
	ctx.Step(`^the checkout session completes$`, s.theCheckoutSessionCompletes)

	// Subscription lifecycle
	ctx.Step(`^the user cancels the subscription$`, s.theUserCancelsTheSubscription)
	ctx.Step(`^the user has an active subscription$`, s.theUserHasAnActiveSubscription)

	// Upgrade/downgrade via portal API
	ctx.Step(`^the user upgrades to the ([^"]+) plan$`, s.theUserUpgradesToThePlan)
	ctx.Step(`^the user downgrades to the ([^"]+) plan$`, s.theUserDowngradesToThePlan)
	ctx.Step(`^the user's current plan is ([^"]+)$`, s.theUsersCurrentPlanIs)
	ctx.Step(`^the user has an active subscription on the ([^"]+) plan$`, s.theUserHasAnActiveSubscriptionOnThePlan)
	ctx.Step(`^the user has an active ([^"]+) subscription on the ([^"]+) plan$`, s.theUserHasAnActiveSubscriptionOnThePlanWithCadence)
	ctx.Step(`^the user selects ([^"]+) billing$`, s.theUserSelectsBillingCadence)

	// Test infrastructure
	ctx.Step(`^a pricing plan period exists$`, s.aPricingPlanPeriodExists)

	// Subscription status verification via portal API
	ctx.Step(`^the subscription status is "([^"]*)"$`, s.theSubscriptionStatusIs)

	// Payment simulation via gateway mock
	ctx.Step(`^the user pays for the next billing period$`, s.theUserPaysForTheNextBillingPeriod)
}

// theUserInitiatesCheckoutForThePlan initiates checkout for a pricing plan
func (s *UserSubscriptionSteps) theUserInitiatesCheckoutForThePlan(ctx context.Context) (context.Context, error) {
	infra, _, ok := helpers.GetBillingInfrastructure(ctx)
	if !ok || infra == nil {
		return ctx, fmt.Errorf("no billing infrastructure available")
	}

	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	checkoutUI, err := userClient.GetCheckoutUI(ctx, fmt.Sprint(infra.FirstPlanID()),
		account.WithPeriodID(fmt.Sprint(infra.FirstPeriodID())),
		account.WithGateway(helpers.GetActiveGatewaySDKName(ctx)),
	)
	if err != nil {
		return ctx, fmt.Errorf("failed to get checkout UI: %w", err)
	}

	ctx = helpers.SetGatewayCheckoutUI(ctx, checkoutUI)

	if checkoutUI.SessionId != nil {
		ctx = helpers.SetGatewayCheckoutSessionID(ctx, *checkoutUI.SessionId)
	}

	return ctx, nil
}

// theCheckoutCompletesSuccessfully completes the checkout via the active gateway mock
func (s *UserSubscriptionSteps) theCheckoutCompletesSuccessfully(ctx context.Context) (context.Context, error) {
	sessionID, ok := helpers.GetGatewayCheckoutSessionID(ctx)
	if !ok || sessionID == "" {
		return ctx, fmt.Errorf("no checkout session ID available")
	}

	subscriptionID, err := helpers.CompleteGatewayCheckout(ctx, sessionID)
	if err != nil {
		return ctx, fmt.Errorf("failed to complete checkout: %w", err)
	}

	if subscriptionID != "" {
		ctx = helpers.SetGatewaySubscriptionID(ctx, subscriptionID)
	}

	err = helpers.PollUserSubscriptionStatus(ctx, true, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("subscription did not become active: %w", err)
	}

	return ctx, nil
}

// theUserCreatesACheckoutSessionForThePlan creates a checkout session for the plan
func (s *UserSubscriptionSteps) theUserCreatesACheckoutSessionForThePlan(ctx context.Context) (context.Context, error) {
	return s.theUserInitiatesCheckoutForThePlan(ctx)
}

// theCheckoutSessionCompletes completes the checkout
func (s *UserSubscriptionSteps) theCheckoutSessionCompletes(ctx context.Context) (context.Context, error) {
	return s.theCheckoutCompletesSuccessfully(ctx)
}

// theUserCancelsTheSubscription cancels the user's subscription via portal API
func (s *UserSubscriptionSteps) theUserCancelsTheSubscription(ctx context.Context) (context.Context, error) {
	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	result, err := userClient.CancelSubscription(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	ctx = helpers.SetUserCancelResult(ctx, result)

	return ctx, nil
}

// theUserUpgradesToThePlan upgrades the user's subscription plan
func (s *UserSubscriptionSteps) theUserUpgradesToThePlan(ctx context.Context, planName string) (context.Context, error) {
	return s.changeUserPlanTo(ctx, planName)
}

// theUserDowngradesToThePlan downgrades the user's subscription plan
func (s *UserSubscriptionSteps) theUserDowngradesToThePlan(ctx context.Context, planName string) (context.Context, error) {
	return s.changeUserPlanTo(ctx, planName)
}

// changeUserPlanTo is a helper that changes the user's subscription to a specified plan
func (s *UserSubscriptionSteps) changeUserPlanTo(ctx context.Context, planName string) (context.Context, error) {
	infra, _, ok := helpers.GetBillingInfrastructure(ctx)
	if !ok || infra == nil {
		return ctx, fmt.Errorf("no billing infrastructure available")
	}

	planID := infra.PlanIDByName(planName)
	if planID == 0 {
		return ctx, fmt.Errorf("plan '%s' not found", planName)
	}

	periodID := infra.PeriodIDByPlanAndCadence(planName, "monthly")
	if periodID == 0 {
		return ctx, fmt.Errorf("monthly period not found for plan '%s'", planName)
	}

	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	result, err := userClient.ChangePlan(ctx, int(periodID))
	if err != nil {
		return ctx, fmt.Errorf("failed to change plan to %s: %w", planName, err)
	}

	ctx, err = s.handlePlanChangeResult(ctx, planName, result)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

// handlePlanChangeResult processes the ManagementResult from a plan change.
// For Atlos, the portal requires completing a new checkout after a plan change.
// The ManagementResult.Action will be "checkout_required" and the portal cancels
// the old subscription during ExecutePlanChange. For Stripe, the plan change is
// immediate (Action="complete") and no additional step is needed.
func (s *UserSubscriptionSteps) handlePlanChangeResult(ctx context.Context, planName string, result *account.ManagementResult) (context.Context, error) {
	if result == nil || result.Action != "checkout_required" {
		return ctx, nil
	}

	infra, _, ok := helpers.GetBillingInfrastructure(ctx)
	if !ok || infra == nil {
		return ctx, fmt.Errorf("no billing infrastructure available")
	}

	planID := infra.PlanIDByName(planName)
	periodID := infra.PeriodIDByPlanAndCadence(planName, "monthly")

	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	checkoutUI, err := userClient.GetCheckoutUI(ctx, fmt.Sprint(planID),
		account.WithPeriodID(fmt.Sprint(periodID)),
		account.WithGateway(helpers.GetActiveGatewaySDKName(ctx)),
	)
	if err != nil {
		return ctx, fmt.Errorf("failed to get checkout UI for plan change: %w", err)
	}

	ctx = helpers.SetGatewayCheckoutUI(ctx, checkoutUI)

	if checkoutUI.SessionId != nil {
		ctx = helpers.SetGatewayCheckoutSessionID(ctx, *checkoutUI.SessionId)
	}

	ctx, err = s.theCheckoutCompletesSuccessfully(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to complete checkout for plan change to %s: %w", planName, err)
	}

	return ctx, nil
}

// theUsersCurrentPlanIs verifies the user's current subscription plan
func (s *UserSubscriptionSteps) theUsersCurrentPlanIs(ctx context.Context, expectedPlanName string) (context.Context, error) {
	err := helpers.PollUserSubscriptionStatus(ctx, true, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("subscription is not active after plan change to %s: %w", expectedPlanName, err)
	}
	return ctx, nil
}

// aPricingPlanPeriodExists verifies that a pricing plan period exists
func (s *UserSubscriptionSteps) aPricingPlanPeriodExists(ctx context.Context) (context.Context, error) {
	infra, _, ok := helpers.GetBillingInfrastructure(ctx)
	if !ok || infra == nil {
		return ctx, fmt.Errorf("no billing infrastructure available")
	}
	if infra.FirstPeriodID() == 0 {
		return ctx, fmt.Errorf("no pricing plan period found in billing infrastructure")
	}
	return ctx, nil
}

// theUserHasAnActiveSubscription ensures the user has an active subscription
// Uses gateway-agnostic checkout completion via the active gateway mock
func (s *UserSubscriptionSteps) theUserHasAnActiveSubscription(ctx context.Context) (context.Context, error) {
	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	subscriptionStatus, err := userClient.GetSubscriptionStatus(ctx)
	if err == nil && subscriptionStatus.IsSubscribed {
		return ctx, nil
	}

	ctx, err = s.theUserInitiatesCheckoutForThePlan(ctx)
	if err != nil {
		return ctx, err
	}

	return s.theCheckoutCompletesSuccessfully(ctx)
}

// theUserPaysForTheNextBillingPeriod simulates payment via the active gateway mock
func (s *UserSubscriptionSteps) theUserPaysForTheNextBillingPeriod(ctx context.Context) (context.Context, error) {
	subscriptionID, ok := helpers.GetGatewaySubscriptionID(ctx)
	if !ok || subscriptionID == "" {
		return ctx, fmt.Errorf("no subscription ID available")
	}

	if err := helpers.SimulateGatewayPayment(ctx, subscriptionID); err != nil {
		return ctx, fmt.Errorf("failed to simulate payment: %w", err)
	}

	return ctx, nil
}

// theSubscriptionStatusIs verifies the subscription status via portal API
func (s *UserSubscriptionSteps) theSubscriptionStatusIs(ctx context.Context, status string) (context.Context, error) {
	return ctx, helpers.VerifySubscriptionStatus(ctx, status)
}

// theUserHasAnActiveSubscriptionOnThePlan ensures user has active subscription on specific plan
func (s *UserSubscriptionSteps) theUserHasAnActiveSubscriptionOnThePlan(ctx context.Context, planName string) (context.Context, error) {
	infra, _, ok := helpers.GetBillingInfrastructure(ctx)
	if !ok || infra == nil {
		return ctx, fmt.Errorf("no billing infrastructure available")
	}

	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	subscriptionStatus, err := userClient.GetSubscriptionStatus(ctx)
	if err == nil && subscriptionStatus.IsSubscribed {
		return ctx, nil
	}

	planID := infra.PlanIDByName(planName)
	if planID == 0 {
		return ctx, fmt.Errorf("plan '%s' not found", planName)
	}

	periodID := infra.PeriodIDByPlanAndCadence(planName, "monthly")
	if periodID == 0 {
		return ctx, fmt.Errorf("no monthly period found for plan '%s'", planName)
	}

	checkoutUI, err := userClient.GetCheckoutUI(ctx, fmt.Sprint(planID),
		account.WithPeriodID(fmt.Sprint(periodID)),
		account.WithGateway(helpers.GetActiveGatewaySDKName(ctx)),
	)
	if err != nil {
		return ctx, fmt.Errorf("failed to get checkout UI: %w", err)
	}

	ctx = helpers.SetGatewayCheckoutUI(ctx, checkoutUI)

	if checkoutUI.SessionId != nil {
		ctx = helpers.SetGatewayCheckoutSessionID(ctx, *checkoutUI.SessionId)
	}

	ctx, err = s.theCheckoutCompletesSuccessfully(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to complete checkout: %w", err)
	}

	return ctx, nil
}

// theUserSelectsBillingCadence stores billing cadence selection
func (s *UserSubscriptionSteps) theUserSelectsBillingCadence(ctx context.Context, cadence string) (context.Context, error) {
	if cadence != "monthly" && cadence != "yearly" {
		return ctx, fmt.Errorf("invalid cadence '%s': must be 'monthly' or 'yearly'", cadence)
	}
	ctx = helpers.SetBillingCadence(ctx, cadence)
	return ctx, nil
}

// theUserHasAnActiveSubscriptionOnThePlanWithCadence ensures active subscription with cadence
func (s *UserSubscriptionSteps) theUserHasAnActiveSubscriptionOnThePlanWithCadence(ctx context.Context, cadence, planName string) (context.Context, error) {
	ctx, err := s.theUserSelectsBillingCadence(ctx, cadence)
	if err != nil {
		return ctx, err
	}

	infra, _, ok := helpers.GetBillingInfrastructure(ctx)
	if !ok || infra == nil {
		return ctx, fmt.Errorf("no billing infrastructure available")
	}

	userClient, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user client: %w", err)
	}

	subscriptionStatus, err := userClient.GetSubscriptionStatus(ctx)
	if err == nil && subscriptionStatus.IsSubscribed {
		return ctx, nil
	}

	planID := infra.PlanIDByName(planName)
	if planID == 0 {
		return ctx, fmt.Errorf("plan '%s' not found", planName)
	}

	periodID := infra.PeriodIDByPlanAndCadence(planName, cadence)
	if periodID == 0 {
		return ctx, fmt.Errorf("no %s period found for plan '%s'", cadence, planName)
	}

	checkoutUI, err := userClient.GetCheckoutUI(ctx, fmt.Sprint(planID),
		account.WithPeriodID(fmt.Sprint(periodID)),
		account.WithGateway(helpers.GetActiveGatewaySDKName(ctx)),
	)
	if err != nil {
		return ctx, fmt.Errorf("failed to get checkout UI: %w", err)
	}

	ctx = helpers.SetGatewayCheckoutUI(ctx, checkoutUI)

	if checkoutUI.SessionId != nil {
		ctx = helpers.SetGatewayCheckoutSessionID(ctx, *checkoutUI.SessionId)
	}

	ctx, err = s.theCheckoutCompletesSuccessfully(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to complete checkout: %w", err)
	}

	return ctx, nil
}
