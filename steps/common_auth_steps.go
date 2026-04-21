package steps

import (
	"context"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

// CommonAuthSteps holds shared authentication step definitions
// that are used across multiple step files
// First-registered step wins in godog, so this file must initialize BEFORE
// service-specific step files to ensure shared steps are available
type CommonAuthSteps struct{}

// NewCommonAuthSteps creates a new CommonAuthSteps instance
func NewCommonAuthSteps() *CommonAuthSteps {
	return &CommonAuthSteps{}
}

// InitializeScenario registers all common auth step definitions with godog
// Must be called BEFORE other step files that depend on these steps
func (s *CommonAuthSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// User creation and login steps
	ctx.Step(`^an existing registered user$`, s.anExistingRegisteredUser)
	ctx.Step(`^the user is logged in$`, s.theUserIsLoggedIn)
}

// anExistingRegisteredUser creates and registers a new test user
// Used by scenarios that need a fresh user account before testing other features
func (s *CommonAuthSteps) anExistingRegisteredUser(ctx context.Context) (context.Context, error) {
	return helpers.RegisterTestUser(ctx)
}

// theUserIsLoggedIn authenticates the test user and stores credentials in context
// Ensures subsequent steps have access to authenticated API clients
func (s *CommonAuthSteps) theUserIsLoggedIn(ctx context.Context) (context.Context, error) {
	return helpers.LoginTestUser(ctx)
}
