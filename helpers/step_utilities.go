package helpers

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	account "go.lumeweb.com/portal-sdk"
)

// BaseSteps provides common functionality and patterns for all step definition types.
// This struct DRYs up the implementation details shared across auth, account, password,
// and profile step definitions.
type BaseSteps struct {
	// accountAPI is the unauthenticated client for initial requests
	accountAPI account.AccountAPI

	// portalHost is the vhost hostname for API requests
	portalHost string

	// portalEndpoint is the full endpoint URL for API requests
	portalEndpoint string

	// portalTarget is the target address for connection
	portalTarget string
}

// InitializeStepCommon initializes the BaseSteps with common configuration.
// This consolidates the client setup pattern used across all step types.
func InitializeStepCommon() BaseSteps {
	return BaseSteps{
		accountAPI:     GetUnauthenticatedClient(),
		portalHost:     GetPortalHost(),
		portalEndpoint: GetPortalEndpoint(),
		portalTarget:   GetPortalTarget(),
	}
}

// RequireAPIKey retrieves the API key from context or returns an error.
// This DRYs up the common error checking pattern used across step implementations.
func RequireAPIKey(ctx context.Context) (string, error) {
	apiKey, ok := GetAPIKey(ctx)
	if !ok || apiKey == "" {
		return "", fmt.Errorf("no API key available")
	}
	return apiKey, nil
}

// RequireAPIKeyUUID retrieves the API key UUID from context or returns an error.
// This DRYs up the common error checking pattern used across step implementations.
func RequireAPIKeyUUID(ctx context.Context) (string, error) {
	uuid, ok := GetAPIKeyUUID(ctx)
	if !ok || uuid == "" {
		return "", fmt.Errorf("no API key UUID available")
	}
	return uuid, nil
}

// RequireOTPSecret retrieves the OTP secret from context or returns an error.
// This DRYs up the common error checking pattern used across step implementations.
func RequireOTPSecret(ctx context.Context) (string, error) {
	secret, ok := GetOTPSecret(ctx)
	if !ok {
		return "", fmt.Errorf("no OTP secret in context")
	}
	// Allow empty string - secret may not have been required/generated yet
	return secret, nil
}

// RequirePasswordResetToken retrieves the password reset token from context or returns an error.
// This DRYs up the common error checking pattern used across step implementations.
func RequirePasswordResetToken(ctx context.Context) (string, error) {
	token, ok := GetPasswordResetToken(ctx)
	if !ok || token == "" {
		return "", fmt.Errorf("no password reset token available")
	}
	return token, nil
}

// RequireVerificationToken retrieves the verification token from context or returns an error.
// This DRYs up the common error checking pattern used across step implementations.
func RequireVerificationToken(ctx context.Context) (string, error) {
	token, ok := GetVerificationToken(ctx)
	if !ok || token == "" {
		return "", fmt.Errorf("no verification token available")
	}
	return token, nil
}

// RequireJWTToken retrieves the JWT token from context or returns an error.
// This DRYs up the common error checking pattern used across step implementations.
func RequireJWTToken(ctx context.Context) (string, error) {
	token, ok := GetJWTToken(ctx)
	if !ok || token == "" {
		return "", fmt.Errorf("no JWT token available")
	}
	return token, nil
}

// RequireAPIKeysList retrieves the API keys list from context or returns an error.
// This DRYs up the common error checking pattern used across step implementations.
func RequireAPIKeysList(ctx context.Context) ([]*account.APIKey, error) {
	apiKeys, ok := GetAPIKeysList(ctx)
	if !ok || apiKeys == nil {
		return nil, fmt.Errorf("no API keys list available")
	}
	return apiKeys, nil
}

// StepError creates a standardized error message for step failures.
// This provides consistent error formatting across all step implementations.
func StepError(step, message string) error {
	return fmt.Errorf("%s: %s", step, message)
}

// StepSuccess is a no-op function for steps that just need to return success.
// This makes step implementations clearer and maintains consistency.
func StepSuccess() error {
	return nil
}

// ContextSuccess is a no-op function for steps that just need to return context and success.
// This makes step implementations clearer and maintains consistency.
func ContextSuccess(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

// WithAuthenticatedClient retrieves the authenticated client from context or returns an error.
// This helper wraps the common "get and check" pattern used across step implementations.
func WithAuthenticatedClient(ctx context.Context, step string) (account.AccountAPI, context.Context, error) {
	api := GetAuthenticatedClientFromContext(ctx)
	if api == nil {
		return nil, ctx, StepError(step, "no authenticated client available")
	}
	return api, ctx, nil
}

// RegisterCommonStepPattern registers the step definition and verifies the previous step succeeded.
// This DRYs up the common pattern for steps that just verify the previous step completed.
func RegisterCommonStepPattern(ctx *godog.ScenarioContext, stepPattern string, handlerFunc func(context.Context) error) {
	ctx.Step(stepPattern, func(ctx context.Context) error {
		return handlerFunc(ctx)
	})
}

// CreateValidTOTP generates a valid TOTP code using the stored OTP secret.
// This consolidates the TOTP creation pattern used in 2FA-related steps.
func CreateValidTOTP(ctx context.Context) (string, error) {
	// First try to get the secret from context
	secret, hasSecret := GetOTPSecret(ctx)
	if hasSecret && secret != "" {
		// Secret exists in context, use it
		code, err := GenerateValidTOTPCode(secret)
		if err != nil {
			return "", fmt.Errorf("failed to generate TOTP code: %w", err)
		}
		return code, nil
	}
	
	return "", fmt.Errorf("no OTP secret available in context")
}

// RegisterAndLogin creates a new test user and logs them in.
// This DRYs up the common pattern of registration followed by login.
func RegisterAndLogin(ctx context.Context) (context.Context, error) {
	ctx, err := RegisterAndLoginTestUser(ctx)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

// StoreAPIKeyWithUUID stores an API key and its UUID in context, tracked for cleanup.
// This DRYs up the common API key storage pattern used across step implementations.
func StoreAPIKeyWithUUID(ctx context.Context, apiKey *account.APIKey) context.Context {
	uuidStr := apiKey.Uuid.String()
	ctx = AddAPIKeyUUIDCleanup(ctx, uuidStr)
	ctx = SetAPIKeyUUID(ctx, uuidStr)
	ctx = SetAPIKey(ctx, apiKey.Token)
	return ctx
}

// VerifyStringValue verifies that a string matches an expected value and returns an error if not.
// This DRYs up common string comparison validation across step implementations.
func VerifyStringValue(actual, expected, fieldName string) error {
	if actual != expected {
		return fmt.Errorf("expected %s='%s', got '%s'", fieldName, expected, actual)
	}
	return nil
}

// VerifyStringNotEmpty verifies a string is not empty and returns an error if it is.
// This DRYs up common empty value validation across step implementations.
func VerifyStringNotEmpty(value, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s is empty", fieldName)
	}
	return nil
}

// VerifyContains verifies that a string contains a substring and returns an error if not.
// This DRYs up common string containment validation across step implementations.
func VerifyContains(text, substring, fieldName string) error {
	if !strings.Contains(text, substring) {
		return fmt.Errorf("%s does not contain '%s': %s", fieldName, substring, text)
	}
	return nil
}

// VerifyStringLength verifies a string's length is within bounds and returns an error if not.
// This DRYs up common string length validation across step implementations.
func VerifyStringLength(value string, minLength, maxLength int, fieldName string) error {
	if len(value) < minLength {
		return fmt.Errorf("%s length %d is less than minimum %d", fieldName, len(value), minLength)
	}
	if maxLength > 0 && len(value) > maxLength {
		return fmt.Errorf("%s length %d exceeds maximum %d", fieldName, len(value), maxLength)
	}
	return nil
}
