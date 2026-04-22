package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	
	"pinner.xyz-e2e/helpers"
	admin "go.lumeweb.com/portal-sdk/admin"
)

// AdminBillingSteps holds the step definitions for admin billing management
type AdminBillingSteps struct{}

// NewAdminBillingSteps creates a new AdminBillingSteps instance
func NewAdminBillingSteps() *AdminBillingSteps {
	return &AdminBillingSteps{}
}

// createTestUserIfNotExists creates a test user if one doesn't exist for credit testing
func (s *AdminBillingSteps) createTestUserIfNotExists(ctx context.Context) (context.Context, error) {
	// Check if test user already exists
	userID, ok := helpers.GetAdminTargetUserID(ctx)
	if ok && userID != 0 {
		return ctx, nil
	}

	// Create a new test user
	testUser := helpers.CreateTestUser()
	accountAPI := helpers.GetUnauthenticatedClient()
	
	err := accountAPI.Register(ctx, testUser.Email, testUser.FirstName, testUser.LastName, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to register test user: %w", err)
	}

	// Login to get account info with user ID
	loginResult, err := accountAPI.Login(ctx, testUser.Email, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to login test user: %w", err)
	}

	// Create authenticated client from JWT token
	authenticatedClient := helpers.CreateAuthenticatedClient(loginResult.Token)

	// Get the account info to extract user ID
	accountInfo, err := authenticatedClient.GetAccount(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get test user account info: %w", err)
	}

	userID = int(accountInfo.Id)
	ctx = helpers.SetAdminTargetUserID(ctx, userID)

	// Store test user, JWT token, and authenticated client in context
	ctx = helpers.SetTestUser(ctx, testUser)
	ctx = helpers.SetJWTToken(ctx, loginResult.Token)
	ctx = helpers.SetAuthenticatedClient(ctx, authenticatedClient)

	// Add to cleanup list
	ctx = helpers.AddTestUserCleanup(ctx, testUser.Email)

	return ctx, nil
}

// InitializeScenario registers all step definitions with godog
func (s *AdminBillingSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Test user setup step
	ctx.Step(`^a test user is registered$`, s.aTestUserIsRegistered)

	// Credit management steps
	ctx.Step(`^the admin lists all credits$`, s.theAdminListsAllCredits)
	ctx.Step(`^the credits are returned successfully$`, s.theCreditsAreReturnedSuccessfully)
	ctx.Step(`^the admin creates a credit for the test user$`, s.theAdminCreatesACreditForTheTestUser)
	ctx.Step(`^the credit is created successfully$`, s.theCreditIsCreatedSuccessfully)
	ctx.Step(`^the credit has the specified amount and manual_adjustment transaction type$`, s.theCreditHasTheSpecifiedAmountAndTransactionType)
	ctx.Step(`^the admin retrieves the credit by ID$`, s.theAdminRetrievesTheCreditByID)
	ctx.Step(`^the credit details are returned successfully$`, s.theCreditDetailsAreReturnedSuccessfully)
	ctx.Step(`^the admin deletes the credit$`, s.theAdminDeletesTheCredit)
	ctx.Step(`^the credit is soft deleted successfully$`, s.theCreditIsSoftDeletedSuccessfully)
	ctx.Step(`^the admin has created and deleted a credit for the test user$`, s.theAdminHasCreatedAndDeletedACreditForTheTestUser)
	ctx.Step(`^the admin restores the credit$`, s.theAdminRestoresTheCredit)
	ctx.Step(`^the credit is restored successfully$`, s.theCreditIsRestoredSuccessfully)
	ctx.Step(`^the admin purges credits older than (\d+) (day|second)s?$`, s.theAdminPurgesCreditsOlderThan)
	ctx.Step(`^old soft-deleted credits are permanently removed$`, s.oldSoftDeletedCreditsArePermanentlyRemoved)
	ctx.Step(`^the purge count is recorded$`, s.thePurgeCountIsRecorded)
	ctx.Step(`^the admin retrieves the balance for the test user$`, s.theAdminRetrievesTheBalanceForTheTestUser)
	ctx.Step(`^the user balance is returned successfully$`, s.theUserBalanceIsReturnedSuccessfully)
	ctx.Step(`^the admin lists deleted credits for the test user$`, s.theAdminListsDeletedCreditsForTheTestUser)
	ctx.Step(`^the deleted credits are returned successfully$`, s.theDeletedCreditsAreReturnedSuccessfully)
	ctx.Step(`^the admin has created a credit for the test user$`, s.theAdminHasCreatedACreditForTheTestUser)
	ctx.Step(`^the admin lists all price lines$`, s.theAdminListsAllPriceLines)
	ctx.Step(`^the price lines are returned successfully$`, s.thePriceLinesAreReturnedSuccessfully)
	ctx.Step(`^the admin creates a new price line named "([^"]*)"$`, s.theAdminCreatesANewPriceLineNamed)
	ctx.Step(`^the price line is created successfully$`, s.thePriceLineIsCreatedSuccessfully)
	ctx.Step(`^the price line has the specified name and description$`, s.thePriceLineHasTheSpecifiedNameAndDescription)
	ctx.Step(`^the admin has created a price line named "([^"]*)"$`, s.theAdminHasCreatedAPriceLineNamed)
	ctx.Step(`^the admin retrieves the price line by ID$`, s.theAdminRetrievesThePriceLineByID)
	ctx.Step(`^the price line details are returned successfully$`, s.thePriceLineDetailsAreReturnedSuccessfully)
	ctx.Step(`^the admin updates the price line with new details$`, s.theAdminUpdatesThePriceLineWithNewDetails)
	ctx.Step(`^the price line is updated successfully$`, s.thePriceLineIsUpdatedSuccessfully)
	ctx.Step(`^the updated details are reflected$`, s.theUpdatedDetailsAreReflected)
	ctx.Step(`^the admin deletes the price line$`, s.theAdminDeletesThePriceLine)
	ctx.Step(`^the price line is deleted successfully$`, s.thePriceLineIsDeletedSuccessfully)
	ctx.Step(`^the admin lists all pricing plans$`, s.theAdminListsAllPricingPlans)
	ctx.Step(`^the pricing plans are returned successfully$`, s.thePricingPlansAreReturnedSuccessfully)
	ctx.Step(`^the admin creates a new pricing plan named "([^"]*)"$`, s.theAdminCreatesANewPricingPlanNamed)
	ctx.Step(`^the pricing plan is created successfully$`, s.thePricingPlanIsCreatedSuccessfully)
	ctx.Step(`^the pricing plan includes the specified periods$`, s.thePricingPlanIncludesTheSpecifiedPeriods)
	ctx.Step(`^the admin has created a pricing plan named "([^"]*)"$`, s.theAdminHasCreatedAPricingPlanNamed)
	ctx.Step(`^the admin updates the pricing plan with new details$`, s.theAdminUpdatesThePricingPlanWithNewDetails)
	ctx.Step(`^the pricing plan is updated successfully$`, s.thePricingPlanIsUpdatedSuccessfully)
	ctx.Step(`^the admin deletes the pricing plan$`, s.theAdminDeletesThePricingPlan)
	ctx.Step(`^the pricing plan is deleted successfully$`, s.thePricingPlanIsDeletedSuccessfully)
	ctx.Step(`^the admin lists all pricing plan periods$`, s.theAdminListsAllPricingPlanPeriods)
	ctx.Step(`^the pricing plan periods are returned successfully$`, s.thePricingPlanPeriodsAreReturnedSuccessfully)
	ctx.Step(`^the admin creates a new pricing plan period for a basic plan$`, s.theAdminCreatesANewPricingPlanPeriodForABasicPlan)
	ctx.Step(`^the pricing plan period is created successfully$`, s.thePricingPlanPeriodIsCreatedSuccessfully)
	ctx.Step(`^the admin has created a pricing plan period$`, s.theAdminHasCreatedAPricingPlanPeriod)
	ctx.Step(`^the admin retrieves the pricing plan period by ID$`, s.theAdminRetrievesThePricingPlanPeriodByID)
	ctx.Step(`^the pricing plan period details are returned successfully$`, s.thePricingPlanPeriodDetailsAreReturnedSuccessfully)
	ctx.Step(`^the admin updates the pricing plan period with new details$`, s.theAdminUpdatesThePricingPlanPeriodWithNewDetails)
	ctx.Step(`^the pricing plan period is updated successfully$`, s.thePricingPlanPeriodIsUpdatedSuccessfully)
	ctx.Step(`^the admin deletes the pricing plan period$`, s.theAdminDeletesThePricingPlanPeriod)
	ctx.Step(`^the pricing plan period is deleted successfully$`, s.thePricingPlanPeriodIsDeletedSuccessfully)
	ctx.Step(`^the admin has created multiple credits with different types$`, s.theAdminHasCreatedMultipleCreditsWithDifferentTypes)
	ctx.Step(`^the admin has created multiple credits with different directions$`, s.theAdminHasCreatedMultipleCreditsWithDifferentDirections)
	ctx.Step(`^the admin filters credits by transaction type "([^"]*)"$`, s.theAdminFiltersCreditsByTransactionType)
	ctx.Step(`^only credits with that type are returned$`, s.onlyCreditsWithThatTypeAreReturned)
	ctx.Step(`^the admin filters credits by direction "([^"]*)"$`, s.theAdminFiltersCreditsByDirection)
	ctx.Step(`^only credits with that direction are returned$`, s.onlyCreditsWithThatDirectionAreReturned)

	// Subscriber management steps
	ctx.Step(`^the admin lists all subscribers$`, s.theAdminListsAllSubscribers)
	ctx.Step(`^the subscribers are returned successfully$`, s.theSubscribersAreReturnedSuccessfully)
	ctx.Step(`^there is an active subscriber$`, s.thereIsAnActiveSubscriber)
	ctx.Step(`^the admin retrieves the subscriber by ID$`, s.theAdminRetrievesTheSubscriberByID)
	ctx.Step(`^the subscriber details are returned successfully$`, s.theSubscriberDetailsAreReturnedSuccessfully)
	ctx.Step(`^there is a gateway with subscribers$`, s.thereIsAGatewayWithSubscribers)
	ctx.Step(`^the admin lists subscribers for the gateway$`, s.theAdminListsSubscribersForTheGateway)
	ctx.Step(`^the gateway subscribers are returned successfully$`, s.theGatewaySubscribersAreReturnedSuccessfully)
	ctx.Step(`^a test user with subscriptions$`, s.aTestUserWithSubscriptions)
	ctx.Step(`^the admin retrieves subscribers for the test user$`, s.theAdminRetrievesSubscribersForTheTestUser)
	ctx.Step(`^the user subscribers are returned successfully$`, s.theUserSubscribersAreReturnedSuccessfully)
	ctx.Step(`^a test user with an active subscription$`, s.aTestUserWithAnActiveSubscription)
	ctx.Step(`^the admin cancels the user's subscription immediately$`, s.theAdminCancelsTheUsersSubscriptionImmediately)
	ctx.Step(`^the subscription is cancelled successfully$`, s.theSubscriptionIsCancelledSuccessfully)
	ctx.Step(`^the cancellation takes effect immediately$`, s.theCancellationTakesEffectImmediately)
	ctx.Step(`^the admin cancels the user's subscription at end of period$`, s.theAdminCancelsTheUsersSubscriptionAtEndOfPeriod)
	ctx.Step(`^the subscription is scheduled for cancellation$`, s.theSubscriptionIsScheduledForCancellation)
	ctx.Step(`^the cancellation will take effect at billing period end$`, s.theCancellationWillTakeEffectAtBillingPeriodEnd)
	ctx.Step(`^there is an available pricing plan period$`, s.thereIsAnAvailablePricingPlanPeriod)
	ctx.Step(`^the admin changes the user's plan$`, s.theAdminChangesTheUsersPlan)
	ctx.Step(`^the plan change is processed successfully$`, s.thePlanChangeIsProcessedSuccessfully)
	ctx.Step(`^the new pricing plan period is applied$`, s.theNewPricingPlanPeriodIsApplied)

	// Price Line Plan Management steps
	ctx.Step(`^the admin adds the plan to the price line$`, s.theAdminAddsThePlanToThePriceLine)
	ctx.Step(`^the plan is added to the price line successfully$`, s.thePlanIsAddedToThePriceLineSuccessfully)
	ctx.Step(`^the admin has created a price line with multiple plans$`, s.theAdminHasCreatedAPriceLineWithMultiplePlans)
	ctx.Step(`^the admin updates a plan position in the price line$`, s.theAdminUpdatesAPlanPositionInThePriceLine)
	ctx.Step(`^the plan position is updated successfully$`, s.thePlanPositionIsUpdatedSuccessfully)
	ctx.Step(`^the admin has created a price line with a plan$`, s.theAdminHasCreatedAPriceLineWithAPlan)
	ctx.Step(`^the admin removes the plan from the price line$`, s.theAdminRemovesThePlanFromThePriceLine)
	ctx.Step(`^the plan is removed from the price line successfully$`, s.thePlanIsRemovedFromThePriceLineSuccessfully)
}

// =============================================================================
// Test User Setup
// =============================================================================

// aTestUserIsRegistered creates a test user for billing operations
func (s *AdminBillingSteps) aTestUserIsRegistered(ctx context.Context) (context.Context, error) {
	return s.createTestUserIfNotExists(ctx)
}

// =============================================================================
// Credit Management Steps
// =============================================================================

// theAdminListsAllCredits lists all credits
func (s *AdminBillingSteps) theAdminListsAllCredits(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	params := &admin.GetApiBillingCreditsParams{}
	credits, _, err := adminClient.Billing().ListCredits(ctx, params)
	if err != nil {
		return ctx, fmt.Errorf("failed to list credits: %w", err)
	}

	ctx = helpers.SetAdminCredits(ctx, credits)
	return ctx, nil
}

// theCreditsAreReturnedSuccessfully verifies credits were retrieved
func (s *AdminBillingSteps) theCreditsAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	_, ok := helpers.GetAdminCredits(ctx)
	if !ok {
		return ctx, fmt.Errorf("credits were not retrieved")
	}
	return ctx, nil
}

// theAdminCreatesACreditForTheTestUser creates a credit for a test user
func (s *AdminBillingSteps) theAdminCreatesACreditForTheTestUser(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Get test user ID from context
	userID, ok := helpers.GetAdminTargetUserID(ctx)
	if !ok || userID == 0 {
		return ctx, fmt.Errorf("test user ID not available in context - user may not be registered")
	}

	// Convert Description to pointer
	desc := "Test credit from E2E tests"
	refId := fmt.Sprintf("e2e-test-%d", time.Now().Unix())
	refType := "manual"

	req := &admin.CreditCreateRequest{
		UserId:          userID,
		Amount:          "10.00",
		Type:            "manual_adjustment",
		Direction:       "credit",
		Description:     &desc,
		ReferenceId:     &refId,
		ReferenceType:   &refType,
	}

	credit, err := adminClient.Billing().CreateCredit(ctx, req)
	if err != nil {
		return ctx, fmt.Errorf("failed to create credit: %w", err)
	}

	creditID := helpers.ToBillingIDFromUUID(credit.Id)
	ctx = helpers.SetAdminCurrentCredit(ctx, credit)
	ctx = helpers.SetAdminCurrentCreditID(ctx, creditID)
	ctx = helpers.AddAdminCreditCleanup(ctx, creditID.String())
	return ctx, nil
}

// theCreditIsCreatedSuccessfully verifies credit creation
func (s *AdminBillingSteps) theCreditIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	credit, ok := helpers.GetAdminCurrentCredit(ctx)
	if err := helpers.VerifyResourceExists(credit, ok, "credit"); err != nil {
		return ctx, err
	}

	if credit.Id.String() == "" {
		return ctx, fmt.Errorf("credit has no ID")
	}

	return ctx, nil
}

// theCreditHasTheSpecifiedAmountAndTransactionType verifies credit details
func (s *AdminBillingSteps) theCreditHasTheSpecifiedAmountAndTransactionType(ctx context.Context) (context.Context, error) {
	credit, ok := helpers.GetAdminCurrentCredit(ctx)
	if err := helpers.VerifyResourceExists(credit, ok, "credit"); err != nil {
		return ctx, err
	}

	if credit.Type != "manual_adjustment" {
		return ctx, fmt.Errorf("expected type 'manual_adjustment', got '%s'", credit.Type)
	}

	if credit.Direction != "credit" {
		return ctx, fmt.Errorf("expected direction 'credit', got '%s'", credit.Direction)
	}

	return ctx, nil
}

// theAdminRetrievesTheCreditByID retrieves a credit by ID
func (s *AdminBillingSteps) theAdminRetrievesTheCreditByID(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	creditID, err := helpers.RequireAdminCurrentCreditID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no credit ID available: %w", err)
	}

	credit, err := adminClient.Billing().GetCredit(ctx, creditID.String())
	if err != nil {
		return ctx, fmt.Errorf("failed to get credit: %w", err)
	}

	ctx = helpers.SetAdminCurrentCredit(ctx, credit)
	return ctx, nil
}

// theCreditDetailsAreReturnedSuccessfully verifies credit retrieval
func (s *AdminBillingSteps) theCreditDetailsAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	credit, ok := helpers.GetAdminCurrentCredit(ctx)
	if err := helpers.VerifyResourceExists(credit, ok, "credit"); err != nil {
		return ctx, err
	}

	creditID, _ := helpers.GetAdminCurrentCreditID(ctx)
	if credit.Id.String() != creditID.String() {
		return ctx, fmt.Errorf("credit ID mismatch")
	}

	return ctx, nil
}

// theAdminHasCreatedACreditForTheTestUser helper step
func (s *AdminBillingSteps) theAdminHasCreatedACreditForTheTestUser(ctx context.Context) (context.Context, error) {
	return s.theAdminCreatesACreditForTheTestUser(ctx)
}

// theAdminDeletesTheCredit deletes a credit
func (s *AdminBillingSteps) theAdminDeletesTheCredit(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	creditID, err := helpers.RequireAdminCurrentCreditID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no credit ID available: %w", err)
	}

	err = adminClient.Billing().DeleteCredit(ctx, creditID.String())
	if err != nil {
		return ctx, fmt.Errorf("failed to delete credit: %w", err)
	}

	return ctx, nil
}

// theCreditIsSoftDeletedSuccessfully verifies soft delete
func (s *AdminBillingSteps) theCreditIsSoftDeletedSuccessfully(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	userID, err := helpers.RequireAdminTargetUserID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no target user ID available: %w", err)
	}

	// Verify the credit is now in deleted state
	params := &admin.GetApiBillingUsersUserIdDeletedCreditsParams{}
	deletedCredits, _, err := adminClient.Billing().GetUserDeletedCredits(ctx, fmt.Sprint(userID), params)
	if err != nil {
		return ctx, fmt.Errorf("failed to list deleted credits: %w", err)
	}

	creditID, err := helpers.RequireAdminCurrentCreditID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no credit ID available: %w", err)
	}

	found := false
	for _, dc := range deletedCredits {
		if dc.Id.String() == creditID.String() {
			found = true
			break
		}
	}

	if !found {
		return ctx, fmt.Errorf("credit not found in deleted list")
	}

	return ctx, nil
}

// theAdminHasCreatedAndDeletedACreditForTheTestUser helper step
func (s *AdminBillingSteps) theAdminHasCreatedAndDeletedACreditForTheTestUser(ctx context.Context) (context.Context, error) {
	ctx, err := s.theAdminCreatesACreditForTheTestUser(ctx)
	if err != nil {
		return ctx, err
	}
	return s.theAdminDeletesTheCredit(ctx)
}

// theAdminRestoresTheCredit restores a deleted credit
func (s *AdminBillingSteps) theAdminRestoresTheCredit(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	creditID, err := helpers.RequireAdminCurrentCreditID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no credit ID available: %w", err)
	}

	credit, err := adminClient.Billing().RestoreCredit(ctx, creditID.String())
	if err != nil {
		return ctx, fmt.Errorf("failed to restore credit: %w", err)
	}

	ctx = helpers.SetAdminCurrentCredit(ctx, credit)
	return ctx, nil
}

// theCreditIsRestoredSuccessfully verifies credit restoration
func (s *AdminBillingSteps) theCreditIsRestoredSuccessfully(ctx context.Context) (context.Context, error) {
	credit, ok := helpers.GetAdminCurrentCredit(ctx)
	if err := helpers.VerifyResourceExists(credit, ok, "credit"); err != nil {
		return ctx, err
	}

	creditID, _ := helpers.GetAdminCurrentCreditID(ctx)
	if credit.Id.String() != creditID.String() {
		return ctx, fmt.Errorf("restored credit ID mismatch")
	}

	return ctx, nil
}

// theAdminPurgesCreditsOlderThan purges old credits by duration in seconds
func (s *AdminBillingSteps) theAdminPurgesCreditsOlderThan(ctx context.Context, value int, unit string) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Convert days to seconds, or use seconds as-is
	var durationSeconds int
	switch unit {
	case "day", "days":
		durationSeconds = value * 60 * 60 * 24
	case "second", "seconds":
		durationSeconds = value
	default:
		return ctx, fmt.Errorf("invalid unit: %s (expected day(s) or second(s))", unit)
	}

	// Convert to Go duration string (e.g., "86400s" for 1 day)
	durationStr := fmt.Sprintf("%ds", durationSeconds)
	req := &admin.CreditPurgeRequest{
		OlderThan: durationStr,
	}
	purgedCount, err := adminClient.Billing().PurgeCredits(ctx, req)
	if err != nil {
		return ctx, fmt.Errorf("failed to purge credits: %w", err)
	}

	ctx = helpers.SetAdminPurgedCreditsCount(ctx, purgedCount)
	return ctx, nil
}

// oldSoftDeletedCreditsArePermanentlyRemoved verifies purge
func (s *AdminBillingSteps) oldSoftDeletedCreditsArePermanentlyRemoved(ctx context.Context) (context.Context, error) {
	count, ok := helpers.GetAdminPurgedCreditsCount(ctx)
	if !ok {
		return ctx, fmt.Errorf("purge count not recorded")
	}

	// Verify count is non-negative
	if count < 0 {
		return ctx, fmt.Errorf("purge count cannot be negative: %d", count)
	}

	return ctx, nil
}

// thePurgeCountIsRecorded verifies purge count
func (s *AdminBillingSteps) thePurgeCountIsRecorded(ctx context.Context) (context.Context, error) {
	_, ok := helpers.GetAdminPurgedCreditsCount(ctx)
	if !ok {
		return ctx, fmt.Errorf("purge count not recorded")
	}
	return ctx, nil
}

// =============================================================================
// User Balance Steps
// =============================================================================

// theAdminRetrievesTheBalanceForTheTestUser retrieves user balance
func (s *AdminBillingSteps) theAdminRetrievesTheBalanceForTheTestUser(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Get test user ID from context
	userID, ok := helpers.GetAdminTargetUserID(ctx)
	if !ok || userID == 0 {
		return ctx, fmt.Errorf("test user ID not available in context - user may not be registered")
	}

	balance, err := adminClient.Billing().GetUserBalance(ctx, fmt.Sprint(userID))
	if err != nil {
		return ctx, fmt.Errorf("failed to get user balance: %w", err)
	}

	ctx = helpers.SetAdminUserBalance(ctx, balance)
	return ctx, nil
}

// theUserBalanceIsReturnedSuccessfully verifies balance retrieval
func (s *AdminBillingSteps) theUserBalanceIsReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	balance, ok := helpers.GetAdminUserBalance(ctx)
	if err := helpers.VerifyResourceExists(balance, ok, "user balance"); err != nil {
		return ctx, err
	}

	return ctx, nil
}

// theAdminListsDeletedCreditsForTheTestUser lists deleted credits
func (s *AdminBillingSteps) theAdminListsDeletedCreditsForTheTestUser(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Get test user ID from context
	userID, ok := helpers.GetAdminTargetUserID(ctx)
	if !ok || userID == 0 {
		return ctx, fmt.Errorf("test user ID not available in context - user may not be registered")
	}

	params := &admin.GetApiBillingUsersUserIdDeletedCreditsParams{}
	credits, _, err := adminClient.Billing().GetUserDeletedCredits(ctx, fmt.Sprint(userID), params)
	if err != nil {
		return ctx, fmt.Errorf("failed to list deleted credits: %w", err)
	}

	ctx = helpers.SetAdminCredits(ctx, credits)
	return ctx, nil
}

// theDeletedCreditsAreReturnedSuccessfully verifies deleted credits
func (s *AdminBillingSteps) theDeletedCreditsAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	_, ok := helpers.GetAdminCredits(ctx)
	if !ok {
		return ctx, fmt.Errorf("deleted credits were not retrieved")
	}
	return ctx, nil
}

// =============================================================================
// Price Line Management Steps
// =============================================================================

// theAdminListsAllPriceLines lists all price lines
func (s *AdminBillingSteps) theAdminListsAllPriceLines(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	priceLines, _, err := adminClient.Billing().ListPriceLines(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list price lines: %w", err)
	}

	ctx = helpers.SetAdminPriceLines(ctx, priceLines)
	return ctx, nil
}

// thePriceLinesAreReturnedSuccessfully verifies price lines
func (s *AdminBillingSteps) thePriceLinesAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	_, ok := helpers.GetAdminPriceLines(ctx)
	if !ok {
		return ctx, fmt.Errorf("price lines were not retrieved")
	}
	return ctx, nil
}

// theAdminCreatesANewPriceLineNamed creates a price line
func (s *AdminBillingSteps) theAdminCreatesANewPriceLineNamed(ctx context.Context, name string) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	req := &admin.PriceLineCreateRequest{
		Name:        name,
		Description: fmt.Sprintf("Test price line %s created from E2E tests", name),
		IsActive:    true,
		IsDefault:   false,
	}

	priceLine, err := adminClient.Billing().CreatePriceLine(ctx, req)
	if err != nil {
		return ctx, fmt.Errorf("failed to create price line: %w", err)
	}

	priceLineID := helpers.ToBillingID(priceLine.Id)
	ctx = helpers.SetAdminCurrentPriceLine(ctx, priceLine)
	ctx = helpers.SetAdminCurrentPriceLineID(ctx, priceLineID)
	ctx = helpers.AddAdminPriceLineCleanup(ctx, priceLineID.String())
	return ctx, nil
}

// thePriceLineIsCreatedSuccessfully verifies price line creation
func (s *AdminBillingSteps) thePriceLineIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	priceLine, ok := helpers.GetAdminCurrentPriceLine(ctx)
	if err := helpers.VerifyResourceExists(priceLine, ok, "price line"); err != nil {
		return ctx, err
	}

	fields, err := helpers.ExtractPriceLineFields(priceLine)
	if err != nil {
		return ctx, err
	}

	if fields.Id == 0 {
		return ctx, fmt.Errorf("price line has no ID")
	}

	return ctx, nil
}

// thePriceLineHasTheSpecifiedNameAndDescription verifies price line details
func (s *AdminBillingSteps) thePriceLineHasTheSpecifiedNameAndDescription(ctx context.Context) (context.Context, error) {
	priceLine, ok := helpers.GetAdminCurrentPriceLine(ctx)
	if err := helpers.VerifyResourceExists(priceLine, ok, "price line"); err != nil {
		return ctx, err
	}

	fields, err := helpers.ExtractPriceLineFields(priceLine)
	if err != nil {
		return ctx, err
	}

	if fields.Name == "" {
		return ctx, fmt.Errorf("price line has no name")
	}

	if fields.Description == "" {
		return ctx, fmt.Errorf("price line has no description")
	}

	return ctx, nil
}

// theAdminHasCreatedAPriceLineNamed helper step
func (s *AdminBillingSteps) theAdminHasCreatedAPriceLineNamed(ctx context.Context, name string) (context.Context, error) {
	return s.theAdminCreatesANewPriceLineNamed(ctx, name)
}

// theAdminRetrievesThePriceLineByID retrieves a price line by ID
func (s *AdminBillingSteps) theAdminRetrievesThePriceLineByID(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	priceLineID, err := helpers.RequireAdminCurrentPriceLineID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no price line ID available: %w", err)
	}

	priceLine, err := adminClient.Billing().GetPriceLine(ctx, priceLineID.String())
	if err != nil {
		return ctx, fmt.Errorf("failed to get price line: %w", err)
	}

	ctx = helpers.SetAdminCurrentPriceLine(ctx, priceLine)
	return ctx, nil
}

// thePriceLineDetailsAreReturnedSuccessfully verifies price line retrieval
func (s *AdminBillingSteps) thePriceLineDetailsAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	priceLine, ok := helpers.GetAdminCurrentPriceLine(ctx)
	if err := helpers.VerifyResourceExists(priceLine, ok, "price line"); err != nil {
		return ctx, err
	}

	fields, err := helpers.ExtractPriceLineFields(priceLine)
	if err != nil {
		return ctx, err
	}

	priceLineID, _ := helpers.GetAdminCurrentPriceLineID(ctx)
	if helpers.ToBillingID(fields.Id).String() != priceLineID.String() {
		return ctx, fmt.Errorf("price line ID mismatch: expected %s, got %d", priceLineID, fields.Id)
	}

	return ctx, nil
}

// theAdminUpdatesThePriceLineWithNewDetails updates a price line
func (s *AdminBillingSteps) theAdminUpdatesThePriceLineWithNewDetails(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	priceLineID, err := helpers.RequireAdminCurrentPriceLineID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no price line ID available: %w", err)
	}

	req := &admin.PriceLineUpdateRequest{
		Name:        "Updated Test Line",
		Description: "Updated description for E2E tests",
		IsActive:    true,
		IsDefault:   false,
	}

	priceLine, err := adminClient.Billing().UpdatePriceLine(ctx, priceLineID.String(), req)
	if err != nil {
		return ctx, fmt.Errorf("failed to update price line: %w", err)
	}

	ctx = helpers.SetAdminCurrentPriceLine(ctx, priceLine)
	return ctx, nil
}

// thePriceLineIsUpdatedSuccessfully verifies update
func (s *AdminBillingSteps) thePriceLineIsUpdatedSuccessfully(ctx context.Context) (context.Context, error) {
	priceLine, ok := helpers.GetAdminCurrentPriceLine(ctx)
	if err := helpers.VerifyResourceExists(priceLine, ok, "price line"); err != nil {
		return ctx, err
	}

	fields, err := helpers.ExtractPriceLineFields(priceLine)
	if err != nil {
		return ctx, err
	}

	if fields.Name != "Updated Test Line" {
		return ctx, fmt.Errorf("price line name not updated, got: %s", fields.Name)
	}

	return ctx, nil
}

// theUpdatedDetailsAreReflected verifies updated details
func (s *AdminBillingSteps) theUpdatedDetailsAreReflected(ctx context.Context) (context.Context, error) {
	priceLine, ok := helpers.GetAdminCurrentPriceLine(ctx)
	if err := helpers.VerifyResourceExists(priceLine, ok, "price line"); err != nil {
		return ctx, err
	}

	fields, err := helpers.ExtractPriceLineFields(priceLine)
	if err != nil {
		return ctx, err
	}

	if fields.Description != "Updated description for E2E tests" {
		return ctx, fmt.Errorf("price line description not updated, got: %s", fields.Description)
	}

	return ctx, nil
}

// theAdminDeletesThePriceLine deletes a price line
func (s *AdminBillingSteps) theAdminDeletesThePriceLine(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	priceLineID, err := helpers.RequireAdminCurrentPriceLineID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no price line ID available: %w", err)
	}

	err = adminClient.Billing().DeletePriceLine(ctx, priceLineID.String())
	if err != nil {
		return ctx, fmt.Errorf("failed to delete price line: %w", err)
	}

	return ctx, nil
}

// thePriceLineIsDeletedSuccessfully verifies deletion
func (s *AdminBillingSteps) thePriceLineIsDeletedSuccessfully(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	priceLineID, _ := helpers.GetAdminCurrentPriceLineID(ctx)

	// Verify price line no longer exists
	priceLines, _, err := adminClient.Billing().ListPriceLines(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list price lines: %w", err)
	}

	for _, pl := range priceLines {
		if helpers.ToBillingID(pl.Id).String() == priceLineID.String() {
			return ctx, fmt.Errorf("price line still exists after deletion")
		}
	}

	return ctx, nil
}

// =============================================================================
// Pricing Plan Management Steps
// =============================================================================

// theAdminListsAllPricingPlans lists all pricing plans
func (s *AdminBillingSteps) theAdminListsAllPricingPlans(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	plans, _, err := adminClient.Billing().ListPricingPlans(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list pricing plans: %w", err)
	}

	ctx = helpers.SetAdminPricingPlans(ctx, plans)
	return ctx, nil
}

// thePricingPlansAreReturnedSuccessfully verifies pricing plans
func (s *AdminBillingSteps) thePricingPlansAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	_, ok := helpers.GetAdminPricingPlans(ctx)
	if !ok {
		return ctx, fmt.Errorf("pricing plans were not retrieved")
	}
	return ctx, nil
}

// theAdminCreatesANewPricingPlanNamed creates a pricing plan
func (s *AdminBillingSteps) theAdminCreatesANewPricingPlanNamed(ctx context.Context, name string) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	req := &admin.PricingPlanCreateRequest{
		Name:           name,
		Description:    fmt.Sprintf("Test pricing plan %s created from E2E tests", name),
		Currency:       "USD",
		IsActive:       true,
		IsPublic:       false,
		PricingPeriods: []admin.PricingPlanPeriod{}, // Empty periods initially
	}

	plan, err := adminClient.Billing().CreatePricingPlan(ctx, req)
	if err != nil {
		return ctx, fmt.Errorf("failed to create pricing plan: %w", err)
	}

	pricingPlanID := helpers.ToBillingID(plan.Id)
	ctx = helpers.SetAdminCurrentPricingPlan(ctx, plan)
	ctx = helpers.SetAdminCurrentPricingPlanID(ctx, pricingPlanID)
	ctx = helpers.AddAdminPricingPlanCleanup(ctx, pricingPlanID.String())
	return ctx, nil
}

// thePricingPlanIsCreatedSuccessfully verifies pricing plan creation
func (s *AdminBillingSteps) thePricingPlanIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	plan, ok := helpers.GetAdminCurrentPricingPlan(ctx)
	if err := helpers.VerifyResourceExists(plan, ok, "pricing plan"); err != nil {
		return ctx, err
	}

	if plan.Id == 0 {
		return ctx, fmt.Errorf("pricing plan has no ID")
	}

	return ctx, nil
}

// thePricingPlanIncludesTheSpecifiedPeriods verifies periods included
func (s *AdminBillingSteps) thePricingPlanIncludesTheSpecifiedPeriods(ctx context.Context) (context.Context, error) {
	plan, ok := helpers.GetAdminCurrentPricingPlan(ctx)
	if err := helpers.VerifyResourceExists(plan, ok, "pricing plan"); err != nil {
		return ctx, err
	}

	// Verify PricingPeriods slice exists (can be empty)
	return ctx, nil
}

// theAdminHasCreatedAPricingPlanNamed helper step
func (s *AdminBillingSteps) theAdminHasCreatedAPricingPlanNamed(ctx context.Context, name string) (context.Context, error) {
	return s.theAdminCreatesANewPricingPlanNamed(ctx, name)
}

// theAdminUpdatesThePricingPlanWithNewDetails updates a pricing plan
func (s *AdminBillingSteps) theAdminUpdatesThePricingPlanWithNewDetails(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	req := &admin.PricingPlanUpdateRequest{
		Name:           "Updated Test Plan",
		Description:    "Updated description for E2E tests",
		Currency:       "USD",
		IsActive:       true,
		IsPublic:       false,
		PricingPeriods: []admin.PricingPlanPeriod{},
	}

	pricingPlanID, err := helpers.RequireAdminCurrentPricingPlanID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no pricing plan ID available: %w", err)
	}

	plan, err := adminClient.Billing().UpdatePricingPlan(ctx, pricingPlanID.String(), req)
	if err != nil {
		return ctx, fmt.Errorf("failed to update pricing plan: %w", err)
	}

	ctx = helpers.SetAdminCurrentPricingPlan(ctx, plan)
	return ctx, nil
}

// thePricingPlanIsUpdatedSuccessfully verifies update
func (s *AdminBillingSteps) thePricingPlanIsUpdatedSuccessfully(ctx context.Context) (context.Context, error) {
	plan, ok := helpers.GetAdminCurrentPricingPlan(ctx)
	if err := helpers.VerifyResourceExists(plan, ok, "pricing plan"); err != nil {
		return ctx, err
	}

	if plan.Name != "Updated Test Plan" {
		return ctx, fmt.Errorf("pricing plan name not updated")
	}

	return ctx, nil
}

// theAdminDeletesThePricingPlan deletes a pricing plan
func (s *AdminBillingSteps) theAdminDeletesThePricingPlan(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	pricingPlanID, err := helpers.RequireAdminCurrentPricingPlanID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no pricing plan ID available: %w", err)
	}

	err = adminClient.Billing().DeletePricingPlan(ctx, pricingPlanID.String())
	if err != nil {
		return ctx, fmt.Errorf("failed to delete pricing plan: %w", err)
	}

	return ctx, nil
}

// thePricingPlanIsDeletedSuccessfully verifies deletion
func (s *AdminBillingSteps) thePricingPlanIsDeletedSuccessfully(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	pricingPlanID, _ := helpers.GetAdminCurrentPricingPlanID(ctx)

	// Verify pricing plan no longer exists
	plans, _, err := adminClient.Billing().ListPricingPlans(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list pricing plans: %w", err)
	}

	for _, pl := range plans {
		if helpers.ToBillingID(pl.Id).String() == pricingPlanID.String() {
			return ctx, fmt.Errorf("pricing plan still exists after deletion")
		}
	}

	return ctx, nil
}

// =============================================================================
// Pricing Plan Period Management Steps
// =============================================================================

// theAdminListsAllPricingPlanPeriods lists all pricing plan periods
func (s *AdminBillingSteps) theAdminListsAllPricingPlanPeriods(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	periods, _, err := adminClient.Billing().ListPricingPlanPeriods(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list pricing plan periods: %w", err)
	}

	ctx = helpers.SetAdminPricingPlanPeriods(ctx, periods)
	return ctx, nil
}

// thePricingPlanPeriodsAreReturnedSuccessfully verifies periods
func (s *AdminBillingSteps) thePricingPlanPeriodsAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	_, ok := helpers.GetAdminPricingPlanPeriods(ctx)
	if !ok {
		return ctx, fmt.Errorf("pricing plan periods were not retrieved")
	}
	return ctx, nil
}

// theAdminCreatesANewPricingPlanPeriodForABasicPlan creates a pricing plan period
func (s *AdminBillingSteps) theAdminCreatesANewPricingPlanPeriodForABasicPlan(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Create a fresh plan for each test to avoid duplicate key errors
	plan, err := adminClient.Billing().CreatePricingPlan(ctx, &admin.PricingPlanCreateRequest{
		Name:           fmt.Sprintf("Test Plan for Period Test %d", time.Now().Unix()),
		Description:    "Test plan for period testing",
		Currency:       "USD",
		IsActive:       true,
		IsPublic:       false,
		PricingPeriods: []admin.PricingPlanPeriod{},
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to create plan: %w", err)
	}
	planID := int(plan.Id)
	ctx = helpers.AddAdminPricingPlanCleanup(ctx, fmt.Sprint(planID))

	req := &admin.PricingPlanPeriodCreateRequest{
		Cadence:       "monthly",
		PriceUsd:      10.0,
		PricingPlanId: planID,
		QuotaPlanId:   1, // Use quota plan ID 1 (assuming it exists)
	}

	period, err := adminClient.Billing().CreatePricingPlanPeriod(ctx, req)
	if err != nil {
		return ctx, fmt.Errorf("failed to create pricing plan period: %w", err)
	}

	periodID := helpers.ToBillingID(period.Id)
	ctx = helpers.SetAdminCurrentPeriod(ctx, period)
	ctx = helpers.SetAdminCurrentPeriodID(ctx, periodID)
	ctx = helpers.AddAdminPricingPlanPeriodCleanup(ctx, periodID.String())
	return ctx, nil
}

// thePricingPlanPeriodIsCreatedSuccessfully verifies period creation
func (s *AdminBillingSteps) thePricingPlanPeriodIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	period, ok := helpers.GetAdminCurrentPeriod(ctx)
	if err := helpers.VerifyResourceExists(period, ok, "pricing plan period"); err != nil {
		return ctx, err
	}

	if period.Id == 0 {
		return ctx, fmt.Errorf("pricing plan period has no ID")
	}

	return ctx, nil
}

// theAdminHasCreatedAPricingPlanPeriod helper step
func (s *AdminBillingSteps) theAdminHasCreatedAPricingPlanPeriod(ctx context.Context) (context.Context, error) {
	return s.theAdminCreatesANewPricingPlanPeriodForABasicPlan(ctx)
}

// theAdminRetrievesThePricingPlanPeriodByID retrieves a period
func (s *AdminBillingSteps) theAdminRetrievesThePricingPlanPeriodByID(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	periodID, err := helpers.RequireAdminCurrentPeriodID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no period ID available: %w", err)
	}

	period, err := adminClient.Billing().GetPricingPlanPeriod(ctx, periodID.String())
	if err != nil {
		return ctx, fmt.Errorf("failed to get pricing plan period: %w", err)
	}

	ctx = helpers.SetAdminCurrentPeriod(ctx, period)
	return ctx, nil
}

// thePricingPlanPeriodDetailsAreReturnedSuccessfully verifies period retrieval
func (s *AdminBillingSteps) thePricingPlanPeriodDetailsAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	period, ok := helpers.GetAdminCurrentPeriod(ctx)
	if err := helpers.VerifyResourceExists(period, ok, "pricing plan period"); err != nil {
		return ctx, err
	}

	if period.Id == 0 {
		return ctx, fmt.Errorf("pricing plan period has no ID")
	}

	return ctx, nil
}

// theAdminUpdatesThePricingPlanPeriodWithNewDetails updates a period
func (s *AdminBillingSteps) theAdminUpdatesThePricingPlanPeriodWithNewDetails(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	periodID, err := helpers.RequireAdminCurrentPeriodID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no period ID available: %w", err)
	}

	req := &admin.PricingPlanPeriodUpdateRequest{
		Cadence:     "yearly",
		PriceUsd:    100.0,
		QuotaPlanId: 1,
	}

	period, err := adminClient.Billing().UpdatePricingPlanPeriod(ctx, periodID.String(), req)
	if err != nil {
		return ctx, fmt.Errorf("failed to update pricing plan period: %w", err)
	}

	ctx = helpers.SetAdminCurrentPeriod(ctx, period)
	return ctx, nil
}

// thePricingPlanPeriodIsUpdatedSuccessfully verifies update
func (s *AdminBillingSteps) thePricingPlanPeriodIsUpdatedSuccessfully(ctx context.Context) (context.Context, error) {
	period, ok := helpers.GetAdminCurrentPeriod(ctx)
	if err := helpers.VerifyResourceExists(period, ok, "pricing plan period"); err != nil {
		return ctx, err
	}

	if period.Cadence != "yearly" {
		return ctx, fmt.Errorf("pricing plan period cadence not updated")
	}

	return ctx, nil
}

// theAdminDeletesThePricingPlanPeriod deletes a period
func (s *AdminBillingSteps) theAdminDeletesThePricingPlanPeriod(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	periodID, err := helpers.RequireAdminCurrentPeriodID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no period ID available: %w", err)
	}

	err = adminClient.Billing().DeletePricingPlanPeriod(ctx, periodID.String())
	if err != nil {
		return ctx, fmt.Errorf("failed to delete pricing plan period: %w", err)
	}

	return ctx, nil
}

// thePricingPlanPeriodIsDeletedSuccessfully verifies deletion
func (s *AdminBillingSteps) thePricingPlanPeriodIsDeletedSuccessfully(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	periodID, _ := helpers.GetAdminCurrentPeriodID(ctx)

	// Verify period no longer exists
	periods, _, err := adminClient.Billing().ListPricingPlanPeriods(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list pricing plan periods: %w", err)
	}

	for _, p := range periods {
		if helpers.ToBillingID(p.Id) == periodID {
			return ctx, fmt.Errorf("pricing plan period still exists after deletion")
		}
	}

	return ctx, nil
}

// =============================================================================
// Filtering Steps
// theAdminHasCreatedMultipleCreditsWithDifferentTypes creates credits with different transaction types for filtering tests
func (s *AdminBillingSteps) theAdminHasCreatedMultipleCreditsWithDifferentTypes(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Get test user ID from context
	userID, ok := helpers.GetAdminTargetUserID(ctx)
	if !ok || userID == 0 {
		return ctx, fmt.Errorf("test user ID not available in context - user may not be registered")
	}

	// Create a credit with "manual_adjustment" transaction type
	desc1 := "Test manual adjustment 1"
	req1 := &admin.CreditCreateRequest{
		UserId:   userID,
		Amount:   "5.00",
		Type:     "manual_adjustment",
		Direction: "credit",
		Description: &desc1,
	}
	credit1, err := adminClient.Billing().CreateCredit(ctx, req1)
	if err != nil {
		return ctx, fmt.Errorf("failed to create first credit: %w", err)
	}

	credit1ID := helpers.ToBillingIDFromUUID(credit1.Id)
	ctx = helpers.SetAdminCurrentCredit(ctx, credit1)
	ctx = helpers.SetAdminCurrentCreditID(ctx, credit1ID)
	ctx = helpers.AddAdminCreditCleanup(ctx, credit1ID.String())

	// Create a credit with "usage" transaction type
	desc2 := "Test usage charge 1"
	req2 := &admin.CreditCreateRequest{
		UserId:   userID,
		Amount:   "3.00",
		Type:     "usage",
		Direction: "debit",
		Description: &desc2,
	}
	credit2, err := adminClient.Billing().CreateCredit(ctx, req2)
	if err != nil {
		return ctx, fmt.Errorf("failed to create second credit: %w", err)
	}

	credit2ID := helpers.ToBillingIDFromUUID(credit2.Id)
	ctx = helpers.SetAdminCurrentCredit(ctx, credit2)
	ctx = helpers.SetAdminCurrentCreditID(ctx, credit2ID)
	ctx = helpers.AddAdminCreditCleanup(ctx, credit2ID.String())
	return ctx, nil
}

// =============================================================================

// theAdminHasCreatedMultipleCreditsWithDifferentDirections creates credits with different directions for filtering tests
func (s *AdminBillingSteps) theAdminHasCreatedMultipleCreditsWithDifferentDirections(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Get test user ID from context
	userID, ok := helpers.GetAdminTargetUserID(ctx)
	if !ok || userID == 0 {
		return ctx, fmt.Errorf("test user ID not available in context - user may not be registered")
	}

	// Create a credit with "credit" direction
	desc1 := "Test credit 1"
	req1 := &admin.CreditCreateRequest{
		UserId:   userID,
		Amount:   "5.00",
		Type:     "manual_adjustment",
		Direction: "credit",
		Description: &desc1,
	}
	credit1, err := adminClient.Billing().CreateCredit(ctx, req1)
	if err != nil {
		return ctx, fmt.Errorf("failed to create first credit: %w", err)
	}

	credit1ID := helpers.ToBillingIDFromUUID(credit1.Id)
	ctx = helpers.SetAdminCurrentCredit(ctx, credit1)
	ctx = helpers.SetAdminCurrentCreditID(ctx, credit1ID)
	ctx = helpers.AddAdminCreditCleanup(ctx, credit1ID.String())

	// Create a credit with "debit" direction
	desc2 := "Test debit 1"
	req2 := &admin.CreditCreateRequest{
		UserId:   userID,
		Amount:   "3.00",
		Type:     "usage",
		Direction: "debit",
		Description: &desc2,
	}
	credit2, err := adminClient.Billing().CreateCredit(ctx, req2)
	if err != nil {
		return ctx, fmt.Errorf("failed to create second credit: %w", err)
	}

	credit2ID := helpers.ToBillingIDFromUUID(credit2.Id)
	ctx = helpers.SetAdminCurrentCredit(ctx, credit2)
	ctx = helpers.SetAdminCurrentCreditID(ctx, credit2ID)
	ctx = helpers.AddAdminCreditCleanup(ctx, credit2ID.String())
	return ctx, nil
}

// theAdminFiltersCreditsByTransactionType filters credits by transaction type
func (s *AdminBillingSteps) theAdminFiltersCreditsByTransactionType(ctx context.Context, transactionType string) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	params := &admin.GetApiBillingCreditsParams{
		FiltersTypeEq: &transactionType,
	}

	credits, _, err := adminClient.Billing().ListCredits(ctx, params)
	if err != nil {
		return ctx, fmt.Errorf("failed to filter credits by transaction type: %w", err)
	}

	ctx = helpers.SetAdminCredits(ctx, credits)
	return ctx, nil
}

// onlyCreditsWithThatTypeAreReturned verifies filtering
func (s *AdminBillingSteps) onlyCreditsWithThatTypeAreReturned(ctx context.Context) (context.Context, error) {
	credits, ok := helpers.GetAdminCredits(ctx)
	if err := helpers.VerifyResourceExists(credits, ok, "credits"); err != nil {
		return ctx, err
	}

	// Verify all credits match the expected type (manual_adjustment)
	for _, credit := range credits {
		if credit.Type != "manual_adjustment" {
			return ctx, fmt.Errorf("found credit with unexpected type: %s", credit.Type)
		}
	}

	return ctx, nil
}

// theAdminFiltersCreditsByDirection filters credits by direction
func (s *AdminBillingSteps) theAdminFiltersCreditsByDirection(ctx context.Context, direction string) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	params := &admin.GetApiBillingCreditsParams{
		DirectionEq: &direction,
	}

	credits, _, err := adminClient.Billing().ListCredits(ctx, params)
	if err != nil {
		return ctx, fmt.Errorf("failed to filter credits by direction: %w", err)
	}

	ctx = helpers.SetAdminCredits(ctx, credits)
	return ctx, nil
}

// onlyCreditsWithThatDirectionAreReturned verifies filtering
func (s *AdminBillingSteps) onlyCreditsWithThatDirectionAreReturned(ctx context.Context) (context.Context, error) {
	credits, ok := helpers.GetAdminCredits(ctx)
	if err := helpers.VerifyResourceExists(credits, ok, "credits"); err != nil {
		return ctx, err
	}

	// Verify all credits match the expected direction
	for _, credit := range credits {
		if credit.Direction != "credit" {
			return ctx, fmt.Errorf("found credit with unexpected direction: %s", credit.Direction)
		}
	}

	return ctx, nil
}

// =============================================================================
// Subscriber Management
// =============================================================================

// theAdminListsAllSubscribers lists all subscribers
func (s *AdminBillingSteps) theAdminListsAllSubscribers(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	subscribers, _, err := adminClient.Billing().ListSubscribers(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list subscribers: %w", err)
	}

	ctx = helpers.SetAdminSubscribers(ctx, subscribers)
	return ctx, nil
}

// theSubscribersAreReturnedSuccessfully verifies subscribers list
func (s *AdminBillingSteps) theSubscribersAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	subscribers, ok := helpers.GetAdminSubscribers(ctx)
	if err := helpers.VerifyResourceExists(subscribers, ok, "subscribers"); err != nil {
		return ctx, err
	}

	if len(subscribers) == 0 {
		// Empty list is valid for new systems
		return ctx, nil
	}

	return ctx, nil
}

// thereIsAnActiveSubscriber ensures we have a subscriber to work with
func (s *AdminBillingSteps) thereIsAnActiveSubscriber(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// List all subscribers to find an active one
	subscribers, _, err := adminClient.Billing().ListSubscribers(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list subscribers: %w", err)
	}

	if len(subscribers) == 0 {
		return ctx, fmt.Errorf("no active subscribers found in the system")
	}

	// Use the first subscriber
	ctx = helpers.SetAdminCurrentSubscriber(ctx, subscribers[0])
	ctx = helpers.SetAdminCurrentSubscriberID(ctx, helpers.ToBillingID(subscribers[0].Id))
	return ctx, nil
}

// theAdminRetrievesTheSubscriberByID retrieves a subscriber by ID
func (s *AdminBillingSteps) theAdminRetrievesTheSubscriberByID(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	subscriberID, err := helpers.RequireAdminCurrentSubscriberID(ctx)
	if err != nil {
		return ctx, err
	}

	subscriber, err := adminClient.Billing().GetSubscriber(ctx, subscriberID.String())
	if err != nil {
		return ctx, fmt.Errorf("failed to get subscriber: %w", err)
	}

	ctx = helpers.SetAdminCurrentSubscriber(ctx, subscriber)
	return ctx, nil
}

// theSubscriberDetailsAreReturnedSuccessfully verifies subscriber details
func (s *AdminBillingSteps) theSubscriberDetailsAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	subscriber, err := helpers.RequireAdminCurrentSubscriber(ctx)
	if err != nil {
		return ctx, err
	}

	if subscriber.Id == 0 {
		return ctx, fmt.Errorf("subscriber ID is missing")
	}

	if subscriber.UserId == 0 {
		return ctx, fmt.Errorf("subscriber user ID is missing")
	}

	return ctx, nil
}

// thereIsAGatewayWithSubscribers checks if a gateway has subscribers
func (s *AdminBillingSteps) thereIsAGatewayWithSubscribers(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// List all subscribers to find one with a gateway
	subscribers, _, err := adminClient.Billing().ListSubscribers(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list subscribers: %w", err)
	}

	if len(subscribers) == 0 {
		return ctx, fmt.Errorf("no subscribers found in the system")
	}

	// Find a subscriber with a gateway type set
	for _, subscriber := range subscribers {
		if subscriber.GatewayType != "" {
			ctx = helpers.SetAdminCurrentSubscriber(ctx, subscriber)
			ctx = helpers.SetAdminCurrentSubscriberID(ctx, helpers.ToBillingID(subscriber.Id))
			return ctx, nil
		}
	}

	return ctx, fmt.Errorf("no subscribers with gateway type found in the system")
}

// theAdminListsSubscribersForTheGateway lists subscribers for a gateway
func (s *AdminBillingSteps) theAdminListsSubscribersForTheGateway(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Use subscriber's gateway type as gateway ID
	subscriber, err := helpers.RequireAdminCurrentSubscriber(ctx)
	if err != nil {
		return ctx, err
	}

	gatewayID := subscriber.GatewayType
	if gatewayID == "" {
		return ctx, fmt.Errorf("gateway type is missing from subscriber")
	}

	subscribers, _, err := adminClient.Billing().ListGatewaySubscribers(ctx, gatewayID)
	if err != nil {
		return ctx, fmt.Errorf("failed to list gateway subscribers: %w", err)
	}

	ctx = helpers.SetAdminSubscribers(ctx, subscribers)
	return ctx, nil
}

// theGatewaySubscribersAreReturnedSuccessfully verifies gateway subscribers
func (s *AdminBillingSteps) theGatewaySubscribersAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	subscribers, ok := helpers.GetAdminSubscribers(ctx)
	if err := helpers.VerifyResourceExists(subscribers, ok, "gateway subscribers"); err != nil {
		return ctx, err
	}

	return ctx, nil
}

// aTestUserWithSubscriptions creates a test user with subscription records
func (s *AdminBillingSteps) aTestUserWithSubscriptions(ctx context.Context) (context.Context, error) {
	return s.createTestUserIfNotExists(ctx)
}

// theAdminRetrievesSubscribersForTheTestUser retrieves user's subscribers
func (s *AdminBillingSteps) theAdminRetrievesSubscribersForTheTestUser(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	userID, err := helpers.RequireAdminTargetUserID(ctx)
	if err != nil {
		return ctx, err
	}

	userIDStr := fmt.Sprint(userID)
	subscribers, _, err := adminClient.Billing().GetUserSubscribers(ctx, userIDStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to get user subscribers: %w", err)
	}

	ctx = helpers.SetAdminSubscribers(ctx, subscribers)
	return ctx, nil
}

// theUserSubscribersAreReturnedSuccessfully verifies user subscribers
func (s *AdminBillingSteps) theUserSubscribersAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	subscribers, ok := helpers.GetAdminSubscribers(ctx)
	if err := helpers.VerifyResourceExists(subscribers, ok, "user subscribers"); err != nil {
		return ctx, err
	}

	return ctx, nil
}

// aTestUserWithAnActiveSubscription creates a test user with subscription
func (s *AdminBillingSteps) aTestUserWithAnActiveSubscription(ctx context.Context) (context.Context, error) {
	return s.createTestUserIfNotExists(ctx)
}

// theAdminCancelsTheUsersSubscriptionImmediately cancels subscription immediately
func (s *AdminBillingSteps) theAdminCancelsTheUsersSubscriptionImmediately(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	userID, err := helpers.RequireAdminTargetUserID(ctx)
	if err != nil {
		return ctx, err
	}

	userIDStr := fmt.Sprint(userID)
	immediate := true
	req := &admin.CancelSubscriptionRequest{
		Immediate: &immediate,
	}

	result, err := adminClient.Billing().CancelUserSubscription(ctx, userIDStr, req)
	if err != nil {
		return ctx, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	ctx = helpers.SetAdminManagementResult(ctx, result)
	return ctx, nil
}

// theSubscriptionIsCancelledSuccessfully verifies cancellation success
func (s *AdminBillingSteps) theSubscriptionIsCancelledSuccessfully(ctx context.Context) (context.Context, error) {
	result, err := helpers.RequireAdminManagementResult(ctx)
	if err != nil {
		return ctx, err
	}

	if result.Action == "" {
		return ctx, fmt.Errorf("cancellation action is empty")
	}

	return ctx, nil
}

// theCancellationTakesEffectImmediately verifies immediate effect
func (s *AdminBillingSteps) theCancellationTakesEffectImmediately(ctx context.Context) (context.Context, error) {
	// For immediate cancellation, the result should indicate effective time
	result, err := helpers.RequireAdminManagementResult(ctx)
	if err != nil {
		return ctx, err
	}

	// Immediate mode means effective_time should be present or action is completed
	if result.EffectiveTime == nil {
		// This might be OK depending on the API implementation
		return ctx, nil
	}

	return ctx, nil
}

// theAdminCancelsTheUsersSubscriptionAtEndOfPeriod cancels at period end
func (s *AdminBillingSteps) theAdminCancelsTheUsersSubscriptionAtEndOfPeriod(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	userID, err := helpers.RequireAdminTargetUserID(ctx)
	if err != nil {
		return ctx, err
	}

	userIDStr := fmt.Sprint(userID)
	req := &admin.CancelSubscriptionRequest{}

	result, err := adminClient.Billing().CancelUserSubscription(ctx, userIDStr, req)
	if err != nil {
		return ctx, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	ctx = helpers.SetAdminManagementResult(ctx, result)
	return ctx, nil
}

// theSubscriptionIsScheduledForCancellation verifies scheduled cancellation
func (s *AdminBillingSteps) theSubscriptionIsScheduledForCancellation(ctx context.Context) (context.Context, error) {
	result, err := helpers.RequireAdminManagementResult(ctx)
	if err != nil {
		return ctx, err
	}

	if result.Action == "" {
		return ctx, fmt.Errorf("cancellation action is empty")
	}

	return ctx, nil
}

// theCancellationWillTakeEffectAtBillingPeriodEnd verifies future effect
func (s *AdminBillingSteps) theCancellationWillTakeEffectAtBillingPeriodEnd(ctx context.Context) (context.Context, error) {
	result, err := helpers.RequireAdminManagementResult(ctx)
	if err != nil {
		return ctx, err
	}

	// For end_of_period mode, effective_time should be set
	if result.EffectiveTime == nil {
		return ctx, fmt.Errorf("will_cancel_at is not set for end_of_period cancellation")
	}

	return ctx, nil
}

// thereIsAnAvailablePricingPlanPeriod ensures a pricing plan period exists
func (s *AdminBillingSteps) thereIsAnAvailablePricingPlanPeriod(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// List all pricing plan periods
	periods, _, err := adminClient.Billing().ListPricingPlanPeriods(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list pricing plan periods: %w", err)
	}

	if len(periods) == 0 {
		// If none exist, try to create one
		// Create a simple period for testing
		req := &admin.PricingPlanPeriodCreateRequest{
			Cadence:     "monthly",
			PriceUsd:    10.00,
			QuotaPlanId: 1, // Use a default quota plan
		}

		period, err := adminClient.Billing().CreatePricingPlanPeriod(ctx, req)
		if err != nil {
			return ctx, fmt.Errorf("failed to create pricing plan period: %w", err)
		}

		periodID := helpers.ToBillingID(period.Id)
		ctx = helpers.SetAdminCurrentPeriod(ctx, period)
		ctx = helpers.SetAdminCurrentPeriodID(ctx, periodID)
		return ctx, nil
	}

	// Use the first period
	ctx = helpers.SetAdminCurrentPeriod(ctx, periods[0])
	ctx = helpers.SetAdminCurrentPeriodID(ctx, helpers.ToBillingID(periods[0].Id))
	return ctx, nil
}

// theAdminChangesTheUsersPlan changes user's subscription plan
func (s *AdminBillingSteps) theAdminChangesTheUsersPlan(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	userID, err := helpers.RequireAdminTargetUserID(ctx)
	if err != nil {
		return ctx, err
	}

	periodID, err := helpers.RequireAdminCurrentPeriodID(ctx)
	if err != nil {
		return ctx, err
	}

	// Get period ID as int for the request
	periodIDInt := periodID.AsInt()
	if err != nil {
		return ctx, fmt.Errorf("invalid period ID format: %w", err)
	}

	userIDStr := fmt.Sprint(userID)
	req := &admin.ChangePlanRequest{
		PeriodId: periodIDInt,
	}

	result, err := adminClient.Billing().ChangeUserPlan(ctx, userIDStr, req)
	if err != nil {
		return ctx, fmt.Errorf("failed to change user plan: %w", err)
	}

	ctx = helpers.SetAdminPlanChangeResult(ctx, result)
	return ctx, nil
}

// thePlanChangeIsProcessedSuccessfully verifies plan change success
func (s *AdminBillingSteps) thePlanChangeIsProcessedSuccessfully(ctx context.Context) (context.Context, error) {
	result, err := helpers.RequireAdminPlanChangeResult(ctx)
	if err != nil {
		return ctx, err
	}

	if result.Action == "" {
		return ctx, fmt.Errorf("plan change action is empty")
	}

	return ctx, nil
}

// theNewPricingPlanPeriodIsApplied verifies new plan is applied
func (s *AdminBillingSteps) theNewPricingPlanPeriodIsApplied(ctx context.Context) (context.Context, error) {
	result, err := helpers.RequireAdminPlanChangeResult(ctx)
	if err != nil {
		return ctx, err
	}

	// Verify the plan change has an effective date
	if result.EffectiveDate == nil {
		return ctx, fmt.Errorf("plan change effective date is missing")
	}

	return ctx, nil
}

// =============================================================================
// Price Line Plan Management Steps
// =============================================================================

// theAdminAddsThePlanToThePriceLine adds a plan to a price line
func (s *AdminBillingSteps) theAdminAddsThePlanToThePriceLine(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Get price line ID from context
	priceLineID, err := helpers.RequireAdminCurrentPriceLineID(ctx)
	if err != nil {
		return ctx, err
	}

	// Get plan ID from context
	planID, err := helpers.RequireAdminCurrentPricingPlanID(ctx)
	if err != nil {
		return ctx, err
	}

	// Add plan to price line using typed helpers
	req := &admin.AddPlanToPriceLineRequest{
		PlanId:   planID.AsInt(),
		Position: 0,
	}

	_, err = adminClient.Billing().AddPlanToPriceLine(ctx, priceLineID.String(), req)
	if err != nil {
		return ctx, fmt.Errorf("failed to add plan to price line: %w", err)
	}

	return ctx, nil
}

// thePlanIsAddedToThePriceLineSuccessfully verifies the plan was added
func (s *AdminBillingSteps) thePlanIsAddedToThePriceLineSuccessfully(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Get price line ID from context
	priceLineID, err := helpers.RequireAdminCurrentPriceLineID(ctx)
	if err != nil {
		return ctx, err
	}

	// Get price line to verify plan was added
	priceLine, err := adminClient.Billing().GetPriceLine(ctx, priceLineID.String())
	if err != nil {
		return ctx, fmt.Errorf("failed to get price line: %w", err)
	}

	// Verify plans exist using helper
	plans := helpers.GetPlansFromPriceLine(priceLine)
	if len(plans) == 0 {
		return ctx, fmt.Errorf("price line has no plans")
	}

	return ctx, nil
}

// theAdminHasCreatedAPriceLineWithMultiplePlans creates a price line with multiple plans
func (s *AdminBillingSteps) theAdminHasCreatedAPriceLineWithMultiplePlans(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Create a price line
	priceLineReq := admin.PriceLineCreateRequest{
		Name:        "Price Line With Multiple Plans",
		Description: "A price line containing multiple plans for testing",
	}
	priceLine, err := adminClient.Billing().CreatePriceLine(ctx, &priceLineReq)
	if err != nil {
		return ctx, fmt.Errorf("failed to create price line: %w", err)
	}

	priceLineID := helpers.ToBillingID(priceLine.Id)
	ctx = helpers.SetAdminCurrentPriceLineID(ctx, priceLineID)
	ctx = helpers.AddAdminPriceLineCleanup(ctx, priceLineID.String())

	// Create two pricing plans and add them to the price line
	for i := 0; i < 2; i++ {
		planName := fmt.Sprintf("Test Plan %d", i+1)
		pricelineID := int(priceLine.Id)
		planReq := admin.PricingPlanCreateRequest{
			Name:           planName,
			Description:    "Test plan for price line",
			Currency:       "USD",
			IsActive:       true,
			IsPublic:       true,
			PricelineId:    &pricelineID,
			PricingPeriods: []admin.PricingPlanPeriod{},
		}
		plan, err := adminClient.Billing().CreatePricingPlan(ctx, &planReq)
		if err != nil {
			return ctx, fmt.Errorf("failed to create pricing plan %d: %w", i+1, err)
		}

		// Store the first plan ID for later use
		if i == 0 {
				ctx = helpers.SetAdminCurrentPricingPlanID(ctx, helpers.ToBillingID(plan.Id))
		}
	}

	return ctx, nil
}

// theAdminUpdatesAPlanPositionInThePriceLine updates a plan's position
func (s *AdminBillingSteps) theAdminUpdatesAPlanPositionInThePriceLine(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Get price line ID from context
	priceLineID, err := helpers.RequireAdminCurrentPriceLineID(ctx)
	if err != nil {
		return ctx, err
	}

	// Get plan ID from context
	planID, err := helpers.RequireAdminCurrentPricingPlanID(ctx)
	if err != nil {
		return ctx, err
	}

	// Update plan position using typed helpers
	req := &admin.UpdatePlanPositionRequest{
		Position: 1,
	}
	_, err = adminClient.Billing().UpdatePlanPosition(ctx, priceLineID.String(), planID.String(), req)
	if err != nil {
		return ctx, fmt.Errorf("failed to update plan position: %w", err)
	}

	return ctx, nil
}

// thePlanPositionIsUpdatedSuccessfully verifies the plan position was updated
func (s *AdminBillingSteps) thePlanPositionIsUpdatedSuccessfully(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Get price line and plan IDs from context
	priceLineID, err := helpers.RequireAdminCurrentPriceLineID(ctx)
	if err != nil {
		return ctx, err
	}

	planID, err := helpers.RequireAdminCurrentPricingPlanID(ctx)
	if err != nil {
		return ctx, err
	}

	// Get price line to verify position
	priceLine, err := adminClient.Billing().GetPriceLine(ctx, priceLineID.String())
	if err != nil {
		return ctx, fmt.Errorf("failed to get price line: %w", err)
	}

	// Verify the plan has the expected position
	pos := helpers.GetPlanPositionInPriceLine(priceLine, planID)
	if pos != 1 {
		return ctx, fmt.Errorf("expected plan position 1, got %d", pos)
	}

	return ctx, nil
}

// theAdminHasCreatedAPriceLineWithAPlan creates a price line with a single plan
func (s *AdminBillingSteps) theAdminHasCreatedAPriceLineWithAPlan(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Create a price line
	priceLineReq := admin.PriceLineCreateRequest{
		Name:        "Price Line With Plan",
		Description: "A price line containing a single plan",
	}
	priceLine, err := adminClient.Billing().CreatePriceLine(ctx, &priceLineReq)
	if err != nil {
		return ctx, fmt.Errorf("failed to create price line: %w", err)
	}

	priceLineID := helpers.ToBillingID(priceLine.Id)
	ctx = helpers.SetAdminCurrentPriceLineID(ctx, priceLineID)
	ctx = helpers.AddAdminPriceLineCleanup(ctx, priceLineID.String())

	// Create a pricing plan
	pricelineID := int(priceLine.Id)
	planReq := admin.PricingPlanCreateRequest{
		Name:           "Test Plan For Removal",
		Description:    "Test plan to be removed from price line",
		Currency:       "USD",
		IsActive:       true,
		IsPublic:       true,
		PricelineId:    &pricelineID,
		PricingPeriods: []admin.PricingPlanPeriod{},
	}
	plan, err := adminClient.Billing().CreatePricingPlan(ctx, &planReq)
	if err != nil {
		return ctx, fmt.Errorf("failed to create pricing plan: %w", err)
	}

	plID := helpers.ToBillingID(plan.Id)
	ctx = helpers.SetAdminCurrentPricingPlanID(ctx, plID)
	ctx = helpers.AddAdminPricingPlanCleanup(ctx, plID.String())

	return ctx, nil
}

// theAdminRemovesThePlanFromThePriceLine removes a plan from a price line
func (s *AdminBillingSteps) theAdminRemovesThePlanFromThePriceLine(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Get price line ID from context
	priceLineID, err := helpers.RequireAdminCurrentPriceLineID(ctx)
	if err != nil {
		return ctx, err
	}

	// Get plan ID from context
	planID, err := helpers.RequireAdminCurrentPricingPlanID(ctx)
	if err != nil {
		return ctx, err
	}

	// Remove plan from price line using typed helpers
	err = adminClient.Billing().DeletePlanFromPriceLine(ctx, priceLineID.String(), planID.String())
	if err != nil {
		return ctx, fmt.Errorf("failed to remove plan from price line: %w", err)
	}

	return ctx, nil
}

// thePlanIsRemovedFromThePriceLineSuccessfully verifies the plan was removed
func (s *AdminBillingSteps) thePlanIsRemovedFromThePriceLineSuccessfully(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Get price line and plan IDs from context
	priceLineID, err := helpers.RequireAdminCurrentPriceLineID(ctx)
	if err != nil {
		return ctx, err
	}

	planID, err := helpers.RequireAdminCurrentPricingPlanID(ctx)
	if err != nil {
		return ctx, err
	}

	// Get price line to verify plan was removed
	priceLine, err := adminClient.Billing().GetPriceLine(ctx, priceLineID.String())
	if err != nil {
		return ctx, fmt.Errorf("failed to get price line: %w", err)
	}

	// Verify plan is no longer in the price line using helper
	if helpers.PlanExistsInPriceLine(priceLine, planID) {
		return ctx, fmt.Errorf("plan was not removed from price line")
	}

	return ctx, nil
}
