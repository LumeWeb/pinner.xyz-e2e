package steps

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cucumber/godog"

	"pinner.xyz-e2e/helpers"
	account "go.lumeweb.com/portal-sdk"
)

// PasswordResetSteps provides step definitions for password reset and email verification
type PasswordResetSteps struct {
	accountAPI account.AccountAPI
	maildev    *helpers.MailDevClient
}

// NewPasswordResetSteps creates a new PasswordResetSteps instance
func NewPasswordResetSteps() *PasswordResetSteps {
	return &PasswordResetSteps{}
}

// InitializeScenario registers steps with the scenario context
func (s *PasswordResetSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Password reset steps
	ctx.Step(`^an existing registered user$`, s.anExistingRegisteredUser)
	ctx.Step(`^the user requests a password reset for their email$`, s.userRequestsPasswordReset)
	ctx.Step(`^the password reset request is successful$`, s.passwordResetRequestSuccessful)
	ctx.Step(`^the user has requested a password reset$`, s.userHasRequestedPasswordReset)
	ctx.Step(`^the password reset email has been received with a token$`, s.passwordResetEmailReceived)
	ctx.Step(`^the user resets their password using the token$`, s.userResetsPasswordWithToken)
	ctx.Step(`^the password is updated successfully$`, s.passwordUpdatedSuccessfully)
	ctx.Step(`^the user can login with the new password$`, s.userCanLoginWithNewPassword)
	ctx.Step(`^the user attempts to reset their password with an invalid token$`, s.userAttemptsResetWithInvalidToken)
	ctx.Step(`^the password reset fails$`, s.passwordResetFails)

	// Email verification steps
	ctx.Step(`^a newly registered user$`, s.aNewlyRegisteredUser)
	ctx.Step(`^the email verification email has been received with a token$`, s.verificationEmailReceived)
	ctx.Step(`^the user verifies their email using the token$`, s.userVerifiesEmailWithToken)
	ctx.Step(`^the email is verified successfully$`, s.emailVerifiedSuccessfully)
	ctx.Step(`^the user attempts to verify their email with an invalid token$`, s.userAttemptsVerifyWithInvalidToken)
	ctx.Step(`^the email verification fails$`, s.emailVerificationFails)
	ctx.Step(`^the user requests a new verification email$`, s.userRequestsNewVerificationEmail)
	ctx.Step(`^the verification email is sent successfully$`, s.verificationEmailSentSuccessfully)
	ctx.Step(`^the verification email is received$`, s.verificationEmailReceived)
}

func (s *PasswordResetSteps) anExistingRegisteredUser(ctx context.Context) (context.Context, error) {
	return helpers.RegisterTestUser(ctx)
}

func (s *PasswordResetSteps) userRequestsPasswordReset(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	api := helpers.GetUnauthenticatedClient()
	err = api.RequestPasswordReset(ctx, testUser.Email)
	if err != nil {
		return ctx, fmt.Errorf("failed to request password reset: %w", err)
	}

	return ctx, nil
}

func (s *PasswordResetSteps) passwordResetRequestSuccessful(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}

func (s *PasswordResetSteps) userHasRequestedPasswordReset(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	api := helpers.GetUnauthenticatedClient()
	err = api.RequestPasswordReset(ctx, testUser.Email)
	if err != nil {
		return ctx, fmt.Errorf("failed to request password reset: %w", err)
	}

	return ctx, nil
}

func (s *PasswordResetSteps) passwordResetEmailReceived(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	maildev := helpers.NewMailDevClient("")

	// Wait for password reset email
	portalName := helpers.GetPortalName()
	subject := fmt.Sprintf("Reset Your Password for %s", portalName)
	email, err := maildev.WaitForEmail(testUser.Email, subject, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("failed to receive password reset email: %w", err)
	}

	// Extract token from email content
	token, err := maildev.ExtractTokenFromEmail(email, helpers.DefaultTokenPattern)
	if err != nil {
		return ctx, fmt.Errorf("failed to extract token from email: %w", err)
	}

	ctx = helpers.SetPasswordResetToken(ctx, token)
	return ctx, nil
}

func (s *PasswordResetSteps) userResetsPasswordWithToken(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	token, err := helpers.RequirePasswordResetToken(ctx)
	if err != nil {
		return ctx, err
	}

	api := helpers.GetUnauthenticatedClient()

	// Reset password with token
	newPassword := "NewPassword123!"
	err = api.ConfirmPasswordReset(ctx, testUser.Email, token, newPassword)
	if err != nil {
		return ctx, fmt.Errorf("failed to reset password: %w", err)
	}

	// Update test user password for verification
	testUser.Password = newPassword
	ctx = helpers.SetTestUser(ctx, testUser)

	return ctx, nil
}

func (s *PasswordResetSteps) passwordUpdatedSuccessfully(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}

func (s *PasswordResetSteps) userCanLoginWithNewPassword(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	api := helpers.GetUnauthenticatedClient()

	// Try to login with new password
	loginResult, err := api.Login(ctx, testUser.Email, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to login with new password: %w", err)
	}

	if loginResult.Token == "" {
		return ctx, fmt.Errorf("login succeeded but no token returned")
	}

	return ctx, nil
}

func (s *PasswordResetSteps) userAttemptsResetWithInvalidToken(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	api := helpers.GetUnauthenticatedClient()

	// Try to reset password with invalid token
	err = api.ConfirmPasswordReset(ctx, testUser.Email, os.Getenv("TEST_INVALID_TOKEN"), "InvalidPassword123!")
	if err == nil {
		return ctx, fmt.Errorf("password reset with invalid token should have failed")
	}

	return ctx, nil
}

func (s *PasswordResetSteps) passwordResetFails(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}

func (s *PasswordResetSteps) aNewlyRegisteredUser(ctx context.Context) (context.Context, error) {
	return helpers.RegisterTestUser(ctx)
}

func (s *PasswordResetSteps) verificationEmailReceived(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	// Request verification email first
	api := helpers.GetUnauthenticatedClient()
	err = api.ResendVerifyEmail(ctx, testUser.Email)
	if err != nil {
		return ctx, fmt.Errorf("failed to send verification email: %w", err)
	}

	maildev := helpers.NewMailDevClient("")

	// Wait for verification email
	portalName := helpers.GetPortalName()
	subject := fmt.Sprintf("Verify Your Email for %s", portalName)
	email, err := maildev.WaitForEmail(testUser.Email, subject, 30*time.Second)
	if err != nil {
		return ctx, fmt.Errorf("failed to receive verification email: %w", err)
	}

	// Extract token from email content
	token, err := maildev.ExtractTokenFromEmail(email, helpers.DefaultTokenPattern)
	if err != nil {
		return ctx, fmt.Errorf("failed to extract token from email: %w", err)
	}

	ctx = helpers.SetVerificationToken(ctx, token)
	return ctx, nil
}

func (s *PasswordResetSteps) userVerifiesEmailWithToken(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	token, err := helpers.RequireVerificationToken(ctx)
	if err != nil {
		return ctx, err
	}

	api := helpers.GetUnauthenticatedClient()

	// Verify email with token
	err = api.VerifyEmail(ctx, testUser.Email, token)
	if err != nil {
		return ctx, fmt.Errorf("failed to verify email: %w", err)
	}

	return ctx, nil
}

func (s *PasswordResetSteps) emailVerifiedSuccessfully(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}

func (s *PasswordResetSteps) userAttemptsVerifyWithInvalidToken(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	api := helpers.GetUnauthenticatedClient()

	// Try to verify email with invalid token
	err = api.VerifyEmail(ctx, testUser.Email, os.Getenv("TEST_INVALID_TOKEN"))
	if err == nil {
		return ctx, fmt.Errorf("email verification with invalid token should have failed")
	}

	return ctx, nil
}

func (s *PasswordResetSteps) emailVerificationFails(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}

func (s *PasswordResetSteps) userRequestsNewVerificationEmail(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	api := helpers.GetUnauthenticatedClient()

	// Request new verification email
	err = api.ResendVerifyEmail(ctx, testUser.Email)
	if err != nil {
		return ctx, fmt.Errorf("failed to resend verification email: %w", err)
	}

	return ctx, nil
}

func (s *PasswordResetSteps) verificationEmailSentSuccessfully(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}
