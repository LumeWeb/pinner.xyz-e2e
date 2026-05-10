package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

// SiaAppAccountSteps holds step definitions for Sia app account management
type SiaAppAccountSteps struct{}

// NewSiaAppAccountSteps creates a new SiaAppAccountSteps instance
func NewSiaAppAccountSteps() *SiaAppAccountSteps {
	return &SiaAppAccountSteps{}
}

// InitializeScenario registers all Sia app account step definitions with godog
func (s *SiaAppAccountSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the user creates a new Sia app account$`, s.theUserCreatesANewSIAAppAccount)
	ctx.Step(`^the SIA app account is linked to the user$`, s.theAppAccountIsLinkedToTheUser)
}

// anAuthenticatedUserExists registers and logs in a test user
func (s *SiaAppAccountSteps) anAuthenticatedUserExists(ctx context.Context) (context.Context, error) {
	ctx, err := helpers.RegisterAndLoginTestUser(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to register and login test user: %w", err)
	}
	return ctx, nil
}

// theUserCreatesANewSIAAppAccount performs the full Sia connect flow to create an app account
func (s *SiaAppAccountSteps) theUserCreatesANewSIAAppAccount(ctx context.Context) (context.Context, error) {
	ctx, err := helpers.SiaSDKConnect(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to create SIA app account: %w", err)
	}
	return ctx, nil
}

// theAppAccountIsLinkedToTheUser verifies the app account is linked by checking SDK and app key
func (s *SiaAppAccountSteps) theAppAccountIsLinkedToTheUser(ctx context.Context) (context.Context, error) {
	sdk, ok := helpers.GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return ctx, fmt.Errorf("SIA app account not linked: no SDK in context")
	}

	appKey, ok := helpers.GetSiaAppKey(ctx)
	if !ok || appKey == "" {
		return ctx, fmt.Errorf("SIA app account not linked: no app key in context")
	}

	return ctx, nil
}
