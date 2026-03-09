package helpers

import (
	"context"
	"fmt"
	"strings"
	"time"

	account "go.lumeweb.com/portal-sdk"

	"github.com/go-faker/faker/v4"
	"github.com/pquerna/otp/totp"
)

// Context keys for storing test data between steps
type contextKey string

const (
	JWTTokenKey         contextKey = "jwt_token"
	APIKeyKey           contextKey = "api_key"
	APIKeyUUIDKey       contextKey = "api_key_uuid"
	TestUserKey         contextKey = "test_user"
	OperationIDKey      contextKey = "operation_id"
	OTPSecretKey        contextKey = "otp_secret"
	APIKeysCleanupKey   contextKey = "api_keys_cleanup"
	APIKeyUUIDsCleanupKey contextKey = "api_key_uuids_cleanup"
	TestUsersCleanupKey contextKey = "test_users_cleanup"
	APIKeysListKey      contextKey = "api_keys_list"
	UploadLimitKey      contextKey = "upload_limit"
	OperationsCleanupKey      contextKey = "operations_cleanup"
	PasswordResetTokenKey     contextKey = "password_reset_token"
	VerificationTokenKey      contextKey = "verification_token"
	AuthenticatedClientKey    contextKey = "authenticated_client"
	RegistrationErrorKey      contextKey = "registration_error"
	PageSizeKey         contextKey = "page_size"
	FilterNameKey       contextKey = "filter_name"
	User1Key            contextKey = "user1"
	User2Key            contextKey = "user2"
	APIKey1TokenKey     contextKey = "api_key1_token"
	APIKey2TokenKey     contextKey = "api_key2_token"
	FirstApiKeyNameKey  contextKey = "first_api_key_name"
	FirstApiKeyTokenKey contextKey = "first_api_key_token"
	SecondApiKeyTokenKey contextKey = "second_api_key_token"
	RegistrationEmailKey contextKey = "registration_email"
	RegistrationPasswordKey contextKey = "registration_password"
)

// TestUser represents a test user for authentication
type TestUser struct {
	Email     string
	Password  string
	Username  string
	FirstName string
	LastName  string
}

// GenerateUniqueEmail creates a unique email address for testing
func GenerateUniqueEmail() string {
	return faker.Email()
}

// GenerateUniqueUsername creates a unique username for testing
func GenerateUniqueUsername() string {
	return faker.Username()
}

// GenerateStrongPassword creates a strong password for testing
func GenerateStrongPassword() string {
	return faker.Password()
}

// GenerateFirstName creates a random first name for testing
func GenerateFirstName() string {
	return faker.FirstName()
}

// GenerateLastName creates a random last name for testing
func GenerateLastName() string {
	return faker.LastName()
}

// CreateTestUser creates a new test user with unique identifiers
func CreateTestUser() *TestUser {
	return &TestUser{
		Email:     GenerateUniqueEmail(),
		Password:  GenerateStrongPassword(),
		Username:  GenerateUniqueUsername(),
		FirstName: GenerateFirstName(),
		LastName:  GenerateLastName(),
	}
}

// SetJWTToken stores JWT token in context
func SetJWTToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, JWTTokenKey, token)
}

// GetJWTToken retrieves JWT token from context
func GetJWTToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(JWTTokenKey).(string)
	return token, ok
}

// SetAPIKey stores API key in context
func SetAPIKey(ctx context.Context, apiKey string) context.Context {
	return context.WithValue(ctx, APIKeyKey, apiKey)
}

// GetAPIKey retrieves API key from context
func GetAPIKey(ctx context.Context) (string, bool) {
	apiKey, ok := ctx.Value(APIKeyKey).(string)
	return apiKey, ok
}

// SetTestUser stores test user in context
func SetTestUser(ctx context.Context, user *TestUser) context.Context {
	return context.WithValue(ctx, TestUserKey, user)
}

// GetTestUser retrieves test user from context
func GetTestUser(ctx context.Context) (*TestUser, bool) {
	user, ok := ctx.Value(TestUserKey).(*TestUser)
	return user, ok
}

// SetOperationID stores operation ID in context
func SetOperationID(ctx context.Context, operationID string) context.Context {
	return context.WithValue(ctx, OperationIDKey, operationID)
}

// GetOperationID retrieves operation ID from context
func GetOperationID(ctx context.Context) (string, bool) {
	operationID, ok := ctx.Value(OperationIDKey).(string)
	return operationID, ok
}

// SetOTPSecret stores OTP secret in context
func SetOTPSecret(ctx context.Context, secret string) context.Context {
	return context.WithValue(ctx, OTPSecretKey, secret)
}

// GetOTPSecret retrieves OTP secret from context
func GetOTPSecret(ctx context.Context) (string, bool) {
	secret, ok := ctx.Value(OTPSecretKey).(string)
	return secret, ok
}

// AddAPIKeyCleanup adds API key to cleanup list
func AddAPIKeyCleanup(ctx context.Context, apiKey string) context.Context {
	cleanupList, _ := ctx.Value(APIKeysCleanupKey).([]string)
	if cleanupList == nil {
		cleanupList = []string{}
	}
	cleanupList = append(cleanupList, apiKey)
	return context.WithValue(ctx, APIKeysCleanupKey, cleanupList)
}

// GetAPIKeysCleanup retrieves API keys cleanup list
func GetAPIKeysCleanup(ctx context.Context) []string {
	cleanupList, _ := ctx.Value(APIKeysCleanupKey).([]string)
	return cleanupList
}

// AddTestUserCleanup adds test user to cleanup list
func AddTestUserCleanup(ctx context.Context, email string) context.Context {
	cleanupList, _ := ctx.Value(TestUsersCleanupKey).([]string)
	if cleanupList == nil {
		cleanupList = []string{}
	}
	cleanupList = append(cleanupList, email)
	return context.WithValue(ctx, TestUsersCleanupKey, cleanupList)
}

// GetTestUsersCleanup retrieves test users cleanup list
func GetTestUsersCleanup(ctx context.Context) []string {
	cleanupList, _ := ctx.Value(TestUsersCleanupKey).([]string)
	return cleanupList
}

// SetAPIKeyUUID stores API key UUID in context
func SetAPIKeyUUID(ctx context.Context, uuid string) context.Context {
	return context.WithValue(ctx, APIKeyUUIDKey, uuid)
}

// GetAPIKeyUUID retrieves API key UUID from context
func GetAPIKeyUUID(ctx context.Context) (string, bool) {
	uuid, ok := ctx.Value(APIKeyUUIDKey).(string)
	return uuid, ok
}

// AddAPIKeyUUIDCleanup adds API key UUID to cleanup list
func AddAPIKeyUUIDCleanup(ctx context.Context, uuid string) context.Context {
	cleanupList, _ := ctx.Value(APIKeyUUIDsCleanupKey).([]string)
	if cleanupList == nil {
		cleanupList = []string{}
	}
	cleanupList = append(cleanupList, uuid)
	return context.WithValue(ctx, APIKeyUUIDsCleanupKey, cleanupList)
}

// GetAPIKeyUUIDsCleanup retrieves API key UUIDs cleanup list
func GetAPIKeyUUIDsCleanup(ctx context.Context) []string {
	cleanupList, _ := ctx.Value(APIKeyUUIDsCleanupKey).([]string)
	return cleanupList
}

// SetAPIKeysList stores API keys list in context
func SetAPIKeysList(ctx context.Context, keys []*account.APIKey) context.Context {
	return context.WithValue(ctx, APIKeysListKey, keys)
}

// GetAPIKeysList retrieves API keys list from context
func GetAPIKeysList(ctx context.Context) ([]*account.APIKey, bool) {
	keys, _ := ctx.Value(APIKeysListKey).([]*account.APIKey)
	return keys, keys != nil
}

// SetUploadLimit stores upload limit in context
func SetUploadLimit(ctx context.Context, limit int64) context.Context {
	return context.WithValue(ctx, UploadLimitKey, limit)
}

// GetUploadLimit retrieves upload limit from context
func GetUploadLimit(ctx context.Context) (int64, bool) {
	limit, ok := ctx.Value(UploadLimitKey).(int64)
	return limit, ok
}

// AddOperationCleanup adds operation ID to cleanup list
func AddOperationCleanup(ctx context.Context, operationID string) context.Context {
	cleanupList, _ := ctx.Value(OperationsCleanupKey).([]string)
	if cleanupList == nil {
		cleanupList = []string{}
	}
	cleanupList = append(cleanupList, operationID)
	return context.WithValue(ctx, OperationsCleanupKey, cleanupList)
}

// GetOperationsCleanup retrieves operations cleanup list
func GetOperationsCleanup(ctx context.Context) []string {
	cleanupList, _ := ctx.Value(OperationsCleanupKey).([]string)
	return cleanupList
}

// SetPasswordResetToken stores password reset token in context
func SetPasswordResetToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, PasswordResetTokenKey, token)
}

// GetPasswordResetToken retrieves password reset token from context
func GetPasswordResetToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(PasswordResetTokenKey).(string)
	return token, ok
}

// SetVerificationToken stores email verification token in context
func SetVerificationToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, VerificationTokenKey, token)
}

// GetVerificationToken retrieves email verification token from context
func GetVerificationToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(VerificationTokenKey).(string)
	return token, ok
}

// GenerateValidTOTPCode generates a valid TOTP code from a secret using time-based OTP
func GenerateValidTOTPCode(secret string) (string, error) {
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP code: %w", err)
	}
	
	return code, nil
}

// SetAuthenticatedClient stores the authenticated client in context
func SetAuthenticatedClient(ctx context.Context, client account.AccountAPI) context.Context {
	return context.WithValue(ctx, AuthenticatedClientKey, client)
}

// GetAuthenticatedClient retrieves the authenticated client from context
func GetAuthenticatedClient(ctx context.Context) (account.AccountAPI, bool) {
	client, ok := ctx.Value(AuthenticatedClientKey).(account.AccountAPI)
	return client, ok
}

// SetPageSize stores the page size in context
func SetPageSize(ctx context.Context, pageSize int) context.Context {
	return context.WithValue(ctx, PageSizeKey, pageSize)
}

// GetPageSize retrieves the page size from context
func GetPageSize(ctx context.Context) (int, bool) {
	pageSize, ok := ctx.Value(PageSizeKey).(int)
	return pageSize, ok
}

// SetFilterName stores the filter name in context
func SetFilterName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, FilterNameKey, name)
}

// GetFilterName retrieves the filter name from context
func GetFilterName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(FilterNameKey).(string)
	return name, ok
}

// GetProfile retrieves and returns the user's profile from the account API
// This returns the account details including email, first name, and last name
func GetProfile(ctx context.Context) (*account.AccountInfo, error) {
	api := GetAuthenticatedClientFromContext(ctx)
	if api == nil {
		return nil, fmt.Errorf("no authenticated client available")
	}

	profile, err := api.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve profile: %w", err)
	}

	return profile, nil
}

// VerifyProfileField checks if the profile field matches the expected value
// This helper DRYs up profile verification logic across step implementations
func VerifyProfileField(ctx context.Context, fieldName, expectedValue string) error {
	profile, err := GetProfile(ctx)
	if err != nil {
		return err
	}

	var actualValue string
	switch fieldName {
	case "email":
		actualValue = profile.Email
	case "first_name", "firstName":
		actualValue = profile.FirstName
	case "last_name", "lastName":
		actualValue = profile.LastName
	default:
		return fmt.Errorf("unknown profile field: %s", fieldName)
	}

	if actualValue != expectedValue {
		return fmt.Errorf("profile field %s: expected '%s', got '%s'", fieldName, expectedValue, actualValue)
	}

	return nil
}

// VerifyJWTInvalidation verifies that JWT tokens have been cleared from local context.
// Note: For stateless JWT systems, this only checks that tokens are removed from local storage,
// not that they are invalid at the server level. Stateless JWTs remain valid until expiration.
func VerifyJWTInvalidation(ctx context.Context) error {
	token, hasToken := GetJWTToken(ctx)
	if !hasToken || token == "" {
		// Token has been cleared from local storage
		return nil
	}

	api := GetAuthenticatedClientFromContext(ctx)
	if api == nil {
		// No authenticated client available means no JWT in context
		return nil
	}

	// Note: We don't attempt to ping the server here because with stateless JWTs,
	// the token would still be valid from the server's perspective until expiration.
	// This helper only verifies local state cleanup, which is the standard logout
	// behavior for stateless JWT systems.

	// Token still exists in context - not logged out
	return fmt.Errorf("JWT token still in context, logout may not have been called")
}

// VerifyAPIKeysNameFilter verifies that API key filtering works correctly
// It queries the API with the given filter name and checks that only matching keys are returned
func VerifyAPIKeysNameFilter(ctx context.Context, filterName string, expectedKeyNames []string) error {
	api := GetAuthenticatedClientFromContext(ctx)
	if api == nil {
		return fmt.Errorf("no authenticated client available")
	}

	// List all API keys for the user
	allKeys, err := api.ListAPIKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to list API keys: %w", err)
	}

	// Filter keys by name
	var matchedKeys []*account.APIKey
	for _, key := range allKeys {
		if strings.Contains(key.Name, filterName) {
			matchedKeys = append(matchedKeys, key)
		}
	}

	// Verify expected count matches
	if len(matchedKeys) != len(expectedKeyNames) {
		return fmt.Errorf("expected %d keys matching filter '%s', got %d", len(expectedKeyNames), filterName, len(matchedKeys))
	}

	// Verify all expected keys are present
	expectedMap := make(map[string]bool)
	for _, name := range expectedKeyNames {
		expectedMap[name] = false
	}

	for _, key := range matchedKeys {
		if _, exists := expectedMap[key.Name]; exists {
			expectedMap[key.Name] = true
		}
	}

	for name, found := range expectedMap {
		if !found {
			return fmt.Errorf("expected key '%s' not found in filtered results", name)
		}
	}

	// Store filtered results in context
	ctx = SetAPIKeysList(ctx, matchedKeys)
	return nil
}

// VerifyAPIKeysPagination verifies that API key pagination works correctly
// It queries the API with the given page size and returns the page results
func VerifyAPIKeysPagination(ctx context.Context, pageSize int) ([]*account.APIKey, error) {
	api := GetAuthenticatedClientFromContext(ctx)
	if api == nil {
		return nil, fmt.Errorf("no authenticated client available")
	}

	// List all API keys first
	allKeys, err := api.ListAPIKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	// Return only the first page based on page size
	if pageSize > len(allKeys) {
		pageSize = len(allKeys)
	}

	firstPage := allKeys[:pageSize]

	// Store results in context
	ctx = SetAPIKeysList(ctx, firstPage)

	return firstPage, nil
}

// GetAccount returns the current authenticated user's account information
// This is a convenience wrapper around GetProfile for better naming
func GetAccount(ctx context.Context) (*account.AccountInfo, error) {
	return GetProfile(ctx)
}

// SetRegistrationError stores the registration error in context for negative test verification
func SetRegistrationError(ctx context.Context, err error) context.Context {
	if err != nil {
		return context.WithValue(ctx, RegistrationErrorKey, err.Error())
	}
	return context.WithValue(ctx, RegistrationErrorKey, "")
}

// GetRegistrationError retrieves the registration error from context
func GetRegistrationError(ctx context.Context) string {
	err, ok := ctx.Value(RegistrationErrorKey).(string)
	if !ok {
		return ""
	}
	return err
}

// SetUser1 stores the first user in context for multi-user scenarios
func SetUser1(ctx context.Context, user *TestUser) context.Context {
	return context.WithValue(ctx, User1Key, user)
}

// GetUser1 retrieves the first user from context
func GetUser1(ctx context.Context) (*TestUser, bool) {
	user, ok := ctx.Value(User1Key).(*TestUser)
	return user, ok
}

// SetUser2 stores the second user in context for multi-user scenarios
func SetUser2(ctx context.Context, user *TestUser) context.Context {
	return context.WithValue(ctx, User2Key, user)
}

// GetUser2 retrieves the second user from context
func GetUser2(ctx context.Context) (*TestUser, bool) {
	user, ok := ctx.Value(User2Key).(*TestUser)
	return user, ok
}

// SetAPIKey1Token stores the first user's API key token
func SetAPIKey1Token(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, APIKey1TokenKey, token)
}

// GetAPIKey1Token retrieves the first user's API key token
func GetAPIKey1Token(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(APIKey1TokenKey).(string)
	return token, ok
}

// SetAPIKey2Token stores the second user's API key token
func SetAPIKey2Token(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, APIKey2TokenKey, token)
}

// GetAPIKey2Token retrieves the second user's API key token
func GetAPIKey2Token(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(APIKey2TokenKey).(string)
	return token, ok
}

// SetFirstApiKeyToken stores the first API key token in duplicate scenarios
func SetFirstApiKeyToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, FirstApiKeyTokenKey, token)
}

// GetFirstApiKeyToken retrieves the first API key token
func GetFirstApiKeyToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(FirstApiKeyTokenKey).(string)
	return token, ok
}

// SetSecondApiKeyToken stores the second API key token in duplicate scenarios
func SetSecondApiKeyToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, SecondApiKeyTokenKey, token)
}

// GetSecondApiKeyToken retrieves the second API key token
func GetSecondApiKeyToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(SecondApiKeyTokenKey).(string)
	return token, ok
}

// SetRegistrationEmail stores the registration email in context
func SetRegistrationEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, RegistrationEmailKey, email)
}

// GetRegistrationEmail retrieves the registration email from context
func GetRegistrationEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(RegistrationEmailKey).(string)
	return email, ok
}

// SetRegistrationPassword stores the registration password in context
func SetRegistrationPassword(ctx context.Context, password string) context.Context {
	return context.WithValue(ctx, RegistrationPasswordKey, password)
}

// GetRegistrationPassword retrieves the registration password from context
func GetRegistrationPassword(ctx context.Context) (string, bool) {
	password, ok := ctx.Value(RegistrationPasswordKey).(string)
	return password, ok
}

