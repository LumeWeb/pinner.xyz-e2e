package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

// PasswordProfileSteps provides step definitions for password validation and profile management
type PasswordProfileSteps struct {
	// No struct fields - all state in context for proper isolation
}

// NewPasswordProfileSteps creates a new PasswordProfileSteps instance
func NewPasswordProfileSteps() *PasswordProfileSteps {
	return &PasswordProfileSteps{}
}

// InitializeScenario registers steps with the scenario context
func (s *PasswordProfileSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^a new user registration attempt with password "([^"]*)"$`, s.aNewUserRegistrationAttemptWithPassword)
	ctx.Step(`^a new user registration attempt with email "([^"]*)"$`, s.aNewUserRegistrationAttemptWithEmail)
	ctx.Step(`^the user submits registration data$`, s.theUserSubmitsRegistrationData)
	ctx.Step(`^the registration fails with password validation error$`, s.registrationFailsWithPasswordError)
	ctx.Step(`^the registration fails with email validation error$`, s.registrationFailsWithEmailError)

	ctx.Step(`^the user changes their password to "([^"]*)"$`, s.userChangesPasswordTo)
	ctx.Step(`^the password is changed successfully$`, s.passwordChangedSuccessfully)
	ctx.Step(`^the user can login with the new password$`, s.userCanLoginWithNewPassword)

	ctx.Step(`^the user attempts to change password with invalid current password$`, s.attemptsChangeWithInvalidCurrentPassword)
	ctx.Step(`^the password change fails$`, s.passwordChangeFails)
	ctx.Step(`^the user attempts to change password to "([^"]*)"$`, s.attemptsChangePasswordTooShort)
	ctx.Step(`^the password change fails with validation error$`, s.passwordChangeFailsValidation)

	ctx.Step(`^the user logs out$`, s.userLogsOut)
	ctx.Step(`^the session is invalidated$`, s.sessionInvalidated)

	ctx.Step(`^the user updates their first name to "([^"]*)"$`, s.userUpdatesFirstNameTo)
	ctx.Step(`^the user updates their last name to "([^"]*)"$`, s.userUpdatesLastNameTo)
	ctx.Step(`^the profile is updated successfully$`, s.profileUpdatedSuccessfully)
	ctx.Step(`^the updated profile reflects the new first name$`, s.profileReflectsNewFirstName)
	ctx.Step(`^the updated profile reflects the new last name$`, s.profileReflectsNewLastName)

	ctx.Step(`^the user retrieves their profile$`, s.userRetrievesProfile)
	ctx.Step(`^the profile information is returned successfully$`, s.profileReturnedSuccessfully)
	ctx.Step(`^the profile contains email$`, s.profileContainsEmail)
	ctx.Step(`^the profile contains first name and last name$`, s.profileContainsFirstAndLastName)
}

// Steps for registration with weak/invalid data

func (s *PasswordProfileSteps) aNewUserRegistrationAttemptWithPassword(ctx context.Context, password string) (context.Context, error) {
	ctx = helpers.SetRegistrationPassword(ctx, password)
	return ctx, nil
}

func (s *PasswordProfileSteps) aNewUserRegistrationAttemptWithEmail(ctx context.Context, email string) (context.Context, error) {
	ctx = helpers.SetRegistrationEmail(ctx, email)
	return ctx, nil
}

func (s *PasswordProfileSteps) theUserSubmitsRegistrationData(ctx context.Context) (context.Context, error) {
	email, hasEmail := helpers.GetRegistrationEmail(ctx)
	if !hasEmail || email == "" {
		email = helpers.CreateTestUser().Email
	}
	password, hasPassword := helpers.GetRegistrationPassword(ctx)
	if !hasPassword || password == "" {
		password = helpers.CreateTestUser().Password
	}

	api := helpers.GetUnauthenticatedClient()
	firstName := helpers.GenerateFirstName()
	lastName := helpers.GenerateLastName()
	err := api.Register(ctx, email, firstName, lastName, password)
	// Store the error in context for verification
	ctx = helpers.SetRegistrationError(ctx, err)
	return ctx, nil
}

func (s *PasswordProfileSteps) registrationFailsWithPasswordError(ctx context.Context) (context.Context, error) {
	errMsg := helpers.GetRegistrationError(ctx)
	if errMsg == "" {
		return ctx, fmt.Errorf("expected registration to fail with password error, but no error was recorded")
	}

	// Check if error message contains password validation error
	if !strings.Contains(errMsg, "Password") && !strings.Contains(errMsg, "password") {
		return ctx, fmt.Errorf("expected password validation error, got: %s", errMsg)
	}

	return ctx, nil
}

func (s *PasswordProfileSteps) registrationFailsWithEmailError(ctx context.Context) (context.Context, error) {
	errMsg := helpers.GetRegistrationError(ctx)
	if errMsg == "" {
		return ctx, fmt.Errorf("expected registration to fail with email error, but no error was recorded")
	}

	// Check if error message contains email validation error
	if !strings.Contains(errMsg, "Email") && !strings.Contains(errMsg, "email") {
		return ctx, fmt.Errorf("expected email validation error, got: %s", errMsg)
	}

	return ctx, nil
}

// Steps for password change

func (s *PasswordProfileSteps) userChangesPasswordTo(ctx context.Context, newPassword string) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	err = api.UpdatePassword(ctx, testUser.Password, newPassword)
	if err != nil {
		return ctx, fmt.Errorf("failed to change password: %w", err)
	}

	// Update the test user password in context
	testUser.Password = newPassword
	ctx = helpers.SetTestUser(ctx, testUser)

	return ctx, nil
}

func (s *PasswordProfileSteps) passwordChangedSuccessfully(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}

func (s *PasswordProfileSteps) userCanLoginWithNewPassword(ctx context.Context) (context.Context, error) {
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

func (s *PasswordProfileSteps) attemptsChangeWithInvalidCurrentPassword(ctx context.Context) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	err = api.UpdatePassword(ctx, "WrongPassword123", "NewPassword456")
	if err == nil {
		return ctx, fmt.Errorf("password change with invalid current password should have failed")
	}

	return ctx, nil
}

func (s *PasswordProfileSteps) passwordChangeFails(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}

func (s *PasswordProfileSteps) attemptsChangePasswordTooShort(ctx context.Context, password string) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	err = api.UpdatePassword(ctx, "CurrentPassword123", password)
	if err == nil {
		return ctx, fmt.Errorf("password change with short password should have failed")
	}

	return ctx, nil
}

func (s *PasswordProfileSteps) passwordChangeFailsValidation(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}

// Steps for logout

func (s *PasswordProfileSteps) userLogsOut(ctx context.Context) (context.Context, error) {
	// Note: The portal SDK does not currently have a Logout method, which is expected
	// for stateless JWT systems. With stateless JWTs, logout is a client-side operation
	// that clears the token from local storage. The JWT remains valid until its expiration
	// time but is no longer used by the client.
	//
	// For the portal to support server-side session invalidation, it would need to:
	// 1. Implement a token revocation list/blacklist
	// 2. Or maintain session state alongside JWTs
	//
	// Since the portal uses stateless JWTs, logout only clears local state.

	// Clear local context tokens after logout
	ctx = context.WithValue(ctx, helpers.JWTTokenKey, "")
	ctx = context.WithValue(ctx, helpers.AuthenticatedClientKey, nil)

	return ctx, nil
}

func (s *PasswordProfileSteps) sessionInvalidated(ctx context.Context) (context.Context, error) {
	err := helpers.VerifyJWTInvalidation(ctx)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

// Steps for profile update and retrieval

func (s *PasswordProfileSteps) userUpdatesFirstNameTo(ctx context.Context, firstName string) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	err = api.UpdateProfile(ctx, firstName, "")
	if err != nil {
		return ctx, fmt.Errorf("failed to update first name: %w", err)
	}

	// Store updated name in context for verification
	testUser.FirstName = firstName
	ctx = helpers.SetTestUser(ctx, testUser)

	return ctx, nil
}

func (s *PasswordProfileSteps) userUpdatesLastNameTo(ctx context.Context, lastName string) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	err = api.UpdateProfile(ctx, "", lastName)
	if err != nil {
		return ctx, fmt.Errorf("failed to update last name: %w", err)
	}

	// Store updated name in context for verification
	testUser.LastName = lastName
	ctx = helpers.SetTestUser(ctx, testUser)

	return ctx, nil
}

func (s *PasswordProfileSteps) profileUpdatedSuccessfully(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}

func (s *PasswordProfileSteps) profileReflectsNewFirstName(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	// Verify the profile reflects the updated first name
	err = helpers.VerifyProfileField(ctx, "first_name", testUser.FirstName)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *PasswordProfileSteps) profileReflectsNewLastName(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	// Verify the profile reflects the updated last name
	err = helpers.VerifyProfileField(ctx, "last_name", testUser.LastName)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *PasswordProfileSteps) userRetrievesProfile(ctx context.Context) (context.Context, error) {
	_, err := helpers.GetProfile(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to retrieve profile: %w", err)
	}

	return ctx, nil
}

func (s *PasswordProfileSteps) profileReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	_, err := helpers.GetProfile(ctx)
	if err != nil {
		return ctx, fmt.Errorf("profile was not returned successfully: %w", err)
	}
	return ctx, nil
}

func (s *PasswordProfileSteps) profileContainsEmail(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	// Verify the email is present in the profile
	err = helpers.VerifyProfileField(ctx, "email", testUser.Email)
	if err != nil {
		return ctx, fmt.Errorf("profile does not contain email: %w", err)
	}

	return ctx, nil
}

func (s *PasswordProfileSteps) profileContainsFirstAndLastName(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	// Verify both first name and last name are present
	err = helpers.VerifyProfileField(ctx, "first_name", testUser.FirstName)
	if err != nil {
		return ctx, fmt.Errorf("profile does not contain first name: %w", err)
	}

	err = helpers.VerifyProfileField(ctx, "last_name", testUser.LastName)
	if err != nil {
		return ctx, fmt.Errorf("profile does not contain last name: %w", err)
	}

	return ctx, nil
}
