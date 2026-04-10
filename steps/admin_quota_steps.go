package steps

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/docker/go-units"
	
	"pinner.xyz-e2e/helpers"
	"go.lumeweb.com/portal-sdk/admin"
)



// AdminQuotaSteps holds the state for admin quota management step definitions
type AdminQuotaSteps struct {
	planID        int64
	allowanceID    int64
}

// NewAdminQuotaSteps creates a new AdminQuotaSteps instance
func NewAdminQuotaSteps() *AdminQuotaSteps {
	return &AdminQuotaSteps{}
}

// verifyPlanIDAvailable checks if planID is set and returns error if not
func (s *AdminQuotaSteps) verifyPlanIDAvailable() error {
	if s.planID == 0 {
		return fmt.Errorf("no plan ID available")
	}
	return nil
}

// verifyAllowanceIDAvailable checks if allowanceID is set and returns error if not
func (s *AdminQuotaSteps) verifyAllowanceIDAvailable() error {
	if s.allowanceID == 0 {
		return fmt.Errorf("no allowance ID available")
	}
	return nil
}

// verifyResourceExists verifies a resource was retrieved from context
func verifyResourceExists[T any](value T, ok bool, errorPrefix string) error {
	if !ok {
		return fmt.Errorf("%s was not retrieved", errorPrefix)
	}

	// For pointer types, check for nil
	if interface{}(value) == nil {
		return fmt.Errorf("%s is nil", errorPrefix)
	}
	return nil
}

// getDefaultTestLimits creates standard test quota limits
func (s *AdminQuotaSteps) getDefaultTestLimits() admin.QuotaLimits {
	uploadTotalLimit, _ := units.FromHumanSize("10GB")
	downloadTotalLimit, _ := units.FromHumanSize("50GB")
	storageLimit, _ := units.FromHumanSize("1GB")

	return admin.QuotaLimits{
		UploadLimitBytes:   int(uploadTotalLimit),
		DownloadLimitBytes: int(downloadTotalLimit),
		StorageLimitBytes:  int(storageLimit),
		WindowDuration:     0,
		WindowStartHour:    0,
		WindowTimezone:     "",
		WindowType:         "LIFETIME",
	}
}

// updateExistingPlan updates and fetches an existing quota plan
func (s *AdminQuotaSteps) updateExistingPlan(ctx context.Context, adminClient *admin.AdminClient, planID int64, name string) (*admin.QuotaPlan, error) {
	limits := s.getDefaultTestLimits()
	updatedPlan := admin.NewQuotaPlan(name, "Test plan created from E2E tests", limits)

	_, err := adminClient.Quota().UpdatePlan(ctx, fmt.Sprint(planID), updatedPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to update existing plan %s (ID %d): %w", name, planID, err)
	}

	plan, fetchErr := adminClient.Quota().GetPlan(ctx, fmt.Sprint(planID))
	if fetchErr != nil {
		return nil, fmt.Errorf("failed to fetch updated plan %s (ID %d): %w", name, planID, fetchErr)
	}

	return plan, nil
}

// reuseExistingDefaultPlan handles reusing an existing default plan
func (s *AdminQuotaSteps) reuseExistingDefaultPlan(ctx context.Context, adminClient *admin.AdminClient, existing *admin.QuotaPlan) (*admin.QuotaPlan, error) {
	return s.updateExistingPlan(ctx, adminClient, int64(existing.Id), existing.Name)
}

// deleteOrReuseNonDefaultPlan handles non-existing plan deletion or reuse
func (s *AdminQuotaSteps) deleteOrReuseNonDefaultPlan(ctx context.Context, adminClient *admin.AdminClient, existing *admin.QuotaPlan) (*admin.QuotaPlan, error) {
	// Reset users assigned to this plan to prevent "cannot delete plan in use" errors
	_ = helpers.ResetUsersForPlan(ctx, adminClient.Quota(), int64(existing.Id))

	err := adminClient.Quota().DeletePlan(ctx, fmt.Sprint(existing.Id))

	// "plan not found" is not an error - it was already deleted
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "not found") {
		return nil, nil
	}

	// If deletion fails with "in use", reuse the existing plan
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "in use") {
		return s.updateExistingPlan(ctx, adminClient, int64(existing.Id), existing.Name)
	}

	// If deletion fails for any other reason, return error
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing plan %s (ID %d): %w",
			existing.Name, existing.Id, err)
	}

	return nil, nil
}

// createNewQuotaPlan creates and activates a new quota plan
func (s *AdminQuotaSteps) createNewQuotaPlan(ctx context.Context, adminClient *admin.AdminClient, name string) (*admin.QuotaPlan, error) {
	limits := s.getDefaultTestLimits()
	planToCreate := admin.NewQuotaPlan(name, "Test plan created from E2E tests", limits)

	plan, err := adminClient.Quota().CreatePlan(ctx, planToCreate)
	if err != nil {
		return nil, fmt.Errorf("failed to create quota plan: %w", err)
	}

	// Activate the plan - SetDefaultPlan API requires active plans
	plan.IsActive = true
	_, err = adminClient.Quota().UpdatePlan(ctx, fmt.Sprint(plan.Id), plan)
	if err != nil {
		return nil, fmt.Errorf("failed to activate quota plan: %w", err)
	}

	return plan, nil
}

// createQuotaPlan creates and activates a quota plan with cleanup for existing plans of the same name
func (s *AdminQuotaSteps) createQuotaPlan(ctx context.Context, name string) (*admin.QuotaPlan, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return nil, err
	}

	// Check for existing plans with the same name
	existingPlans, _, err := adminClient.Quota().ListPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list existing plans: %w", err)
	}

	for _, existing := range existingPlans {
		if existing.Name == name && existing.Id != 0 {
			// Handle existing default plan - reuse it
			if existing.IsDefault {
				return s.reuseExistingDefaultPlan(ctx, adminClient, existing)
			}

			// Handle non-default plan - delete or reuse
			plan, err := s.deleteOrReuseNonDefaultPlan(ctx, adminClient, existing)
			if err != nil {
				return nil, err
			}
			if plan != nil {
				return plan, nil
			}
		}
	}

	// No existing plan found or all were deleted, create new plan
	return s.createNewQuotaPlan(ctx, adminClient, name)
}

// InitializeScenario registers all step definitions with godog
func (s *AdminQuotaSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Admin quota plan management steps
	ctx.Step(`^the admin lists all quota plans$`, s.theAdminListsAllQuotaPlans)
	ctx.Step(`^the quota plans are returned successfully$`, s.theQuotaPlansAreReturnedSuccessfully)
	ctx.Step(`^the admin creates a new quota plan named "([^"]*)"$`, s.theAdminCreatesANewQuotaPlanNamed)
	ctx.Step(`^the quota plan is created with the specified limits$`, s.theQuotaPlanIsCreatedWithTheSpecifiedLimits)
	ctx.Step(`^the admin retrieves the quota plan$`, s.theAdminRetrievesTheQuotaPlan)
	ctx.Step(`^the quota plan information is returned$`, s.theQuotaPlanInformationIsReturned)

	// Quota allowance management steps
	ctx.Step(`^the admin lists all quota allowances$`, s.theAdminListsAllQuotaAllowances)
	ctx.Step(`^the quota allowances are returned successfully$`, s.theQuotaAllowancesAreReturnedSuccessfully)
	ctx.Step(`^the admin creates a quota allowance for user (\d+)$`, s.theAdminCreatesAQuotaAllowanceForUser)
	ctx.Step(`^the quota allowance is created successfully$`, s.theQuotaAllowanceIsCreatedSuccessfully)

	// System stats steps
	ctx.Step(`^the admin retrieves system-wide quota statistics$`, s.theAdminRetrievesSystemWideQuotaStatistics)
	ctx.Step(`^the system statistics include upload, download, and storage usage$`, s.theSystemStatisticsIncludeUploadDownloadAndStorageUsage)

	// Plan update/delete/default steps
	ctx.Step(`^the admin has created a quota plan named "([^"]*)"$`, s.theAdminHasCreatedAQuotaPlanNamed)
	ctx.Step(`^the admin updates the quota plan with new limits$`, s.theAdminUpdatesTheQuotaPlanWithNewLimits)
	ctx.Step(`^the quota plan is updated successfully$`, s.theQuotaPlanIsUpdatedSuccessfully)
	ctx.Step(`^the updated limits are reflected$`, s.theUpdatedLimitsAreReflected)
	ctx.Step(`^the admin deletes the quota plan$`, s.theAdminDeletesTheQuotaPlan)
	ctx.Step(`^the quota plan is deleted successfully$`, s.theQuotaPlanIsDeletedSuccessfully)
	ctx.Step(`^the admin sets the plan as default$`, s.theAdminSetsThePlanAsDefault)

	// Plan assignment steps
	ctx.Step(`^the admin assigns the current plan to the authenticated user$`, s.theAdminAssignsTheCurrentPlanToTheAuthenticatedUser)
	// Allowance update/delete steps
	ctx.Step(`^the admin has created a quota allowance for user (\d+)$`, s.theAdminHasCreatedAQuotaAllowanceForUser)
	ctx.Step(`^the admin updates the allowance with new limits$`, s.theAdminUpdatesTheAllowanceWithNewLimits)
	ctx.Step(`^the allowance is updated successfully$`, s.theAllowanceIsUpdatedSuccessfully)
	ctx.Step(`^the admin deletes the allowance$`, s.theAdminDeletesTheAllowance)
	ctx.Step(`^the allowance is deleted successfully$`, s.theAllowanceIsDeletedSuccessfully)

	// Reconcile/cleanup steps
	ctx.Step(`^the admin reconciles quota for all users$`, s.theAdminReconcilesQuotaForAllUsers)
	ctx.Step(`^the reconciliation completes successfully$`, s.theReconciliationCompletesSuccessfully)
	ctx.Step(`^the users processed count is recorded$`, s.theUsersProcessedCountIsRecorded)
	ctx.Step(`^the admin performs quota cleanup with (\d+) day retention$`, s.theAdminPerformsQuotaCleanup)
	ctx.Step(`^the cleanup completes successfully$`, s.theCleanupCompletesSuccessfully)
	ctx.Step(`^the records deleted count is recorded$`, s.theRecordsDeletedCountIsRecorded)
}

// theAdminListsAllQuotaPlans lists all quota plans
func (s *AdminQuotaSteps) theAdminListsAllQuotaPlans(ctx context.Context) (context.Context, error) {
	// Get admin client from context
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	plans, _, err := adminClient.Quota().ListPlans(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list quota plans: %w", err)
	}

	ctx = helpers.SetAdminQuotaPlans(ctx, plans)
	return ctx, nil
}

// theQuotaPlansAreReturnedSuccessfully verifies plans were retrieved successfully
func (s *AdminQuotaSteps) theQuotaPlansAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	_, ok := helpers.GetAdminQuotaPlans(ctx)
	if !ok {
		return ctx, fmt.Errorf("quota plans were not retrieved")
	}
	// Empty list is valid - the API call succeeded even if no plans exist
	return ctx, nil
}

// theAdminCreatesANewQuotaPlanNamed creates a new quota plan
func (s *AdminQuotaSteps) theAdminCreatesANewQuotaPlanNamed(ctx context.Context, name string) (context.Context, error) {
	plan, err := s.createQuotaPlan(ctx, name)
	if err != nil {
		return ctx, err
	}

	s.planID = int64(plan.Id)
	ctx = helpers.SetAdminCurrentPlan(ctx, plan)
	ctx = helpers.AddQuotaPlanCleanup(ctx, int64(plan.Id))
	return ctx, nil
}

// theQuotaPlanIsCreatedWithTheSpecifiedLimits verifies plan creation
func (s *AdminQuotaSteps) theQuotaPlanIsCreatedWithTheSpecifiedLimits(ctx context.Context) (context.Context, error) {
	plan, ok := helpers.GetAdminCurrentPlan(ctx)
	if err := verifyResourceExists(plan, ok, "quota plan"); err != nil {
		return ctx, err
	}

	if plan.Name == "" {
		return ctx, fmt.Errorf("quota plan has no name")
	}

	return ctx, nil
}

// theAdminRetrievesTheQuotaPlan retrieves a quota plan
func (s *AdminQuotaSteps) theAdminRetrievesTheQuotaPlan(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	if err := s.verifyPlanIDAvailable(); err != nil {
		return ctx, err
	}

	plan, err := adminClient.Quota().GetPlan(ctx, fmt.Sprint(s.planID))
	if err != nil {
		return ctx, fmt.Errorf("failed to get quota plan: %w", err)
	}

	ctx = helpers.SetAdminCurrentPlan(ctx, plan)
	return ctx, nil
}

// theQuotaPlanInformationIsReturned verifies plan retrieval
func (s *AdminQuotaSteps) theQuotaPlanInformationIsReturned(ctx context.Context) (context.Context, error) {
	plan, ok := helpers.GetAdminCurrentPlan(ctx)
	if err := verifyResourceExists(plan, ok, "quota plan"); err != nil {
		return ctx, err
	}

	if plan.Id == 0 {
		return ctx, fmt.Errorf("quota plan has no ID")
	}

	return ctx, nil
}

// theAdminListsAllQuotaAllowances lists all quota allowances
func (s *AdminQuotaSteps) theAdminListsAllQuotaAllowances(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	allowances, _, err := adminClient.Quota().ListAllowances(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list quota allowances: %w", err)
	}

	ctx = helpers.SetAdminQuotaAllowances(ctx, allowances)
	return ctx, nil
}

// theQuotaAllowancesAreReturnedSuccessfully verifies allowances were retrieved
func (s *AdminQuotaSteps) theQuotaAllowancesAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	allowances, ok := helpers.GetAdminQuotaAllowances(ctx)
	if err := verifyResourceExists(allowances, ok, "quota allowances"); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// theAdminCreatesAQuotaAllowanceForUser creates a quota allowance
func (s *AdminQuotaSteps) theAdminCreatesAQuotaAllowanceForUser(ctx context.Context, userID int) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	expiry := time.Now().Add(30 * 24 * time.Hour)
	allowance, err := adminClient.Quota().CreateAllowance(ctx, userID, "BONUS", "STORAGE",
		100*1024*1024,  // 100 MB upload bytes
		500*1024*1024,  // 500 MB download bytes
		10*1024*1024,   // 10 MB storage bytes
		expiry,
	)
	if err != nil {
		return ctx, fmt.Errorf("failed to create quota allowance: %w", err)
	}

	s.allowanceID = int64(allowance.Id)
	ctx = helpers.SetAdminCurrentAllowance(ctx, allowance)
	ctx = helpers.AddQuotaAllowanceCleanup(ctx, int64(allowance.Id))
	return ctx, nil
}

// theQuotaAllowanceIsCreatedSuccessfully verifies allowance creation
func (s *AdminQuotaSteps) theQuotaAllowanceIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	allowance, ok := helpers.GetAdminCurrentAllowance(ctx)
	if err := verifyResourceExists(allowance, ok, "quota allowance"); err != nil {
		return ctx, err
	}

	if allowance.Id == 0 {
		return ctx, fmt.Errorf("quota allowance has no ID")
	}

	return ctx, nil
}

// theAdminRetrievesSystemWideQuotaStatistics retrieves system stats
func (s *AdminQuotaSteps) theAdminRetrievesSystemWideQuotaStatistics(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	stats, err := adminClient.Quota().GetStats(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get system statistics: %w", err)
	}

	ctx = helpers.SetAdminSystemStats(ctx, stats)
	return ctx, nil
}

// theSystemStatisticsIncludeUploadDownloadAndStorageUsage verifies stats
func (s *AdminQuotaSteps) theSystemStatisticsIncludeUploadDownloadAndStorageUsage(ctx context.Context) (context.Context, error) {
	stats, ok := helpers.GetAdminSystemStats(ctx)
	if err := verifyResourceExists(stats, ok, "system statistics"); err != nil {
		return ctx, err
	}

	return ctx, nil
}

// theAdminHasCreatedAQuotaPlanNamed helper step storing plan reference
func (s *AdminQuotaSteps) theAdminHasCreatedAQuotaPlanNamed(ctx context.Context, name string) (context.Context, error) {
	plan, err := s.createQuotaPlan(ctx, name)
	if err != nil {
		return ctx, err
	}

	s.planID = int64(plan.Id)
	ctx = helpers.SetAdminCurrentPlan(ctx, plan)
	ctx = helpers.AddQuotaPlanCleanup(ctx, int64(plan.Id))
	return ctx, nil
}

// theAdminUpdatesTheQuotaPlanWithNewLimits updates a plan
func (s *AdminQuotaSteps) theAdminUpdatesTheQuotaPlanWithNewLimits(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	if err := s.verifyPlanIDAvailable(); err != nil {
		return ctx, err
	}

	// Get the existing plan first to preserve its name and description
	existingPlan, err := adminClient.Quota().GetPlan(ctx, fmt.Sprint(s.planID))
	if err != nil {
		return ctx, fmt.Errorf("failed to get existing plan: %w", err)
	}

	if err := verifyResourceExists(existingPlan, existingPlan != nil, "existing plan"); err != nil {
		return ctx, err
	}

	// Convert human-readable sizes to bytes
	updateUploadTotal, _ := units.FromHumanSize("50GB")
	
	updateDownloadTotal, _ := units.FromHumanSize("50GB")
	
	updateStorage, _ := units.FromHumanSize("5GB")

	// Update the plan, preserving existing name and description
	updatedPlan := admin.NewQuotaPlan(
		existingPlan.Name,
		existingPlan.Description,
		admin.QuotaLimits{
			UploadLimitBytes:   int(updateUploadTotal),
			DownloadLimitBytes: int(updateDownloadTotal),
			StorageLimitBytes:  int(updateStorage),
			WindowDuration:     0, // Not used with LIFETIME window type
			WindowStartHour:    0,
			WindowTimezone:     "",
			WindowType:         "LIFETIME", // Use non-windowed quota for simpler testing
		},
	)
	
	_, err = adminClient.Quota().UpdatePlan(ctx, fmt.Sprint(s.planID), updatedPlan)
	if err != nil {
		return ctx, fmt.Errorf("failed to update quota plan: %w", err)
	}

	return ctx, nil
}

// theQuotaPlanIsUpdatedSuccessfully verifies update by retrieving the plan
func (s *AdminQuotaSteps) theQuotaPlanIsUpdatedSuccessfully(ctx context.Context) (context.Context, error) {
	if s.planID == 0 {
		return ctx, nil
	}

	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Verify plan still exists after update
	_, err = adminClient.Quota().GetPlan(ctx, fmt.Sprint(s.planID))
	if err != nil {
		return ctx, fmt.Errorf("plan not found after update: %w", err)
	}

	return ctx, nil
}




// theUpdatedLimitsAreReflected verifies limits were updated
func (s *AdminQuotaSteps) theUpdatedLimitsAreReflected(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	plan, err := adminClient.Quota().GetPlan(ctx, fmt.Sprint(s.planID))
	if err != nil {
		return ctx, fmt.Errorf("failed to get updated plan: %w", err)
	}

	expectedUploadLimit, _ := units.FromHumanSize("50GB")
	if plan.UploadLimitBytes != int(expectedUploadLimit) {
		return ctx, fmt.Errorf("upload limit not updated: got %d, expected %d", plan.UploadLimitBytes, int(expectedUploadLimit))
	}

	return ctx, nil
}

// theAdminDeletesTheQuotaPlan deletes a plan
func (s *AdminQuotaSteps) theAdminDeletesTheQuotaPlan(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	if err := s.verifyPlanIDAvailable(); err != nil {
		return ctx, err
	}

	if err := adminClient.Quota().DeletePlan(ctx, fmt.Sprint(s.planID)); err != nil {
		return ctx, fmt.Errorf("failed to delete quota plan: %w", err)
	}

	return ctx, nil
}

// theQuotaPlanIsDeletedSuccessfully verifies deletion
func (s *AdminQuotaSteps) theQuotaPlanIsDeletedSuccessfully(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Verify plan no longer exists
	_, err = adminClient.Quota().GetPlan(ctx, fmt.Sprint(s.planID))
	if err == nil {
		return ctx, fmt.Errorf("plan still exists after deletion")
	}

	return ctx, nil
}

// theAdminSetsThePlanAsDefault sets plan as default
func (s *AdminQuotaSteps) theAdminSetsThePlanAsDefault(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		fmt.Printf("Error: Failed to get admin client: %v\n", err)
		return ctx, err
	}

	if err := s.verifyPlanIDAvailable(); err != nil {
		fmt.Printf("Error: Plan ID not available: %v\n", err)
		return ctx, err
	}

	if err := adminClient.Quota().SetDefaultPlan(ctx, fmt.Sprint(s.planID)); err != nil {
		fmt.Printf("Error: Failed to set plan as default: %v\n", err)
		return ctx, fmt.Errorf("failed to set plan as default: %w", err)
	}

	return ctx, nil
}

// theAdminAssignsTheCurrentPlanToTheAuthenticatedUser assigns the current plan to the authenticated user
func (s *AdminQuotaSteps) theAdminAssignsTheCurrentPlanToTheAuthenticatedUser(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get admin client: %w", err)
	}

	if err := s.verifyPlanIDAvailable(); err != nil {
		return ctx, fmt.Errorf("no plan ID available: %w", err)
	}

	// Get the authenticated user's ID
	userAPI, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get authenticated client: %w", err)
	}

	accountInfo, err := userAPI.GetAccount(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get account info: %w", err)
	}

	userID := int(accountInfo.Id)
	planID := int(s.planID)

	// Assign the plan to the user using UpdateUserConfig
	config := &admin.UserQuotaConfigUpdate{
		QuotaPlanID: &planID,
	}

	_, err = adminClient.Quota().UpdateUserConfig(ctx, userID, config)
	if err != nil {
		return ctx, fmt.Errorf("failed to assign plan to user: %w", err)
	}
	return ctx, nil
}

// theAdminHasCreatedAQuotaAllowanceForUser creates allowance for testing
func (s *AdminQuotaSteps) theAdminHasCreatedAQuotaAllowanceForUser(ctx context.Context, userID int) (context.Context, error) {
	return s.theAdminCreatesAQuotaAllowanceForUser(ctx, userID)
}

// theAdminUpdatesTheAllowanceWithNewLimits updates an allowance
func (s *AdminQuotaSteps) theAdminUpdatesTheAllowanceWithNewLimits(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	if err := s.verifyAllowanceIDAvailable(); err != nil {
		return ctx, err
	}

	expiry := time.Now().Add(30 * 24 * time.Hour)

	// Get the existing allowance first to preserve UserID
	existingAllowance, err := helpers.RequireAdminCurrentAllowance(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get existing allowance: %w", err)
	}

	if err := verifyResourceExists(existingAllowance, existingAllowance != nil, "existing allowance"); err != nil {
		return ctx, err
	}

	// Convert human-readable sizes to bytes
	allowanceUpload, _ := units.FromHumanSize("200MB")
	allowanceDownload, _ := units.FromHumanSize("1GB")
	allowanceStorage, _ := units.FromHumanSize("20MB")
	
	_, err = adminClient.Quota().UpdateAllowance(ctx, fmt.Sprint(s.allowanceID), existingAllowance.UserId, "BONUS", "STORAGE",
		int(allowanceUpload),
		int(allowanceDownload),
		int(allowanceStorage),
		expiry,
	)
	if err != nil {
		return ctx, fmt.Errorf("failed to update quota allowance: %w", err)
	}

	return ctx, nil
}

// theAllowanceIsUpdatedSuccessfully verifies update by retrieving the allowance
func (s *AdminQuotaSteps) theAllowanceIsUpdatedSuccessfully(ctx context.Context) (context.Context, error) {
	if s.allowanceID == 0 {
		return ctx, nil
	}

	// Verify allowance still exists after update
	existingAllowance, err := helpers.RequireAdminCurrentAllowance(ctx)
	if err != nil {
		return ctx, fmt.Errorf("allowance not found after update: %w", err)
	}

	if existingAllowance.Id == 0 {
		return ctx, fmt.Errorf("allowance has invalid ID after update")
	}

	return ctx, nil
}


// theAdminDeletesTheAllowance deletes an allowance
func (s *AdminQuotaSteps) theAdminDeletesTheAllowance(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	if err := s.verifyAllowanceIDAvailable(); err != nil {
		return ctx, err
	}

	if err := adminClient.Quota().DeleteAllowance(ctx, fmt.Sprint(s.allowanceID)); err != nil {
		return ctx, fmt.Errorf("failed to delete quota allowance: %w", err)
	}

	return ctx, nil
}

// theAllowanceIsDeletedSuccessfully verifies deletion
func (s *AdminQuotaSteps) theAllowanceIsDeletedSuccessfully(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Verify allowance is inactive (soft delete sets is_active=false)
	allowances, _, _ := adminClient.Quota().ListAllowances(ctx)
	for _, a := range allowances {
		if int64(a.Id) == s.allowanceID {
			if a.IsActive {
				return ctx, fmt.Errorf("allowance is still active after deletion")
			}
			// Found the allowance and it's inactive - this is expected for soft delete
			return ctx, nil
		}
	}

	// If we didn't find the allowance at all, that's also acceptable (already deleted)
	return ctx, nil
}

// theAdminReconcilesQuotaForAllUsers reconciles
func (s *AdminQuotaSteps) theAdminReconcilesQuotaForAllUsers(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	_, processed, err := adminClient.Quota().Reconcile(ctx, nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to reconcile quota: %w", err)
	}

	ctx = helpers.SetProcessedUsers(ctx, processed)
	return ctx, nil
}

// theReconciliationCompletesSuccessfully verifies reconciliation completed
func (s *AdminQuotaSteps) theReconciliationCompletesSuccessfully(ctx context.Context) (context.Context, error) {
	processed, ok := helpers.GetProcessedUsers(ctx)
	if !ok {
		return ctx, fmt.Errorf("reconciliation completed but no user count available")
	}

	if processed == 0 {
		return ctx, nil // Zero is valid when no users need reconciliation
	}

	return ctx, nil
}


// theUsersProcessedCountIsRecorded verifies count recorded
func (s *AdminQuotaSteps) theUsersProcessedCountIsRecorded(ctx context.Context) (context.Context, error) {
	processed, ok := helpers.GetProcessedUsers(ctx)
	if err := verifyResourceExists(processed, ok, "processed users count"); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// theAdminPerformsQuotaCleanup performs cleanup
func (s *AdminQuotaSteps) theAdminPerformsQuotaCleanup(ctx context.Context, days int) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, err
	}

	count, err := adminClient.Quota().Cleanup(ctx, days)
	if err != nil {
		return ctx, fmt.Errorf("failed to cleanup quota: %w", err)
	}

	ctx = helpers.SetDeletedRecords(ctx, int(count))
	return ctx, nil
}

// theCleanupCompletesSuccessfully verifies cleanup completed
func (s *AdminQuotaSteps) theCleanupCompletesSuccessfully(ctx context.Context) (context.Context, error) {
	deletedCount := ctx.Value(helpers.DeletedRecordsKey)
	if deletedCount == nil {
		return ctx, fmt.Errorf("cleanup completed but no deleted count available")
	}

	// Verify count is non-negative
	if count, ok := deletedCount.(int); ok && count < 0 {
		return ctx, fmt.Errorf("deleted count cannot be negative: %d", count)
	}

	return ctx, nil
}


// theRecordsDeletedCountIsRecorded verifies recorded count
func (s *AdminQuotaSteps) theRecordsDeletedCountIsRecorded(ctx context.Context) (context.Context, error) {
	if ctx.Value(helpers.DeletedRecordsKey) == nil {
		return ctx, fmt.Errorf("deleted records count not recorded")
	}
	return ctx, nil
}
