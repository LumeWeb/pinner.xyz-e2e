package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/cucumber/godog"
	account "go.lumeweb.com/portal-sdk"
)

const (
	// TagNoAuthReset preserves authentication state across scenarios when applied
	TagNoAuthReset = "@noAuthReset"
)

// StepsCommon provides shared step implementations and hook registration for all test suites

// GetEnv retrieves an environment variable or returns default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetPortalName returns the portal name from environment
func GetPortalName() string {
	return GetEnv("PORTAL__CORE__PORTAL_NAME", "test-portal")
}

// GetPortalURL returns the portal base URL from environment
func GetPortalURL() string {
	domain := GetEnv("PORTAL__CORE__DOMAIN", "localhost")
	port := GetEnv("PORTAL__CORE__PORT", "8080")
	return fmt.Sprintf("http://%s:%s", domain, port)
}

// GetMailDevURL returns the MailDev URL from environment
func GetMailDevURL() string {
	return GetEnv("MAILDEV_URL", "http://localhost:1080")
}

// GetMySQLHost returns the MySQL host from environment
func GetMySQLHost() string {
	return GetEnv("PORTAL__CORE__DB__HOST", "127.0.0.1")
}

// GetMySQLPort returns the MySQL port from environment
func GetMySQLPort() string {
	return GetEnv("PORTAL__CORE__DB__PORT", "3306")
}

// StepsCommon provides shared step implementations and hook registration for all test suites

// RegisterCommonHooks registers shared before/after scenario hooks for cleanup tracking
func RegisterCommonHooks(ctx *godog.ScenarioContext) {
	ctx.Before(beforeScenarioSetup)
	ctx.After(afterScenarioCleanup)
}

// beforeScenarioSetup initializes cleanup tracking for each scenario
func beforeScenarioSetup(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	// Initialize empty cleanup lists for this scenario
	ctx = context.WithValue(ctx, APIKeyUUIDsCleanupKey, []string{})
	ctx = context.WithValue(ctx, TestUsersCleanupKey, []string{})
	ctx = context.WithValue(ctx, OperationsCleanupKey, []string{})
	
	// Clear authenticated client to ensure each scenario starts with fresh auth state,
	// unless the scenario is tagged with @noAuthReset to preserve auth across scenarios
	// NOTE: JWT token is preserved in context so cleanup can recreate authenticated client
	hasNoAuthReset := false
	for _, tag := range sc.Tags {
		if tag.Name == TagNoAuthReset {
			hasNoAuthReset = true
			break
		}
	}
	
	if !hasNoAuthReset {
		ctx = context.WithValue(ctx, AuthenticatedClientKey, nil)
	}
	
	return ctx, nil
}

// afterScenarioCleanup performs cleanup after each scenario
func afterScenarioCleanup(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	// Clean up API keys using their UUIDs
	apiKeysToDelete := GetAPIKeyUUIDsCleanup(ctx)
	for _, uuid := range apiKeysToDelete {
		if deleteErr := DeleteAPIKeyGracefully(ctx, uuid); deleteErr != nil {
			fmt.Printf("Warning: failed to delete API key %s: %v\n", uuid, deleteErr)
		}
	}

	// Clean up test users
	testUsersToDelete := GetTestUsersCleanup(ctx)
	for _, email := range testUsersToDelete {
		if deleteErr := DeleteTestUserGracefully(ctx, email); deleteErr != nil {
			fmt.Printf("Warning: failed to delete test user %s: %v\n", email, deleteErr)
		}
	}

	// Clean up operations (if needed in the future)
	operationIDs := GetOperationsCleanup(ctx)
	for _, idStr := range operationIDs {
		fmt.Printf("Info: operation %s cleanup tracking\n", idStr)
	}

	// Return the original scenario error to preserve test failures
	return ctx, err
}

// GetPortalHost returns the vhost hostname for the portal API (for Host header)
func GetPortalHost() string {
	domain := GetEnv("PORTAL__CORE__DOMAIN", "")
	port := GetEnv("PORTAL__CORE__PORT", "8080")
	return fmt.Sprintf("account.%s:%s", domain, port)
}

// GetPortalTarget returns the target address for the portal API (actual connection target)
// Returns just host:port without protocol for WithHostOverride (e.g., "localhost:8080")
func GetPortalTarget() string {
	domain := GetEnv("PORTAL__CORE__DOMAIN", "")
	port := GetEnv("PORTAL__CORE__PORT", "8080")
	return fmt.Sprintf("%s:%s", domain, port)
}

// GetPortalEndpoint returns the full endpoint URL for the portal API
func GetPortalEndpoint() string {
	secure := GetEnv("PORTAL__CORE__SECURE", "false")
	domain := GetEnv("PORTAL__CORE__DOMAIN", "")
	port := GetEnv("PORTAL__CORE__PORT", "8080")
	
	protocol := "http"
	if secure == "true" || secure == "True" {
		protocol = "https"
	}
	
	return fmt.Sprintf("%s://account.%s:%s", protocol, domain, port)
}

// GetPortalServer returns the server hostname for SDK endpoint configuration
func GetPortalServer() string {
	secure := GetEnv("PORTAL__CORE__SECURE", "false")
	domain := GetEnv("PORTAL__CORE__DOMAIN", "")
	
	protocol := "http"
	if secure == "true" || secure == "True" {
		protocol = "https"
	}
	
	return fmt.Sprintf("%s://account.%s", protocol, domain)
}

// GetUnauthenticatedClient returns an API client without authentication
func GetUnauthenticatedClient() account.AccountAPI {
	return account.NewClient(
		account.WithEndpoint(GetPortalEndpoint()),
		account.WithHostOverride(GetPortalHost(), GetPortalTarget()),
	)
}

// CreateAuthenticatedClient creates an authenticated client with the given JWT token.
// This helper DRYs up the client creation logic used across multiple functions.
func CreateAuthenticatedClient(token string) account.AccountAPI {
	if token == "" {
		panic("CreateAuthenticatedClient called with empty token")
	}
	return account.NewClient(
		account.WithJWT(token),
		account.WithEndpoint(GetPortalEndpoint()),
		account.WithHostOverride(GetPortalHost(), GetPortalTarget()),
	)
}

// RequireAuthenticatedClient retrieves the authenticated client from context or returns an error.
// This DRYs up common error checking across step definitions.
func RequireAuthenticatedClient(ctx context.Context) (account.AccountAPI, error) {
	api := GetAuthenticatedClientFromContext(ctx)
	if api == nil {
		return nil, fmt.Errorf("no authenticated client available")
	}
	return api, nil
}

// RequireTestUser retrieves the test user from context or returns an error.
// This DRYs up common error checking across step definitions.
func RequireTestUser(ctx context.Context) (*TestUser, error) {
	testUser, ok := GetTestUser(ctx)
	if !ok || testUser == nil {
		return nil, fmt.Errorf("no test user available")
	}
	return testUser, nil
}

// createNewAuthenticatedClient creates a new authenticated client with the given JWT token.
// This is a legacy function kept for backward compatibility.
// Prefer using the client stored in context via GetAuthenticatedClientFromContext.
func createNewAuthenticatedClient(token string) account.AccountAPI {
	return CreateAuthenticatedClient(token)
}

// GetAuthenticatedClientFromContext retrieves the shared authenticated client from context
func GetAuthenticatedClientFromContext(ctx context.Context) account.AccountAPI {
	client, ok := GetAuthenticatedClient(ctx)
	if ok && client != nil {
		return client
	}
	
	// Fallback: create new client with JWT token
	token, hasToken := GetJWTToken(ctx)
	if !hasToken || token == "" {
		return nil
	}
	
	return CreateAuthenticatedClient(token)
}

// EnsureAuthenticatedURLHost returns consistent vhost routing configuration for account operations
// This consolidates the portal endpoint configuration pattern
func EnsureAuthenticatedURLHost() (host, endpoint string) {
	host = GetPortalHost()
	endpoint = GetPortalEndpoint()
	return
}

// GetHTTPClient returns an HTTP client for the test suite
func GetHTTPClient(token ...string) *HTTPClient {
	return NewHTTPClient(GetPortalURL(), token...)
}

// RegisterTestUser registers a new test user and stores them in context
func RegisterTestUser(ctx context.Context) (context.Context, error) {
	api := GetUnauthenticatedClient()
	testUser := CreateTestUser()

	err := api.Register(ctx, testUser.Email, testUser.FirstName, testUser.LastName, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to register user: %w", err)
	}

	ctx = SetTestUser(ctx, testUser)
	ctx = AddTestUserCleanup(ctx, testUser.Email)
	return ctx, nil
}

// LoginTestUser authenticates a test user and stores JWT token in context
func LoginTestUser(ctx context.Context) (context.Context, error) {
	testUser, ok := GetTestUser(ctx)
	if !ok {
		return ctx, fmt.Errorf("no test user available")
	}

	// First login without JWT to get the token
	api := GetUnauthenticatedClient()
	loginResult, err := api.Login(ctx, testUser.Email, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to login: %w", err)
	}

	if loginResult == nil || loginResult.Token == "" {
		return ctx, fmt.Errorf("login did not return a token")
	}

	// If 2FA is enabled, we need to provide OTP verification to get the final token
	if loginResult.OTPRequired {
		otpSecret, hasSecret := GetOTPSecret(ctx)
		if !hasSecret {
			return ctx, fmt.Errorf("OTP required but no OTP secret available in context")
		}
		
		// Generate a valid TOTP code
		otpCode, err := GenerateValidTOTPCode(otpSecret)
		if err != nil {
			return ctx, fmt.Errorf("failed to generate TOTP code: %w", err)
		}
		
		// Verify OTP to get the final authenticated token
		loginAPI := GetUnauthenticatedClient()
		finalToken, err := loginAPI.ValidateOTP(ctx, loginResult.Token, otpCode)
		if err != nil {
			return ctx, fmt.Errorf("failed to validate OTP: %w", err)
		}
		
		if finalToken == "" {
			return ctx, fmt.Errorf("OTP verification did not return a token")
		}
		
		// Use the final authenticated token
		loginResult.Token = finalToken
	}

	// Create authenticated client with JWT token
	// The portal accepts JWT tokens via Authorization header
	authClient := CreateAuthenticatedClient(loginResult.Token)
	ctx = SetJWTToken(ctx, loginResult.Token)
	ctx = SetAuthenticatedClient(ctx, authClient)
	return ctx, nil
}

// RegisterAndLoginTestUser combines registration and login in one call
func RegisterAndLoginTestUser(ctx context.Context) (context.Context, error) {
	ctx, err := RegisterTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	return LoginTestUser(ctx)
}


// isAPIKeyNotFoundError checks if the error is a 404 "record not found" error
// which indicates the API key was already deleted or never existed
func isAPIKeyNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "status 404") && strings.Contains(errStr, "record not found")
}

// isAccountConflictError checks if the error is a 409 conflict error
// which indicates an account deletion is already in progress
func isAccountConflictError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "status 409")
}

// DeleteAPIKeyGracefully attempts to delete an API key and ignores 404 errors
func DeleteAPIKeyGracefully(ctx context.Context, keyUUID string) error {
	api := GetAuthenticatedClientFromContext(ctx)
	
	// If no authenticated client, skip cleanup - no API keys were created in this scenario
	if api == nil {
		return nil
	}

	if err := api.DeleteAPIKey(ctx, keyUUID); err != nil {
		// 404 error means the key doesn't exist, which is acceptable for cleanup
		if isAPIKeyNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	return nil
}

// DeleteTestUserGracefully attempts to delete a test user and ignores 409 conflicts
func DeleteTestUserGracefully(ctx context.Context, email string) error {
	api := GetAuthenticatedClientFromContext(ctx)
	
	// If no authenticated client, skip cleanup - no test users were created in this scenario
	if api == nil {
		return nil
	}

	if err := api.DeleteAccount(ctx); err != nil {
		// 409 conflict means deletion is already in progress, which is acceptable for cleanup
		if isAccountConflictError(err) {
			return nil
		}
		return fmt.Errorf("failed to delete test user: %w", err)
	}

	return nil
}

// CleanupAPIKeys deletes all API keys created during the scenario
func CleanupAPIKeys(ctx context.Context) {
	apiKeyUUIDs := GetAPIKeyUUIDsCleanup(ctx)
	if len(apiKeyUUIDs) == 0 {
		return
	}

	api := GetAuthenticatedClientFromContext(ctx)
	if api == nil {
		return
	}

	for _, uuid := range apiKeyUUIDs {
		if deleteErr := api.DeleteAPIKey(ctx, uuid); deleteErr != nil {
			fmt.Printf("Warning: failed to delete API key %s: %v\n", uuid, deleteErr)
		}
	}
}

// CleanupTestUsers deletes all test users created during the scenario
func CleanupTestUsers(ctx context.Context) {
	testUsers := GetTestUsersCleanup(ctx)
	if len(testUsers) == 0 {
		return
	}

	api := GetAuthenticatedClientFromContext(ctx)
	if api == nil {
		for _, email := range testUsers {
			fmt.Printf("Info: test user %s cleanup pending (no client available)\n", email)
		}
		return
	}

	for _, email := range testUsers {
		if deleteErr := api.DeleteAccount(ctx); deleteErr != nil {
			fmt.Printf("Warning: failed to delete test user %s: %v\n", email, deleteErr)
		}
	}
}

// HTTPRequestOptions holds options for HTTP requests
type HTTPRequestOptions struct {
	Headers map[string]string
	Body    []byte
	Query   map[string]string
}

// HTTPClient represents an HTTP client for making requests
type HTTPClient struct {
	client    *http.Client
	baseURL   string
	authToken string
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL string, token ...string) *HTTPClient {
	c := &HTTPClient{
		client:  &http.Client{},
		baseURL: baseURL,
	}
	if len(token) > 0 && token[0] != "" {
		c.authToken = token[0]
	}
	return c
}

// GET performs a GET request
func (c *HTTPClient) GET(path string, opts *HTTPRequestOptions) (*http.Response, error) {
	return c.request(http.MethodGet, path, opts)
}

// POST performs a POST request
func (c *HTTPClient) POST(path string, opts *HTTPRequestOptions) (*http.Response, error) {
	return c.request(http.MethodPost, path, opts)
}

// PUT performs a PUT request
func (c *HTTPClient) PUT(path string, opts *HTTPRequestOptions) (*http.Response, error) {
	return c.request(http.MethodPut, path, opts)
}

// DELETE performs a DELETE request
func (c *HTTPClient) DELETE(path string, opts *HTTPRequestOptions) (*http.Response, error) {
	return c.request(http.MethodDelete, path, opts)
}

// PATCH performs a PATCH request
func (c *HTTPClient) PATCH(path string, opts *HTTPRequestOptions) (*http.Response, error) {
	return c.request(http.MethodPatch, path, opts)
}

// request performs an HTTP request with the given method
func (c *HTTPClient) request(method, path string, opts *HTTPRequestOptions) (*http.Response, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if opts != nil && opts.Body != nil {
		bodyReader = bytes.NewReader(opts.Body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header if token is available
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	// Set headers
	if opts != nil && opts.Headers != nil {
		for key, value := range opts.Headers {
			req.Header.Set(key, value)
		}
	}

	// Set default content type for requests with body
	if opts != nil && opts.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add query parameters
	if opts != nil && opts.Query != nil {
		q := req.URL.Query()
		for key, value := range opts.Query {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	return resp, nil
}

// AssertStatusCode asserts that the response has the expected status code
func AssertStatusCode(resp *http.Response, expected int) error {
	if resp.StatusCode != expected {
		return fmt.Errorf("expected status code %d, got %d", expected, resp.StatusCode)
	}
	return nil
}

// assertJSONValue asserts a JSON value matches expected
func assertJSONValue(value any, expectedValue, fieldOrPath string) error {
	valueStr := fmt.Sprintf("%v", value)
	if valueStr != expectedValue {
		return fmt.Errorf("expected %s=%s, got %s", fieldOrPath, expectedValue, valueStr)
	}
	return nil
}

// AssertJSONField asserts that a JSON response contains a field with the expected value
func AssertJSONField(body []byte, field, expectedValue string) error {
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	value, ok := data[field]
	if !ok {
		return fmt.Errorf("field %s not found in response", field)
	}

	return assertJSONValue(value, expectedValue, field)
}

// AssertJSONPath asserts that a JSON response contains a value at the given path
func AssertJSONPath(body []byte, path, expectedValue string) error {
	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	value, err := getJSONPath(data, path)
	if err != nil {
		return err
	}

	return assertJSONValue(value, expectedValue, path)
}

// getJSONPath retrieves a value from a JSON structure using dot notation
func getJSONPath(data any, path string) (any, error) {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			current = v[part]
		default:
			return nil, fmt.Errorf("path %s not found: invalid segment %s", path, part)
		}

		if current == nil {
			return nil, fmt.Errorf("path %s not found: %s is nil", path, part)
		}
	}

	return current, nil
}

// GetResponseJSON unmarshals the response body into JSON
func GetResponseJSON(resp *http.Response) (map[string]any, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	defer resp.Body.Close()

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return data, nil
}
