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
	"time"

	"github.com/cucumber/godog"
	account "go.lumeweb.com/portal-sdk"
	"go.lumeweb.com/portal-sdk/admin"
)

const (
	// TagNoAuthReset preserves authentication state across scenarios when applied
	TagNoAuthReset = "@noAuthReset"

	// WebsiteStatusPollTimeout is the timeout for website status polling
	// Set to 5 minutes to accommodate background job intervals (janitor runs every 1 minute)
	WebsiteStatusPollTimeout = 5 * time.Minute
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
// This ensures scenarios start clean and resources are properly tracked for cleanup
func beforeScenarioSetup(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	panicHandler := NewPanicHandler("").WithName("beforeScenarioSetup")
	if sc != nil {
		panicHandler.scenarioName = sc.Name
	}
	var recoveredErr error
	defer panicHandler.RecoverFromPanic(&recoveredErr)

	// Record scenario start time for timing metrics
	ctx = SetContextValue(ctx, ScenarioStartTimeKey, time.Now())

	if sc == nil {
		ctx = context.WithValue(ctx, APIKeyUUIDsCleanupKey, []string{})
		ctx = context.WithValue(ctx, TestUsersCleanupKey, []string{})
		ctx = context.WithValue(ctx, OperationsCleanupKey, []string{})
		ctx = context.WithValue(ctx, PinRequestIDsCleanupKey, []string{})
		ctx = context.WithValue(ctx, IPNSKeysCleanupKey, []string{})
		return ctx, nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	logger := NewLogger(sc.Name)
	logger.Debug(ctx, "=== Starting scenario: %s ===", sc.Name)
	
	// Log scenario tags for debugging
	if sc.Tags != nil && len(sc.Tags) > 0 {
		tags := make([]string, len(sc.Tags))
		for i, tag := range sc.Tags {
			tags[i] = tag.Name
		}
		logger.Debug(ctx, "Scenario tags: %v", tags)
	}

	ctx = context.WithValue(ctx, APIKeyUUIDsCleanupKey, []string{})
	ctx = context.WithValue(ctx, TestUsersCleanupKey, []string{})
	ctx = context.WithValue(ctx, OperationsCleanupKey, []string{})
	ctx = context.WithValue(ctx, PinRequestIDsCleanupKey, []string{})
	ctx = context.WithValue(ctx, IPNSKeysCleanupKey, []string{})
	
	// Initialize DNS context keys to prevent cross-scenario contamination when running concurrently
	// Each scenario must have its own isolated DNS context
	ctx = context.WithValue(ctx, DNSZoneCleanupKey, []string{})
	ctx = context.WithValue(ctx, DNSZoneIDKey, "")
	ctx = context.WithValue(ctx, DNSZoneDomainKey, "")
	ctx = context.WithValue(ctx, DNSRecordNameKey, "")
	ctx = context.WithValue(ctx, DNSRecordTypeKey, "")
	logger.Debug(ctx, "Initialized cleanup tracking context")

	token, ok := GetJWTToken(ctx)
	logger.Debug(ctx, "Existing JWT token present: %v", ok)

	hasNoAuthReset := false
	if sc.Tags != nil {
		for _, tag := range sc.Tags {
			if tag != nil && tag.Name == TagNoAuthReset {
				hasNoAuthReset = true
				break
			}
		}
	}

	// Reset authenticated client unless @noAuthReset tag is present
	// This prevents scenarios from leaking authentication state to each other
	if !hasNoAuthReset {
		ctx = context.WithValue(ctx, AuthenticatedClientKey, nil)
		logger.Debug(ctx, "Reset authenticated client")
	}

	// Clean up existing pins if JWT token is available
	// This is necessary for test isolation when multiple scenarios run concurrently
	if ok && token != "" {
		logger.Debug(ctx, "Cleaning up existing pins for test isolation")
		if err := CleanupAllPinsForUser(ctx); err != nil {
			logger.Warn(ctx, "Failed to cleanup existing pins: %v", err)
		} else {
			logger.Info(ctx, "Cleaned up existing pins for test isolation")
		}
	}

	logger.Debug(ctx, "=== Scenario setup complete ===")

	return ctx, recoveredErr
}

// afterScenarioCleanup performs cleanup after each scenario
// Ensures all test resources (API keys, users, operations, IPFS pins, DNS zones, IPNS keys) are cleaned up
func afterScenarioCleanup(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	panicHandler := NewPanicHandler("").WithName("afterScenarioCleanup")
	if sc != nil {
		panicHandler.scenarioName = sc.Name
	}
	defer panicHandler.RecoverFromPanic(&err)

	if ctx == nil {
		return ctx, err
	}

	if sc == nil {
		return ctx, err
	}



	var logger *Logger
	if sc != nil {
		logger = NewLogger(sc.Name)
		logger.Debug(ctx, "=== Starting cleanup for scenario: %s ===", sc.Name)
	}

	// Clean up API keys created during the scenario
	apiKeysToDelete := GetAPIKeyUUIDsCleanup(ctx)
	if logger != nil {
		logger.Debug(ctx, "API keys to cleanup: %d", len(apiKeysToDelete))
	}
	if len(apiKeysToDelete) > 0 {
		for _, uuid := range apiKeysToDelete {
			logger.Debug(ctx, "Deleting API key: %s", uuid)
			if deleteErr := DeleteAPIKeyGracefully(ctx, uuid); deleteErr != nil {
				logger.Error(ctx, "Failed to delete API key %s: %v", uuid, deleteErr)
			} else {
				logger.Debug(ctx, "Successfully deleted API key: %s", uuid)
			}
		}
	}

	// Clean up test users created during the scenario
	testUsersToDelete := GetTestUsersCleanup(ctx)
	if logger != nil {
		logger.Debug(ctx, "Test users to cleanup: %d", len(testUsersToDelete))
	}
	if len(testUsersToDelete) > 0 {
		for _, email := range testUsersToDelete {
			logger.Debug(ctx, "Deleting test user: %s", email)
			if deleteErr := DeleteTestUserGracefully(ctx, email); deleteErr != nil {
				logger.Error(ctx, "Failed to delete test user %s: %v", email, deleteErr)
			} else {
				logger.Debug(ctx, "Successfully deleted test user: %s", email)
			}
		}
	}

	operationIDs := GetOperationsCleanup(ctx)
	if logger != nil {
		logger.Debug(ctx, "Operations to track cleanup: %d", len(operationIDs))
	}
	for _, idStr := range operationIDs {
		logger.Debug(ctx, "Operation %s cleanup tracking", idStr)
	}

	// Clean up IPFS assets created during the scenario
	if logger != nil {
		logger.Debug(ctx, "Cleaning up IPFS assets")
	}
	cleanupIPFSAssets(ctx)

	// Clean up DNS zones created during the scenario
	if logger != nil {
		logger.Debug(ctx, "Cleaning up DNS zones")
	}
	if err := CleanupDNSZones(ctx); err != nil {
		logger.Error(ctx, "Failed to cleanup DNS zones: %v", err)
	}

	// Clean up websites created during the scenario
	// Must happen before IPNS keys cleanup since websites may reference IPNS keys
	if logger != nil {
		logger.Debug(ctx, "Cleaning up websites")
	}
	if err := CleanupWebsites(ctx); err != nil {
		logger.Error(ctx, "Failed to cleanup websites: %v", err)
	}

	// Clean up IPNS keys created during the scenario
	// Must happen after websites cleanup since keys are blocked while referenced by active websites
	if logger != nil {
		logger.Debug(ctx, "Cleaning up IPNS keys")
	}
	if err := CleanupIPNSKeys(ctx); err != nil {
		logger.Error(ctx, "Failed to cleanup IPNS keys: %v", err)
	}

	// Clean up admin quota resources created during the scenario
	if logger != nil {
		logger.Debug(ctx, "Cleaning up admin quota resources")
	}
	if err := CleanupAdminQuota(ctx); err != nil {
		logger.Error(ctx, "Failed to cleanup admin quota resources: %v", err)
	}

	// Record scenario timing after all cleanup to avoid output interleaving
	if startTime, ok := GetContextValue[time.Time](ctx, ScenarioStartTimeKey); ok {
		elapsed := time.Since(startTime)
		RecordScenarioTiming(sc.Name, elapsed)
		if logger != nil {
			logger.Debug(ctx, "Scenario completed in: %v", elapsed)
		}
	}

	if logger != nil {
		logger.Debug(ctx, "=== Cleanup complete ===")
	}

	return ctx, err
}

// getPortalConfig retrieves common portal configuration values
type portalConfig struct {
	domain string
	port   string
	secure string
}

func getPortalConfig() portalConfig {
	return portalConfig{
		domain: GetEnv("PORTAL__CORE__DOMAIN", "localhost"),
		port:   GetEnv("PORTAL__CORE__PORT", "8080"),
		secure: GetEnv("PORTAL__CORE__SECURE", "false"),
	}
}

// buildHost returns a formatted host string
// If subdomain is provided, returns subdomain.domain:port
// If subdomain is empty, returns domain:port
func buildHost(subdomain string) string {
	cfg := getPortalConfig()
	if subdomain != "" {
		return fmt.Sprintf("%s.%s:%s", subdomain, cfg.domain, cfg.port)
	}
	return fmt.Sprintf("%s:%s", cfg.domain, cfg.port)
}

// GetPortalHost returns the vhost hostname for the portal API (for Host header)
func GetPortalHost() string {
	return buildHost("account")
}

// GetIPFSHost returns the vhost hostname for the IPFS API (for Host header)
func GetIPFSHost() string {
	return buildHost("ipfs")
}

// GetPortalTarget returns the target address for the portal API (actual connection target)
// Returns just host:port without protocol for WithHostOverride (e.g., "localhost:8080")
func GetPortalTarget() string {
	return buildHost("")
}

// buildEndpoint creates an endpoint URL for a given subdomain
// subdomain is the portal subdomain (e.g., "account", "ipfs")
// includePort determines whether to include the port number
func buildEndpoint(subdomain string, includePort bool) string {
	cfg := getPortalConfig()

	protocol := "http"
	if cfg.secure == "true" || cfg.secure == "True" {
		protocol = "https"
	}

	if includePort {
		return fmt.Sprintf("%s://%s.%s:%s", protocol, subdomain, cfg.domain, cfg.port)
	}
	return fmt.Sprintf("%s://%s.%s", protocol, subdomain, cfg.domain)
}

// GetPortalEndpoint returns the full endpoint URL for the portal API
func GetPortalEndpoint() string {
	return buildEndpoint("account", true)
}

// GetIPFSEndpoint returns the base URL for the IPFS API without /api suffix
// The SDK will add /api/... paths to this base URL when making requests
func GetIPFSEndpoint() string {
	return buildEndpoint("ipfs", true)
}

// GetPortalServer returns the server hostname for SDK endpoint configuration
func GetPortalServer() string {
	return buildEndpoint("account", false)
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
// Ensures the admin account exists before registering the scenario user
func RegisterTestUser(ctx context.Context) (context.Context, error) {
	// Ensure admin account exists before any user registration
	// The first registered user gets admin privileges, so we must create admin first
	_, err := EnsureAdminAccountExists(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to ensure admin account: %w", err)
	}

	api := GetUnauthenticatedClient()
	testUser := CreateTestUser()

	err = api.Register(ctx, testUser.Email, testUser.FirstName, testUser.LastName, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to register user: %w", err)
	}

	ctx = SetTestUser(ctx, testUser)
	ctx = AddTestUserCleanup(ctx, testUser.Email)
	return ctx, nil
}

// LoginTestUser authenticates a test user and stores JWT token in context
// Handles both password-only and 2FA authentication flows
func LoginTestUser(ctx context.Context) (context.Context, error) {
	testUser, ok := GetTestUser(ctx)
	if !ok {
		return ctx, fmt.Errorf("no test user available")
	}

	api := GetUnauthenticatedClient()
	loginResult, err := api.Login(ctx, testUser.Email, testUser.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to login: %w", err)
	}

	if loginResult == nil || loginResult.Token == "" {
		return ctx, fmt.Errorf("login did not return a token")
	}

	// Handle 2FA verification if required
	// This ensures tests support both 2FA-enabled and 2FA-disabled accounts
	if loginResult.OTPRequired {
		otpSecret, hasSecret := GetOTPSecret(ctx)
		if !hasSecret {
			return ctx, fmt.Errorf("OTP required but no OTP secret available in context")
		}

		otpCode, err := GenerateValidTOTPCode(otpSecret)
		if err != nil {
			return ctx, fmt.Errorf("failed to generate TOTP code: %w", err)
		}

		loginAPI := GetUnauthenticatedClient()
		finalToken, err := loginAPI.ValidateOTP(ctx, loginResult.Token, otpCode)
		if err != nil {
			return ctx, fmt.Errorf("failed to validate OTP: %w", err)
		}

		if finalToken == "" {
			return ctx, fmt.Errorf("OTP verification did not return a token")
		}

		loginResult.Token = finalToken
	}

	// Store authenticated client and JWT token for use by subsequent steps
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

// cleanupIPFSAssets cleans up IPFS assets created during the scenario
// cleanupIPFSAssets removes IPFS pins created during the scenario
// Pins are removed to prevent accumulation across test runs
func cleanupIPFSAssets(ctx context.Context) {
	panicHandler := NewPanicHandler("cleanupIPFSAssets").WithName("cleanupIPFSAssets")
	defer panicHandler.RecoverFromPanic(nil)

	client, err := GetIPFSClient(ctx)
	if err != nil {
		return
	}

	if client == nil {
		return
	}

	pinRequestIDs := GetPinRequestIDsCleanup(ctx)
	pinningAPI := client.Pinning()
	if pinningAPI == nil {
		return
	}

	// Remove all pins created during the scenario
	// Errors are intentionally ignored here since we want to attempt cleanup
	// even if some pins have already been deleted or don't exist
	for _, requestID := range pinRequestIDs {
		pinningAPI.RemovePin(ctx, requestID)
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

// Admin client helpers for E2E testing

// AdminClientContextKey is the context key for storing admin client
const AdminClientContextKey contextKey = "admin_client"

// SetAdminClient stores the admin client in context using generic helper
func SetAdminClient(ctx context.Context, client *admin.AdminClient) context.Context {
	return SetContextValue(ctx, AdminClientContextKey, client)
}

// GetAdminClient retrieves the admin client from context using generic helper
func GetAdminClient(ctx context.Context) *admin.AdminClient {
	client, _ := GetContextValue[*admin.AdminClient](ctx, AdminClientContextKey)
	return client
}

// RequireAdminClient retrieves the admin client from context or returns an error
func RequireAdminClient(ctx context.Context) (*admin.AdminClient, error) {
	client := GetAdminClient(ctx)
	if client == nil {
		fmt.Printf("Error: no admin client available in context\n")
		return nil, fmt.Errorf("no admin client available in context")
	}
	return client, nil
}
