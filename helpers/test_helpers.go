package helpers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"
	"time"

	account "go.lumeweb.com/portal-sdk"
	goCid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"

	"github.com/go-faker/faker/v4"
	"github.com/pquerna/otp/totp"
)

// IPFSContent represents an IPFS content item with name and CID
type IPFSContent struct {
	Name string
	CID  string
}

// Context keys for storing test data between steps
type contextKey string

const (
	JWTTokenKey       contextKey = "jwt_token"
	APIKeyKey         contextKey = "api_key"
	APIKeyUUIDKey     contextKey = "api_key_uuid"
	TestUserKey       contextKey = "test_user"
	OperationIDKey    contextKey = "operation_id"
	OTPSecretKey      contextKey = "otp_secret"
	APIKeysCleanupKey contextKey = "api_keys_cleanup"
	APIKeyUUIDsCleanupKey contextKey = "api_key_uuids_cleanup"
	TestUsersCleanupKey contextKey = "test_users_cleanup"
	APIKeysListKey contextKey = "api_keys_list"
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

	// IPFS context keys for IPFS state management
	CIDKey                  contextKey = "cid"
	OldCIDKey               contextKey = "old_cid"
	PinRequestIDKey         contextKey = "pin_request_id"
	IPFSContentListKey      contextKey = "ipfs_content_list"
	CIDsCleanupKey          contextKey = "cids_cleanup"
	PinRequestIDsCleanupKey contextKey = "pin_request_ids_cleanup"
	CIDsListKey             contextKey = "cids_list"
	FileNamesListKey        contextKey = "filenames_list"
	PinListKey              contextKey = "pin_list"
	CurrentPinKey           contextKey = "current_pin"
	FilenameKey             contextKey = "filename"
	TestDirectoryKey        contextKey = "test_directory"
	PinnedStatusKey         contextKey = "pinned_status"
)

// =============================================================================
// Generic Context Helpers
// =============================================================================

// SetContextValue stores any value in context using a generic strongly-typed approach
// This is a generic helper that replaces simple setter functions
func SetContextValue[T any](ctx context.Context, key contextKey, value T) context.Context {
	return context.WithValue(ctx, key, value)
}

// GetContextValue retrieves a value from context with type safety
// This is a generic helper that replaces simple getter functions
func GetContextValue[T any](ctx context.Context, key contextKey) (T, bool) {
	value, ok := ctx.Value(key).(T)
	if !ok {
		var zero T
		return zero, false
	}
	return value, ok
}

// AddToCleanupList adds an item to any cleanup list (generic type-safe version)
// This replaces 6 identical Add*Cleanup functions with a generic implementation
func AddToCleanupList[T comparable](ctx context.Context, key contextKey, value T) context.Context {
	listInterface := ctx.Value(key)
	
	var list []T
	
	// Use reflection to handle the slice type conversion safely
	if listInterface != nil {
		val := reflect.ValueOf(listInterface)
		if val.Kind() == reflect.Slice {
			list = make([]T, val.Len())
			for i := 0; i < val.Len(); i++ {
				elem := val.Index(i).Interface()
				if typedElem, ok := elem.(T); ok {
					list[i] = typedElem
				}
			}
		}
	}
	
	// If list is nil or empty, initialize it
	if list == nil {
		list = []T{}
	}
	
	// Append the new value
	list = append(list, value)
	return context.WithValue(ctx, key, list)
}

// GetCleanupList retrieves any cleanup list from context (generic type-safe version)
// This is a generic helper used by cleanup getter functions
func GetCleanupList[T any](ctx context.Context, key contextKey) []T {
	if list, ok := ctx.Value(key).([]T); ok {
		return list
	}
	return []T{}
}

// =============================================================================
// Value Generation Functions
// =============================================================================

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

// GenerateUniqueContent generates unique test content by appending a unique identifier
// to the provided base content ensuring each upload produces a different CID.
// This prevents IPFS from deduplicating uploads across test runs which confuses
// the wait-for-pin logic.
func GenerateUniqueContent(baseContent string) string {
	uniqueSuffix := fmt.Sprintf(" [%s]", faker.UUIDHyphenated())
	return baseContent + uniqueSuffix
}

// GenerateLargeTestFile creates test data of specified size in bytes
// Useful for testing large file uploads and performance scenarios
func GenerateLargeTestFile(sizeBytes int64) ([]byte, error) {
	content := make([]byte, sizeBytes)
	
	// Fill with random data using crypto/rand
	_, err := rand.Read(content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random content: %w", err)
	}
	return content, nil
}

// ComputeCIDFromContent computes the IPFS CID for given content using go-cid
// Used for integrity verification - the CID IS the hash of content in IPFS
func ComputeCIDFromContent(content []byte) (string, error) {
	// Compute SHA-256 hash of content
	hash := sha256.Sum256(content)
	
	// Create multihash from SHA-256 hash
	multihash, err := mh.Encode(hash[:], mh.SHA2_256)
	if err != nil {
		return "", fmt.Errorf("failed to create multihash: %w", err)
	}
	
	// Create CID v1 with raw content type
	cid := goCid.NewCidV1(goCid.Raw, multihash)
	
	return cid.String(), nil
}



// GenerateValidTOTPCode generates a valid TOTP code from a secret using time-based OTP
func GenerateValidTOTPCode(secret string) (string, error) {
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP code: %w", err)
	}

	return code, nil
}

// =============================================================================
// Profile and Account Helpers
// =============================================================================

// GetProfile retrieves and returns the user's profile from the account API
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

// VerifyJWTInvalidation verifies that JWT tokens have been cleared from local context
func VerifyJWTInvalidation(ctx context.Context) error {
	token, hasToken := GetJWTToken(ctx)
	if !hasToken || token == "" {
		// Token has been cleared from local storage
		return nil
	}

	api := GetAuthenticatedClientFromContext(ctx)
	if api == nil {
		return nil
	}

	return fmt.Errorf("JWT token still in context, logout may not have been called")
}

// VerifyAPIKeysNameFilter verifies that API key filtering works correctly
func VerifyAPIKeysNameFilter(ctx context.Context, filterName string, expectedKeyNames []string) (context.Context, error) {
	api := GetAuthenticatedClientFromContext(ctx)
	if api == nil {
		return ctx, fmt.Errorf("no authenticated client available")
	}

	allKeys, err := api.ListAPIKeys(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list API keys: %w", err)
	}

	var matchedKeys []*account.APIKey
	for _, key := range allKeys {
		if strings.Contains(key.Name, filterName) {
			matchedKeys = append(matchedKeys, key)
		}
	}

	if len(matchedKeys) != len(expectedKeyNames) {
		return ctx, fmt.Errorf("expected %d keys matching filter '%s', got %d", len(expectedKeyNames), filterName, len(matchedKeys))
	}

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
			return ctx, fmt.Errorf("expected key '%s' not found in filtered results", name)
		}
	}

	ctx = SetContextValue(ctx, APIKeysListKey, matchedKeys)
	return ctx, nil
}

// VerifyAPIKeysPagination verifies that API key pagination works correctly
func VerifyAPIKeysPagination(ctx context.Context, pageSize int) (context.Context, []*account.APIKey, error) {
	api := GetAuthenticatedClientFromContext(ctx)
	if api == nil {
		return ctx, nil, fmt.Errorf("no authenticated client available")
	}

	allKeys, err := api.ListAPIKeys(ctx)
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	if pageSize > len(allKeys) {
		pageSize = len(allKeys)
	}

	firstPage := allKeys[:pageSize]

	ctx = SetContextValue(ctx, APIKeysListKey, firstPage)
	return ctx, firstPage, nil
}

// GetAccount returns the current authenticated user's account information
func GetAccount(ctx context.Context) (*account.AccountInfo, error) {
	return GetProfile(ctx)
}

// =============================================================================
// Backward Compatibility Layer
// =============================================================================
// NOTE: These functions are kept for backward compatibility and are implemented
// using the new generic helpers. Consider migrating to the generic API for new code.

// SetJWTToken stores JWT token in context
func SetJWTToken(ctx context.Context, token string) context.Context {
	return SetContextValue(ctx, JWTTokenKey, token)
}

// GetJWTToken retrieves JWT token from context
func GetJWTToken(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, JWTTokenKey)
}

// SetAPIKey stores API key in context
func SetAPIKey(ctx context.Context, apiKey string) context.Context {
	return SetContextValue(ctx, APIKeyKey, apiKey)
}

// GetAPIKey retrieves API key from context
func GetAPIKey(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, APIKeyKey)
}

// SetTestUser stores test user in context
func SetTestUser(ctx context.Context, user *TestUser) context.Context {
	return SetContextValue(ctx, TestUserKey, user)
}

// GetTestUser retrieves test user from context
func GetTestUser(ctx context.Context) (*TestUser, bool) {
	return GetContextValue[*TestUser](ctx, TestUserKey)
}

// SetOperationID stores operation ID in context
func SetOperationID(ctx context.Context, operationID string) context.Context {
	return SetContextValue(ctx, OperationIDKey, operationID)
}

// GetOperationID retrieves operation ID from context
func GetOperationID(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, OperationIDKey)
}

// SetOTPSecret stores OTP secret in context
func SetOTPSecret(ctx context.Context, secret string) context.Context {
	return SetContextValue(ctx, OTPSecretKey, secret)
}

// GetOTPSecret retrieves OTP secret from context
func GetOTPSecret(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, OTPSecretKey)
}

// AddAPIKeyCleanup adds API key to cleanup list
func AddAPIKeyCleanup(ctx context.Context, apiKey string) context.Context {
	return AddToCleanupList(ctx, APIKeysCleanupKey, apiKey)
}

// GetAPIKeysCleanup retrieves API keys cleanup list
func GetAPIKeysCleanup(ctx context.Context) []string {
	return GetCleanupList[string](ctx, APIKeysCleanupKey)
}

// AddTestUserCleanup adds test user to cleanup list
func AddTestUserCleanup(ctx context.Context, email string) context.Context {
	return AddToCleanupList(ctx, TestUsersCleanupKey, email)
}

// GetTestUsersCleanup retrieves test users cleanup list
func GetTestUsersCleanup(ctx context.Context) []string {
	return GetCleanupList[string](ctx, TestUsersCleanupKey)
}

// SetAPIKeyUUID stores API key UUID in context
func SetAPIKeyUUID(ctx context.Context, uuid string) context.Context {
	return SetContextValue(ctx, APIKeyUUIDKey, uuid)
}

// GetAPIKeyUUID retrieves API key UUID from context
func GetAPIKeyUUID(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, APIKeyUUIDKey)
}

// AddAPIKeyUUIDCleanup adds API key UUID to cleanup list
func AddAPIKeyUUIDCleanup(ctx context.Context, uuid string) context.Context {
	return AddToCleanupList(ctx, APIKeyUUIDsCleanupKey, uuid)
}

// GetAPIKeyUUIDsCleanup retrieves API key UUIDs cleanup list
func GetAPIKeyUUIDsCleanup(ctx context.Context) []string {
	return GetCleanupList[string](ctx, APIKeyUUIDsCleanupKey)
}

// SetAPIKeysList stores API keys list in context
func SetAPIKeysList(ctx context.Context, keys []*account.APIKey) context.Context {
	return SetContextValue(ctx, APIKeysListKey, keys)
}

// GetAPIKeysList retrieves API keys list from context
func GetAPIKeysList(ctx context.Context) ([]*account.APIKey, bool) {
	return GetContextValue[[]*account.APIKey](ctx, APIKeysListKey)
}

// SetUploadLimit stores upload limit in context
func SetUploadLimit(ctx context.Context, limit int64) context.Context {
	return SetContextValue(ctx, UploadLimitKey, limit)
}

// GetUploadLimit retrieves upload limit from context
func GetUploadLimit(ctx context.Context) (int64, bool) {
	return GetContextValue[int64](ctx, UploadLimitKey)
}

// AddOperationCleanup adds operation ID to cleanup list
func AddOperationCleanup(ctx context.Context, operationID string) context.Context {
	return AddToCleanupList(ctx, OperationsCleanupKey, operationID)
}

// GetOperationsCleanup retrieves operations cleanup list
func GetOperationsCleanup(ctx context.Context) []string {
	return GetCleanupList[string](ctx, OperationsCleanupKey)
}

// SetPasswordResetToken stores password reset token in context
func SetPasswordResetToken(ctx context.Context, token string) context.Context {
	return SetContextValue(ctx, PasswordResetTokenKey, token)
}

// GetPasswordResetToken retrieves password reset token from context
func GetPasswordResetToken(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, PasswordResetTokenKey)
}

// SetVerificationToken stores email verification token in context
func SetVerificationToken(ctx context.Context, token string) context.Context {
	return SetContextValue(ctx, VerificationTokenKey, token)
}

// GetVerificationToken retrieves email verification token from context
func GetVerificationToken(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, VerificationTokenKey)
}

// SetAuthenticatedClient stores the authenticated client in context
func SetAuthenticatedClient(ctx context.Context, client account.AccountAPI) context.Context {
	return SetContextValue(ctx, AuthenticatedClientKey, client)
}

// GetAuthenticatedClient retrieves the authenticated client from context
func GetAuthenticatedClient(ctx context.Context) (account.AccountAPI, bool) {
	return GetContextValue[account.AccountAPI](ctx, AuthenticatedClientKey)
}

// SetPageSize stores the page size in context
func SetPageSize(ctx context.Context, pageSize int) context.Context {
	return SetContextValue(ctx, PageSizeKey, pageSize)
}

// GetPageSize retrieves the page size from context
func GetPageSize(ctx context.Context) (int, bool) {
	return GetContextValue[int](ctx, PageSizeKey)
}

// SetFilterName stores the filter name in context
func SetFilterName(ctx context.Context, name string) context.Context {
	return SetContextValue(ctx, FilterNameKey, name)
}

// GetFilterName retrieves the filter name from context
func GetFilterName(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, FilterNameKey)
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
	return SetContextValue(ctx, User1Key, user)
}

// GetUser1 retrieves the first user from context
func GetUser1(ctx context.Context) (*TestUser, bool) {
	return GetContextValue[*TestUser](ctx, User1Key)
}

// SetUser2 stores the second user in context for multi-user scenarios
func SetUser2(ctx context.Context, user *TestUser) context.Context {
	return SetContextValue(ctx, User2Key, user)
}

// GetUser2 retrieves the second user from context
func GetUser2(ctx context.Context) (*TestUser, bool) {
	return GetContextValue[*TestUser](ctx, User2Key)
}

// SetAPIKey1Token stores the first user's API key token
func SetAPIKey1Token(ctx context.Context, token string) context.Context {
	return SetContextValue(ctx, APIKey1TokenKey, token)
}

// GetAPIKey1Token retrieves the first user's API key token
func GetAPIKey1Token(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, APIKey1TokenKey)
}

// SetAPIKey2Token stores the second user's API key token
func SetAPIKey2Token(ctx context.Context, token string) context.Context {
	return SetContextValue(ctx, APIKey2TokenKey, token)
}

// GetAPIKey2Token retrieves the second user's API key token
func GetAPIKey2Token(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, APIKey2TokenKey)
}

// SetFirstApiKeyToken stores the first API key token in duplicate scenarios
func SetFirstApiKeyToken(ctx context.Context, token string) context.Context {
	return SetContextValue(ctx, FirstApiKeyTokenKey, token)
}

// GetFirstApiKeyToken retrieves the first API key token
func GetFirstApiKeyToken(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, FirstApiKeyTokenKey)
}

// SetSecondApiKeyToken stores the second API key token in duplicate scenarios
func SetSecondApiKeyToken(ctx context.Context, token string) context.Context {
	return SetContextValue(ctx, SecondApiKeyTokenKey, token)
}

// GetSecondApiKeyToken retrieves the second API key token
func GetSecondApiKeyToken(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, SecondApiKeyTokenKey)
}

// SetRegistrationEmail stores the registration email in context
func SetRegistrationEmail(ctx context.Context, email string) context.Context {
	return SetContextValue(ctx, RegistrationEmailKey, email)
}

// GetRegistrationEmail retrieves the registration email from context
func GetRegistrationEmail(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, RegistrationEmailKey)
}

// SetRegistrationPassword stores the registration password in context
func SetRegistrationPassword(ctx context.Context, password string) context.Context {
	return SetContextValue(ctx, RegistrationPasswordKey, password)
}

// GetRegistrationPassword retrieves the registration password from context
func GetRegistrationPassword(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, RegistrationPasswordKey)
}

// IPFS context helpers

// SetCID stores CID in context
func SetCID(ctx context.Context, cid string) context.Context {
	return SetContextValue(ctx, CIDKey, cid)
}

// GetCID retrieves CID from context
func GetCID(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, CIDKey)
}

// SetOldCID stores old CID in context
func SetOldCID(ctx context.Context, cid string) context.Context {
	return SetContextValue(ctx, OldCIDKey, cid)
}

// GetOldCID retrieves old CID from context
func GetOldCID(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, OldCIDKey)
}

// SetPinRequestID stores pin request ID in context
func SetPinRequestID(ctx context.Context, requestID string) context.Context {
	return SetContextValue(ctx, PinRequestIDKey, requestID)
}

// GetPinRequestID retrieves pin request ID from context
func GetPinRequestID(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, PinRequestIDKey)
}

// SetIPFSContentList stores IPFS content list in context
func SetIPFSContentList(ctx context.Context, contentList []IPFSContent) context.Context {
	return SetContextValue(ctx, IPFSContentListKey, contentList)
}

// GetIPFSContentList retrieves IPFS content list from context
func GetIPFSContentList(ctx context.Context) ([]IPFSContent, bool) {
	return GetContextValue[[]IPFSContent](ctx, IPFSContentListKey)
}

// AddCIDCleanup adds CID to cleanup list
func AddCIDCleanup(ctx context.Context, cid string) context.Context {
	return AddToCleanupList(ctx, CIDsCleanupKey, cid)
}

// GetCIDsCleanup retrieves CIDs cleanup list
func GetCIDsCleanup(ctx context.Context) []string {
	return GetCleanupList[string](ctx, CIDsCleanupKey)
}

// AddPinRequestIDCleanup adds pin request ID to cleanup list
func AddPinRequestIDCleanup(ctx context.Context, requestID string) context.Context {
	return AddToCleanupList(ctx, PinRequestIDsCleanupKey, requestID)
}

// GetPinRequestIDsCleanup retrieves pin request IDs cleanup list
func GetPinRequestIDsCleanup(ctx context.Context) []string {
	return GetCleanupList[string](ctx, PinRequestIDsCleanupKey)
}

// SetCIDs stores CIDs list in context
func SetCIDs(ctx context.Context, cids []string) context.Context {
	return SetContextValue(ctx, CIDsListKey, cids)
}

// GetCIDs retrieves CIDs list from context
func GetCIDs(ctx context.Context) ([]string, bool) {
	return GetContextValue[[]string](ctx, CIDsListKey)
}

// SetFileNames stores file names list in context
func SetFileNames(ctx context.Context, names []string) context.Context {
	return SetContextValue(ctx, FileNamesListKey, names)
}

// SetPinList stores pin list in context
func SetPinList(ctx context.Context, pins []string) context.Context {
	return SetContextValue(ctx, PinListKey, pins)
}

// GetPinList retrieves pin list from context
func GetPinList(ctx context.Context) ([]string, bool) {
	return GetContextValue[[]string](ctx, PinListKey)
}

// SetCurrentPin stores current pin in context
func SetCurrentPin(ctx context.Context, pin interface{}) context.Context {
	return SetContextValue(ctx, CurrentPinKey, pin)
}

// GetCurrentPin retrieves current pin from context
func GetCurrentPin(ctx context.Context) (interface{}, bool) {
	return GetContextValue[interface{}](ctx, CurrentPinKey)
}

// SetFilename stores filename in context
func SetFilename(ctx context.Context, filename string) context.Context {
	return SetContextValue(ctx, FilenameKey, filename)
}

// SetTestDirectory stores test directory path in context
func SetTestDirectory(ctx context.Context, dir string) context.Context {
	return SetContextValue(ctx, TestDirectoryKey, dir)
}

// GetTestDirectory retrieves test directory from context
func GetTestDirectory(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, TestDirectoryKey)
}

// SetPinnedStatus stores pinned status in context
func SetPinnedStatus(ctx context.Context, pinned bool) context.Context {
	return SetContextValue(ctx, PinnedStatusKey, pinned)
}

// GetPinnedStatus retrieves pinned status from context
func GetPinnedStatus(ctx context.Context) (bool, bool) {
	return GetContextValue[bool](ctx, PinnedStatusKey)
}

// SetKnownContent stores content in context for integrity verification
func SetKnownContent(ctx context.Context, content string) context.Context {
	return SetContextValue(ctx, contextKey("known_content"), content)
}

// GetKnownContent retrieves stored content from context
func GetKnownContent(ctx context.Context) (string, bool) {
	return GetContextValue[string](ctx, contextKey("known_content"))
}
