package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	account "go.lumeweb.com/portal-sdk"
	"pinner.xyz-e2e/helpers"
)

// AuthSteps holds the state for authentication step definitions
type AuthSteps struct {
	accountAPI account.AccountAPI
}

// NewAuthSteps creates a new AuthSteps instance
func NewAuthSteps() *AuthSteps {
	return &AuthSteps{}
}

// InitializeScenario registers all step definitions with godog
func (s *AuthSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	helpers.RegisterCommonHooks(ctx)

	// Registration steps
	ctx.Step(`^a new user registration request$`, s.aNewUserRegistrationRequest)
	ctx.Step(`^the user submits valid registration data$`, s.theUserSubmitsValidRegistrationData)
	ctx.Step(`^the user is successfully registered$`, s.theUserIsSuccessfullyRegistered)
	ctx.Step(`^the user can login with the registered credentials$`, s.theUserCanLoginWithTheRegisteredCredentials)

	// Login with email/password steps
	ctx.Step(`^an existing registered user$`, s.anExistingRegisteredUser)
	ctx.Step(`^the user is logged in$`, s.theUserIsLoggedIn)
	ctx.Step(`^the user submits valid login credentials$`, s.theUserSubmitsValidLoginCredentials)
	ctx.Step(`^the user receives a valid JWT token$`, s.theUserReceivesAValidJWTToken)
	ctx.Step(`^the JWT token can be used for authenticated requests$`, s.theJWTTokenCanBeUsedForAuthenticatedRequests)

	// API key authentication steps
	ctx.Step(`^an existing registered user with an API key$`, s.anExistingRegisteredUserWithAnAPIKey)
	ctx.Step(`^the user authenticates using the API key$`, s.theUserAuthenticatesUsingTheAPIKey)

	// 2FA enable steps
	ctx.Step(`^the user requests to enable 2FA$`, s.theUserRequestsToEnable2FA)
	ctx.Step(`^an OTP secret is generated$`, s.anOTPSecretIsGenerated)
	ctx.Step(`^the user can verify OTP codes$`, s.theUserCanVerifyOTPCodes)
	ctx.Step(`^2FA is successfully enabled$`, s.twoFAIsSuccessfullyEnabled)

	// 2FA login steps
	ctx.Step(`^an existing registered user with 2FA enabled$`, s.anExistingRegisteredUserWith2FAEnabled)
	ctx.Step(`^the user is prompted for OTP verification$`, s.theUserIsPromptedForOTPVerification)
	ctx.Step(`^the user can login with valid OTP code$`, s.theUserCanLoginWithValidOTPCode)
	ctx.Step(`^invalid OTP codes are rejected$`, s.invalidOTPCodesAreRejected)

	// 2FA disable steps
	ctx.Step(`^the user requests to disable 2FA$`, s.theUserRequestsToDisable2FA)
	ctx.Step(`^2FA is successfully disabled$`, s.twoFAIsSuccessfullyDisabled)
	ctx.Step(`^the user can login without OTP verification$`, s.theUserCanLoginWithoutOTPVerification)

	// Setup account API client
	s.accountAPI = helpers.GetUnauthenticatedClient()
}

// Registration step implementations

func (s *AuthSteps) aNewUserRegistrationRequest(ctx context.Context) (context.Context, error) {
	testUser := helpers.CreateTestUser()
	return helpers.SetTestUser(ctx, testUser), nil
}

func (s *AuthSteps) theUserSubmitsValidRegistrationData(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	err = s.accountAPI.Register(ctx, testUser.Email, testUser.FirstName, testUser.LastName, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to register user: %w", err)
	}

	ctx = helpers.AddTestUserCleanup(ctx, testUser.Email)
	return ctx, nil
}

func (s *AuthSteps) theUserIsSuccessfullyRegistered(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}

func (s *AuthSteps) theUserCanLoginWithTheRegisteredCredentials(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	loginAPI := helpers.GetUnauthenticatedClient()
	loginResult, err := loginAPI.Login(ctx, testUser.Email, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to login with registered credentials: %w", err)
	}

	if loginResult == nil || loginResult.Token == "" {
		return ctx, fmt.Errorf("login did not return a token")
	}

	return helpers.SetJWTToken(ctx, loginResult.Token), nil
}

// Login with email/password step implementations

func (s *AuthSteps) anExistingRegisteredUser(ctx context.Context) (context.Context, error) {
	return helpers.RegisterTestUser(ctx)
}

func (s *AuthSteps) theUserIsLoggedIn(ctx context.Context) (context.Context, error) {
	return helpers.LoginTestUser(ctx)
}

func (s *AuthSteps) theUserSubmitsValidLoginCredentials(ctx context.Context) (context.Context, error) {
	return helpers.LoginTestUser(ctx)
}

func (s *AuthSteps) theUserReceivesAValidJWTToken(ctx context.Context) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Verify the token is valid by making a ping request
	err = api.Ping(ctx)
	if err != nil {
		return ctx, fmt.Errorf("JWT token is not valid: %w", err)
	}

	return ctx, nil
}

func (s *AuthSteps) theJWTTokenCanBeUsedForAuthenticatedRequests(ctx context.Context) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Ping to verify the token works
	err = api.Ping(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to make authenticated request: %w", err)
	}

	return ctx, nil
}

// API key authentication step implementations

func (s *AuthSteps) anExistingRegisteredUserWithAnAPIKey(ctx context.Context) (context.Context, error) {
	testUser := helpers.CreateTestUser()

	// Register the user
	err := s.accountAPI.Register(ctx, testUser.Email, testUser.FirstName, testUser.LastName, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to register user: %w", err)
	}

	ctx = helpers.AddTestUserCleanup(ctx, testUser.Email)

	// Login to get JWT
	loginAPI := helpers.GetUnauthenticatedClient()
	loginResult, err := loginAPI.Login(ctx, testUser.Email, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to login: %w", err)
	}

	// Store JWT token in context
	ctx = helpers.SetJWTToken(ctx, loginResult.Token)

	// Create authenticated client using helper
	authClient := helpers.CreateAuthenticatedClient(loginResult.Token)
	ctx = helpers.SetAuthenticatedClient(ctx, authClient)

	// Create API key with JWT
	apiKey, err := authClient.CreateAPIKey(ctx, "test-api-key")
	if err != nil {
		return ctx, fmt.Errorf("failed to create API key: %w", err)
	}

	// Store API key details in context using helper
	ctx = helpers.StoreAPIKeyWithUUID(ctx, apiKey)

	ctx = helpers.SetTestUser(ctx, testUser)
	return ctx, nil
}

func (s *AuthSteps) theUserAuthenticatesUsingTheAPIKey(ctx context.Context) (context.Context, error) {
	apiKey, err := helpers.RequireAPIKey(ctx)
	if err != nil {
		return ctx, err
	}

	loginAPI := helpers.GetUnauthenticatedClient()
	token, err := loginAPI.LoginWithAPIKey(ctx, apiKey)
	if err != nil {
		return ctx, fmt.Errorf("failed to authenticate with API key: %w", err)
	}

	if token == "" {
		return ctx, fmt.Errorf("API key login did not return a token")
	}

	return helpers.SetJWTToken(ctx, token), nil
}

// 2FA enable step implementations

func (s *AuthSteps) theUserRequestsToEnable2FA(ctx context.Context) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	secret, err := api.GenerateOTP(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to generate OTP secret: %w", err)
	}

	return helpers.SetOTPSecret(ctx, secret), nil
}

func (s *AuthSteps) anOTPSecretIsGenerated(ctx context.Context) (context.Context, error) {
	_, err := helpers.RequireOTPSecret(ctx)
	return ctx, err
}

func (s *AuthSteps) theUserCanVerifyOTPCodes(ctx context.Context) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Create valid TOTP code using helper
	testOTP, err := helpers.CreateValidTOTP(ctx)
	if err != nil {
		return ctx, err
	}

	err = api.VerifyOTP(ctx, testOTP)
	if err != nil {
		return ctx, fmt.Errorf("failed to verify OTP code: %w", err)
	}

	return ctx, nil
}

func (s *AuthSteps) twoFAIsSuccessfullyEnabled(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}

// 2FA login step implementations

func (s *AuthSteps) anExistingRegisteredUserWith2FAEnabled(ctx context.Context) (context.Context, error) {
	testUser := helpers.CreateTestUser()

	// Register the user
	err := s.accountAPI.Register(ctx, testUser.Email, testUser.FirstName, testUser.LastName, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to register user: %w", err)
	}

	ctx = helpers.AddTestUserCleanup(ctx, testUser.Email)

	// Login to get JWT
	loginAPI := helpers.GetUnauthenticatedClient()
	loginResult, err := loginAPI.Login(ctx, testUser.Email, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to login: %w", err)
	}

	// Create authenticated client using helper
	authClient := helpers.CreateAuthenticatedClient(loginResult.Token)
	ctx = helpers.SetAuthenticatedClient(ctx, authClient)

	// Enable 2FA
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	secret, err := api.GenerateOTP(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to generate OTP secret: %w", err)
	}

	// Store secret in context before creating TOTP (CreateValidTOTP needs it in context)
	ctx = helpers.SetOTPSecret(ctx, secret)

	// Verify OTP to enable 2FA using helper
	testOTP, err := helpers.CreateValidTOTP(ctx)
	if err != nil {
		return ctx, err
	}

	err = api.VerifyOTP(ctx, testOTP)
	if err != nil {
		return ctx, fmt.Errorf("failed to verify OTP: %w", err)
	}

	ctx = helpers.SetTestUser(ctx, testUser)
	return ctx, nil
}

func (s *AuthSteps) theUserIsPromptedForOTPVerification(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	loginAPI := helpers.GetUnauthenticatedClient()
	loginResult, err := loginAPI.Login(ctx, testUser.Email, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to login: %w", err)
	}

	if !loginResult.OTPRequired {
		return ctx, fmt.Errorf("OTP was not required as expected")
	}

	return ctx, nil
}

func (s *AuthSteps) theUserCanLoginWithValidOTPCode(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	loginAPI := helpers.GetUnauthenticatedClient()
	loginResult, err := loginAPI.Login(ctx, testUser.Email, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to login: %w", err)
	}

	if !loginResult.OTPRequired {
		return ctx, fmt.Errorf("OTP was not required")
	}

	// Create valid TOTP code using helper
	testOTP, err := helpers.CreateValidTOTP(ctx)
	if err != nil {
		return ctx, err
	}

	finalToken, err := loginAPI.ValidateOTP(ctx, loginResult.IntermediateJWT, testOTP)
	if err != nil {
		return ctx, fmt.Errorf("failed to validate OTP: %w", err)
	}

	if finalToken == "" {
		return ctx, fmt.Errorf("OTP validation did not return a token")
	}

	return helpers.SetJWTToken(ctx, finalToken), nil
}

func (s *AuthSteps) invalidOTPCodesAreRejected(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	loginAPI := helpers.GetUnauthenticatedClient()
	loginResult, err := loginAPI.Login(ctx, testUser.Email, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to login: %w", err)
	}

	if !loginResult.OTPRequired {
		return ctx, fmt.Errorf("OTP was not required")
	}

	// Try to validate with an invalid OTP code
	invalidOTP := "000000"
	_, err = loginAPI.ValidateOTP(ctx, loginResult.IntermediateJWT, invalidOTP)
	if err == nil {
		return ctx, fmt.Errorf("invalid OTP code was not rejected")
	}

	return ctx, nil
}

// 2FA disable step implementations

func (s *AuthSteps) theUserRequestsToDisable2FA(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	err = api.DisableOTP(ctx, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to disable 2FA: %w", err)
	}

	return ctx, nil
}

func (s *AuthSteps) twoFAIsSuccessfullyDisabled(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}

func (s *AuthSteps) theUserCanLoginWithoutOTPVerification(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	loginAPI := helpers.GetUnauthenticatedClient()
	loginResult, err := loginAPI.Login(ctx, testUser.Email, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to login: %w", err)
	}

	if loginResult.OTPRequired {
		return ctx, fmt.Errorf("OTP was still required after disabling 2FA")
	}

	if loginResult.Token == "" {
		return ctx, fmt.Errorf("login did not return a token")
	}

	return helpers.SetJWTToken(ctx, loginResult.Token), nil
}
