package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

// CommonPasswordSteps holds shared password verification step definitions
// that are used across multiple step files
// First-registered step wins in godog, so this file must initialize BEFORE
// service-specific step files to ensure shared steps are available
type CommonPasswordSteps struct{}

// NewCommonPasswordSteps creates a new CommonPasswordSteps instance
func NewCommonPasswordSteps() *CommonPasswordSteps {
	return &CommonPasswordSteps{}
}

// InitializeScenario registers all common password step definitions with godog
// Must be called BEFORE other step files that depend on these steps
func (s *CommonPasswordSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Password verification steps
	ctx.Step(`^the user can login with the new password$`, s.userCanLoginWithNewPassword)
}

// userCanLoginWithNewPassword verifies the user can login with the current password in context
// Used after password reset or password change flows
func (s *CommonPasswordSteps) userCanLoginWithNewPassword(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	api := helpers.GetUnauthenticatedClient()
	_, err = api.Login(ctx, testUser.Email, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to login with new password: %w", err)
	}

	return ctx, nil
}
