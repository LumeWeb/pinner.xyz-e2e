package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
	admin "go.lumeweb.com/portal-sdk/admin"
)

// SiaQuotaSteps holds step definitions for Sia quota and funding operations
type SiaQuotaSteps struct {
	quotaExceeded bool
}

// NewSiaQuotaSteps creates a new SiaQuotaSteps instance
func NewSiaQuotaSteps() *SiaQuotaSteps {
	return &SiaQuotaSteps{}
}

// InitializeScenario registers all Sia quota step definitions with godog
func (s *SiaQuotaSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the user has a Sia storage quota limit$`, s.theUserHasAQuotaLimitSet)
	ctx.Step(`^the user uploads data exceeding the Sia quota$`, s.theUserUploadsDataExceedingQuotaViaPortal)
	ctx.Step(`^the SIA upload is blocked with quota exceeded$`, s.theUploadIsBlockedWithQuotaExceeded)
	ctx.Step(`^the user has reached their Sia storage quota$`, s.theUserHasAQuotaLimitReached)
	ctx.Step(`^the SIA account is credited with additional storage$`, s.fundingIsAddedToTheAccount)
	ctx.Step(`^the user can resume Sia uploads up to the new quota$`, s.theUserCanResumeUploadsUpToTheNewQuota)
}

// theUserHasAQuotaLimitSet ensures the user has a quota plan assigned via admin API
func (s *SiaQuotaSteps) theUserHasAQuotaLimitSet(ctx context.Context) (context.Context, error) {
	ctx, err := helpers.CreateAdminClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to create admin client: %w", err)
	}

	ctx, planID, err := helpers.EnsureDefaultQuotaPlan(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to ensure default quota plan: %w", err)
	}

	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get admin client: %w", err)
	}

	userAPI, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get authenticated client: %w", err)
	}

	accountInfo, err := userAPI.GetAccount(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get account info: %w", err)
	}

	userID := int(accountInfo.Id)
	planIDInt := int(planID)

	config := &admin.UserQuotaConfigUpdate{
		QuotaPlanId: &planIDInt,
	}

	_, err = adminClient.Quota().UpdateUserConfig(ctx, userID, config)
	if err != nil {
		return ctx, fmt.Errorf("failed to assign quota plan to user: %w", err)
	}

	return ctx, nil
}

// theUserUploadsDataExceedingQuotaViaPortal attempts to upload data that exceeds the quota
func (s *SiaQuotaSteps) theUserUploadsDataExceedingQuotaViaPortal(ctx context.Context) (context.Context, error) {
	s.quotaExceeded = false

	_, err := helpers.SiaUploadRandomData(ctx, 1024*1024*1024)
	if err != nil {
		s.quotaExceeded = true
		return ctx, nil
	}

	return ctx, fmt.Errorf("expected upload to be blocked by quota, but it succeeded")
}

// theUploadIsBlockedWithQuotaExceeded verifies the upload was blocked due to quota
func (s *SiaQuotaSteps) theUploadIsBlockedWithQuotaExceeded(ctx context.Context) (context.Context, error) {
	if !s.quotaExceeded {
		return ctx, fmt.Errorf("expected upload to be blocked with quota exceeded, but it was not")
	}
	return ctx, nil
}

// theUserHasAQuotaLimitReached sets up a user whose quota is at its limit
func (s *SiaQuotaSteps) theUserHasAQuotaLimitReached(ctx context.Context) (context.Context, error) {
	ctx, err := s.theUserHasAQuotaLimitSet(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to set quota limit: %w", err)
	}

	_, err = helpers.SiaGetAccount(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get account for quota check: %w", err)
	}

	return ctx, nil
}

// fundingIsAddedToTheAccount adds funding to the user's account via admin billing API
func (s *SiaQuotaSteps) fundingIsAddedToTheAccount(ctx context.Context) (context.Context, error) {
	adminClient, err := helpers.RequireAdminClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get admin client: %w", err)
	}

	billingAdmin := adminClient.Billing()
	if billingAdmin == nil {
		return ctx, fmt.Errorf("admin client's Billing() returned nil")
	}

	userAPI, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get authenticated client: %w", err)
	}

	accountInfo, err := userAPI.GetAccount(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get account info: %w", err)
	}

	userID := int(accountInfo.Id)
	desc := "E2E test funding for SIA quota"
	refId := fmt.Sprintf("sia-e2e-%d", userID)
	refType := "manual"

	_, err = billingAdmin.CreateCredit(ctx, &admin.CreditCreateRequest{
		UserId:        userID,
		Amount:        "10.00",
		Type:          "manual_adjustment",
		Direction:     "credit",
		Description:   &desc,
		ReferenceId:   &refId,
		ReferenceType: &refType,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to add funding to account: %w", err)
	}

	return ctx, nil
}

// theUserCanResumeUploadsUpToTheNewQuota verifies the user can upload after funding
func (s *SiaQuotaSteps) theUserCanResumeUploadsUpToTheNewQuota(ctx context.Context) (context.Context, error) {
	ctx, err := helpers.SiaUploadRandomData(ctx, 1024)
	if err != nil {
		return ctx, fmt.Errorf("expected upload to succeed after funding, but it failed: %w", err)
	}

	return ctx, nil
}
