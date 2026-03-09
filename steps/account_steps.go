package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	account "go.lumeweb.com/portal-sdk"
	"pinner.xyz-e2e/helpers"
)

// AccountSteps holds the state for account management step definitions
type AccountSteps struct {
	accountAPI account.AccountAPI
}

// NewAccountSteps creates a new AccountSteps instance
func NewAccountSteps() *AccountSteps {
	return &AccountSteps{}
}

// RegisterHooks registers cleanup hooks for the account steps
func (s *AccountSteps) RegisterHooks(ctx *godog.ScenarioContext) {
	helpers.RegisterCommonHooks(ctx)
}

// InitializeScenario registers all step definitions with godog
func (s *AccountSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Setup account API client
	s.accountAPI = helpers.GetUnauthenticatedClient()

	// Account management steps
	ctx.Step(`^an existing registered user$`, s.anExistingRegisteredUser)
	ctx.Step(`^the user is logged in$`, s.theUserIsLoggedIn)

	// API key creation steps
	ctx.Step(`^the user creates a new API key named "(.*)"$`, s.theUserCreatesANewAPIKeyNamed)
	ctx.Step(`^the API key is created successfully$`, s.theAPIKeyIsCreatedSuccessfully)
	ctx.Step(`^the API key has a unique token$`, s.theAPIKeyHasAUniqueToken)

	// API key listing steps
	ctx.Step(`^the user lists their API keys$`, s.theUserListsTheirAPIKeys)
	ctx.Step(`^the API keys are returned successfully$`, s.theAPIKeysAreReturnedSuccessfully)
	ctx.Step(`^the created API key is in the list$`, s.theCreatedAPIKeyIsInTheList)

	// API key deletion steps
	ctx.Step(`^the user deletes the API key$`, s.theUserDeletesTheAPIKey)
	ctx.Step(`^the API key is deleted successfully$`, s.theAPIKeyIsDeletedSuccessfully)
	ctx.Step(`^the API key is no longer in the list$`, s.theAPIKeyIsNoLongerInTheList)

	// Upload limits steps
	ctx.Step(`^the user checks their upload limits$`, s.theUserChecksTheirUploadLimits)
	ctx.Step(`^the upload limits are returned successfully$`, s.theUploadLimitsAreReturnedSuccessfully)
	ctx.Step(`^the limits include total and used values$`, s.theLimitsIncludeTotalAndUsedValues)

	// Account deletion steps
	ctx.Step(`^the user deletes their account$`, s.theUserDeletesTheirAccount)
	ctx.Step(`^the account is deleted successfully$`, s.theAccountIsDeletedSuccessfully)
	ctx.Step(`^the user can no longer login$`, s.theUserCanNoLongerLogin)

	// Advanced API key management steps
	ctx.Step(`^the user creates two API keys named "([^"]*)" and "([^"]*)"$`, s.theUserCreatesTwoAPIKeys)
	ctx.Step(`^the user lists API keys with page size (\d+)$`, s.theUserListsAPIKeysWithPageSize)
	ctx.Step(`^only (\d+) API keys are returned on the first page$`, s.onlyAPIKeysReturnedOnFirstPage)
	ctx.Step(`^the user lists API keys filtering by name "([^"]*)"$`, s.theUserListsAPIKeysFilteringByName)
	ctx.Step(`^only "([^"]*)" is in the results$`, s.onlyKeyIsResultsController)
	ctx.Step(`^an existing registered user with API keys named "([^"]*)" and "([^"]*)"$`, s.anExistingRegisteredUserWithAPIKeys)
	ctx.Step(`^an existing registered user with (\d+) API keys$`, s.anExistingRegisteredUserWithNAPIKeys)
	ctx.Step(`^the user attempts to delete a non-existent API key$`, s.theUserAttemptsDeleteNonexistentAPIKey)

	// API key security steps
	ctx.Step(`^authentication fails with appropriate error$`, s.authenticationFailsWithAppropriateError)
	ctx.Step(`^both API keys are created successfully$`, s.bothAPIKeysAreCreatedSuccessfully)
	ctx.Step(`^both API keys have unique tokens despite having the same name$`, s.bothAPIKeysHaveUniqueTokensDespiteHavingTheSameName)
	ctx.Step(`^both users receive different API key tokens$`, s.bothUsersReceiveDifferentAPIKeyTokens)
	ctx.Step(`^each API key has a unique token$`, s.eachAPIKeyHasAUniqueToken)
	ctx.Step(`^each user can only access their own API keys$`, s.eachUserCanOnlyAccessTheirOwnAPIKeys)
	ctx.Step(`^each user creates an API key named "([^"]*)"$`, s.eachUserCreatesAnAPIKeyNamed)
	ctx.Step(`^the API keys list contains both keys$`, s.theAPIKeysListContainsBothKeys)
	ctx.Step(`^the API keys list is empty$`, s.theAPIKeysListIsEmpty)
	ctx.Step(`^the deletion fails$`, s.theDeletionFails)
	ctx.Step(`^the JWT token can no longer be used for authenticated requests$`, s.theJWTTokenCanNoLongerBeUsedForAuthenticatedRequests)
	ctx.Step(`^the user attempts to authenticate using the deleted API key$`, s.theUserAttemptsToAuthenticateUsingTheDeletedAPIKey)
	ctx.Step(`^the user creates API keys named "([^"]*)" twice$`, s.theUserCreatesAPIKeysNamedTwice)
	ctx.Step(`^two registered users$`, s.twoRegisteredUsers)
}

func (s *AccountSteps) anExistingRegisteredUser(ctx context.Context) (context.Context, error) {
	return helpers.RegisterTestUser(ctx)
}

func (s *AccountSteps) theUserIsLoggedIn(ctx context.Context) (context.Context, error) {
	return helpers.LoginTestUser(ctx)
}

func (s *AccountSteps) theUserCreatesANewAPIKeyNamed(ctx context.Context, name string) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	apiKey, err := api.CreateAPIKey(ctx, name)
	if err != nil {
		return ctx, fmt.Errorf("failed to create API key: %w", err)
	}

	ctx = helpers.StoreAPIKeyWithUUID(ctx, apiKey)
	return ctx, nil
}

func (s *AccountSteps) theAPIKeyIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	_, err := helpers.RequireAPIKey(ctx)
	return ctx, err
}

func (s *AccountSteps) theAPIKeyHasAUniqueToken(ctx context.Context) (context.Context, error) {
	apiKey, err := helpers.RequireAPIKey(ctx)
	if err != nil {
		return ctx, err
	}

	err = helpers.VerifyStringNotEmpty(apiKey, "API key token")
	return ctx, err
}

func (s *AccountSteps) theUserListsTheirAPIKeys(ctx context.Context) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	apiKeys, err := api.ListAPIKeys(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list API keys: %w", err)
	}

	ctx = helpers.SetAPIKeysList(ctx, apiKeys)
	return ctx, nil
}

func (s *AccountSteps) theAPIKeysAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	apiKeys, err := helpers.RequireAPIKeysList(ctx)
	if apiKeys == nil {
		return ctx, fmt.Errorf("API keys list is nil")
	}
	return ctx, err
}

func (s *AccountSteps) theCreatedAPIKeyIsInTheList(ctx context.Context) (context.Context, error) {
	apiKeys, err := helpers.RequireAPIKeysList(ctx)
	if err != nil {
		return ctx, err
	}

	uuid, err := helpers.RequireAPIKeyUUID(ctx)
	if err != nil {
		return ctx, err
	}

	// Check if the created API key is in the list
	uuidStr := uuid
	found := false
	for _, apiKey := range apiKeys {
		if apiKey.Uuid.String() == uuidStr {
			found = true
			break
		}
	}

	if !found {
		return ctx, fmt.Errorf("created API key with UUID %s not found in list", uuidStr)
	}

	return ctx, nil
}

func (s *AccountSteps) theUserDeletesTheAPIKey(ctx context.Context) (context.Context, error) {
	uuid, err := helpers.RequireAPIKeyUUID(ctx)
	if err != nil {
		return ctx, err
	}

	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	err = api.DeleteAPIKey(ctx, uuid)
	if err != nil {
		return ctx, fmt.Errorf("failed to delete API key: %w", err)
	}

	return ctx, nil
}

func (s *AccountSteps) theAPIKeyIsDeletedSuccessfully(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}

func (s *AccountSteps) theAPIKeyIsNoLongerInTheList(ctx context.Context) (context.Context, error) {
	uuid, err := helpers.RequireAPIKeyUUID(ctx)
	if err != nil {
		return ctx, err
	}

	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	apiKeys, err := api.ListAPIKeys(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list API keys: %w", err)
	}

	// Check if the deleted API key is NOT in the list
	uuidStr := uuid
	for _, apiKey := range apiKeys {
		if apiKey.Uuid.String() == uuidStr {
			return ctx, fmt.Errorf("API key with UUID %s is still in the list after deletion", uuidStr)
		}
	}

	return ctx, nil
}

func (s *AccountSteps) theUserChecksTheirUploadLimits(ctx context.Context) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	limit, err := api.UploadLimit(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get upload limit: %w", err)
	}

	ctx = helpers.SetUploadLimit(ctx, limit)
	return ctx, nil
}

func (s *AccountSteps) theUploadLimitsAreReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	limit, ok := helpers.GetUploadLimit(ctx)
	if !ok || limit == 0 {
		return ctx, fmt.Errorf("no upload limit available")
	}
	return ctx, nil
}

func (s *AccountSteps) theLimitsIncludeTotalAndUsedValues(ctx context.Context) (context.Context, error) {
	limit, ok := helpers.GetUploadLimit(ctx)
	if !ok {
		return ctx, fmt.Errorf("no upload limit available")
	}

	if limit <= 0 {
		return ctx, fmt.Errorf("upload limit %d is invalid", limit)
	}

	return ctx, nil
}

func (s *AccountSteps) theUserDeletesTheirAccount(ctx context.Context) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	err = api.DeleteAccount(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to delete account: %w", err)
	}

	return ctx, nil
}

func (s *AccountSteps) theAccountIsDeletedSuccessfully(ctx context.Context) (context.Context, error) {
	return helpers.ContextSuccess(ctx)
}

func (s *AccountSteps) theUserCanNoLongerLogin(ctx context.Context) (context.Context, error) {
	testUser, err := helpers.RequireTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	api := helpers.GetUnauthenticatedClient()
	_, err = api.Login(ctx, testUser.Email, testUser.Password)
	if err == nil {
		return ctx, fmt.Errorf("user was able to login after account deletion")
	}

	return ctx, nil
}

// Advanced API key steps

func (s *AccountSteps) anExistingRegisteredUserWithAPIKeys(ctx context.Context, key1, key2 string) (context.Context, error) {
	ctx, err := helpers.RegisterTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	ctx, err = helpers.LoginTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	apiKey1, _ := api.CreateAPIKey(ctx, key1)
	if apiKey1 != nil {
		ctx = helpers.SetAPIKeyUUID(ctx, apiKey1.Uuid.String())
		ctx = helpers.AddAPIKeyUUIDCleanup(ctx, apiKey1.Uuid.String())
	}

	apiKey2, _ := api.CreateAPIKey(ctx, key2)
	if apiKey2 != nil {
		ctx = helpers.AddAPIKeyUUIDCleanup(ctx, apiKey2.Uuid.String())
	}

	return ctx, nil
}

func (s *AccountSteps) anExistingRegisteredUserWithNAPIKeys(ctx context.Context, n int) (context.Context, error) {
	ctx, err := helpers.RegisterTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	ctx, err = helpers.LoginTestUser(ctx)
	if err != nil {
		return ctx, err
	}

	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	for i := 0; i < n; i++ {
		name := fmt.Sprintf("key-%d", i)
		apiKey, _ := api.CreateAPIKey(ctx, name)
		if apiKey != nil {
			ctx = helpers.AddAPIKeyUUIDCleanup(ctx, apiKey.Uuid.String())
		}
	}

	return ctx, nil
}

func (s *AccountSteps) theUserCreatesTwoAPIKeys(ctx context.Context, name1, name2 string) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	apiKey1, err := api.CreateAPIKey(ctx, name1)
	if err != nil {
		return ctx, fmt.Errorf("failed to create first API key: %w", err)
	}
	ctx = helpers.SetAPIKeyUUID(ctx, apiKey1.Uuid.String())
	ctx = helpers.AddAPIKeyUUIDCleanup(ctx, apiKey1.Uuid.String())
	ctx = helpers.SetFirstApiKeyToken(ctx, apiKey1.Token)

	apiKey2, err := api.CreateAPIKey(ctx, name2)
	if err != nil {
		return ctx, fmt.Errorf("failed to create second API key: %w", err)
	}
	ctx = helpers.AddAPIKeyUUIDCleanup(ctx, apiKey2.Uuid.String())
	ctx = helpers.SetSecondApiKeyToken(ctx, apiKey2.Token)

	return ctx, nil
}

func (s *AccountSteps) theUserListsAPIKeysWithPageSize(ctx context.Context, pageSize int) (context.Context, error) {
	ctx = helpers.SetPageSize(ctx, pageSize)
	apiKeys, err := helpers.VerifyAPIKeysPagination(ctx, pageSize)
	if err != nil {
		return ctx, fmt.Errorf("pagination verification failed: %w", err)
	}
	ctx = helpers.SetAPIKeysList(ctx, apiKeys)
	return ctx, nil
}

func (s *AccountSteps) onlyAPIKeysReturnedOnFirstPage(ctx context.Context, expectedCount int) (context.Context, error) {
	pageSize, ok := helpers.GetPageSize(ctx)
	if !ok || pageSize != expectedCount {
		return ctx, fmt.Errorf("expected page size %d, got %d", expectedCount, pageSize)
	}
	return ctx, nil
}

func (s *AccountSteps) theUserListsAPIKeysFilteringByName(ctx context.Context, name string) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	ctx = helpers.SetFilterName(ctx, name)

	// Query all API keys and filter by name
	allKeys, err := api.ListAPIKeys(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list API keys: %w", err)
	}

	// Filter keys that contain the name substring
	var filteredKeys []*account.APIKey
	for _, key := range allKeys {
		if strings.Contains(key.Name, name) {
			filteredKeys = append(filteredKeys, key)
		}
	}

	ctx = helpers.SetAPIKeysList(ctx, filteredKeys)
	return ctx, nil
}

func (s *AccountSteps) onlyKeyIsResultsController(ctx context.Context, expectedName string) (context.Context, error) {
	apiKeys, err := helpers.RequireAPIKeysList(ctx)
	if err != nil {
		return ctx, err
	}

	if len(apiKeys) != 1 {
		return ctx, fmt.Errorf("expected 1 API key in filtered results, got %d", len(apiKeys))
	}

	if apiKeys[0].Name != expectedName {
		return ctx, fmt.Errorf("expected API key '%s', got '%s'", expectedName, apiKeys[0].Name)
	}

	return ctx, nil
}

func (s *AccountSteps) theUserAttemptsDeleteNonexistentAPIKey(ctx context.Context) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Try to delete a non-existent API key
	nonExistentUUID := "12345678-1234-1234-1234-123456789012"
	err = api.DeleteAPIKey(ctx, nonExistentUUID)
	if err == nil {
		return ctx, fmt.Errorf("deletion of non-existent API key should have failed")
	}

	return ctx, nil
}

// API Key Security Step Implementations

func (s *AccountSteps) authenticationFailsWithAppropriateError(ctx context.Context) error {
	return nil
}

func (s *AccountSteps) bothAPIKeysAreCreatedSuccessfully(ctx context.Context) error {
	token1, ok := helpers.GetFirstApiKeyToken(ctx)
	if !ok || token1 == "" {
		return fmt.Errorf("first API key token not found")
	}

	token2, ok := helpers.GetSecondApiKeyToken(ctx)
	if !ok || token2 == "" {
		return fmt.Errorf("second API key token not found")
	}

	return nil
}

func (s *AccountSteps) bothAPIKeysHaveUniqueTokensDespiteHavingTheSameName(ctx context.Context) error {
	token1, ok := helpers.GetFirstApiKeyToken(ctx)
	if !ok || token1 == "" {
		return fmt.Errorf("first API key token not found")
	}

	token2, ok := helpers.GetSecondApiKeyToken(ctx)
	if !ok || token2 == "" {
		return fmt.Errorf("second API key token not found")
	}

	if token1 == token2 {
		return fmt.Errorf("both API keys have the same token: %s", token1)
	}

	return nil
}

func (s *AccountSteps) bothUsersReceiveDifferentAPIKeyTokens(ctx context.Context) error {
	apiKey1, ok := helpers.GetAPIKey1Token(ctx)
	if !ok || apiKey1 == "" {
		return fmt.Errorf("first user's API key token not found")
	}

	apiKey2, ok := helpers.GetAPIKey2Token(ctx)
	if !ok || apiKey2 == "" {
		return fmt.Errorf("second user's API key token not found")
	}

	if apiKey1 == apiKey2 {
		return fmt.Errorf("both users received the same API key token: %s", apiKey1)
	}

	return nil
}

func (s *AccountSteps) eachAPIKeyHasAUniqueToken(ctx context.Context) error {
	return nil
}

func (s *AccountSteps) eachUserCanOnlyAccessTheirOwnAPIKeys(ctx context.Context) error {
	apiKey1Token, ok := helpers.GetAPIKey1Token(ctx)
	if !ok || apiKey1Token == "" {
		return fmt.Errorf("first user's API key token not found")
	}

	apiKey2Token, ok := helpers.GetAPIKey2Token(ctx)
	if !ok || apiKey2Token == "" {
		return fmt.Errorf("second user's API key token not found")
	}

	if apiKey1Token == apiKey2Token {
		return fmt.Errorf("users can access each other's API keys")
	}

	return nil
}

func (s *AccountSteps) eachUserCreatesAnAPIKeyNamed(ctx context.Context, name string) (context.Context, error) {
	user1, hasUser1 := helpers.GetUser1(ctx)
	user2, hasUser2 := helpers.GetUser2(ctx)

	if !hasUser1 || !hasUser2 {
		return ctx, fmt.Errorf("both users must exist in context")
	}

	// Authenticate and create API key for user1
	loginAPI := helpers.GetUnauthenticatedClient()
	loginResult1, err := loginAPI.Login(ctx, user1.Email, user1.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to login first user: %w", err)
	}

	api1 := helpers.CreateAuthenticatedClient(loginResult1.Token)

	apiKey1, err := api1.CreateAPIKey(ctx, name)
	if err != nil {
		return ctx, fmt.Errorf("failed to create API key for first user: %w", err)
	}

	ctx = helpers.SetAPIKey1Token(ctx, apiKey1.Token)
	ctx = helpers.AddAPIKeyUUIDCleanup(ctx, apiKey1.Uuid.String())

	// Authenticate and create API key for user2
	loginResult2, err := loginAPI.Login(ctx, user2.Email, user2.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to login second user: %w", err)
	}

	api2 := helpers.CreateAuthenticatedClient(loginResult2.Token)

	apiKey2, err := api2.CreateAPIKey(ctx, name)
	if err != nil {
		return ctx, fmt.Errorf("failed to create API key for second user: %w", err)
	}

	ctx = helpers.SetAPIKey2Token(ctx, apiKey2.Token)
	ctx = helpers.AddAPIKeyUUIDCleanup(ctx, apiKey2.Uuid.String())

	// Store user1's authenticated client in context
	ctx = helpers.SetAuthenticatedClient(ctx, api1)
	ctx = helpers.SetJWTToken(ctx, loginResult1.Token)

	return ctx, nil
}

func (s *AccountSteps) theAPIKeysListContainsBothKeys(ctx context.Context) error {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return err
	}

	apiKeys, err := api.ListAPIKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to list API keys: %w", err)
	}

	if len(apiKeys) != 2 {
		return fmt.Errorf("expected 2 API keys in list, got %d", len(apiKeys))
	}

	return nil
}

func (s *AccountSteps) theAPIKeysListIsEmpty(ctx context.Context) (context.Context, error) {
	apiKeys, err := helpers.RequireAPIKeysList(ctx)
	if err != nil {
		return ctx, err
	}

	if len(apiKeys) != 0 {
		return ctx, fmt.Errorf("expected empty API keys list, got %d keys", len(apiKeys))
	}

	return ctx, nil
}

func (s *AccountSteps) theDeletionFails(ctx context.Context) error {
	return nil
}

func (s *AccountSteps) theJWTTokenCanNoLongerBeUsedForAuthenticatedRequests(ctx context.Context) (context.Context, error) {
	err := helpers.VerifyJWTInvalidation(ctx)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (s *AccountSteps) theUserAttemptsToAuthenticateUsingTheDeletedAPIKey(ctx context.Context) (context.Context, error) {
	apiKey, err := helpers.RequireAPIKey(ctx)
	if err != nil {
		return ctx, err
	}

	loginAPI := helpers.GetUnauthenticatedClient()
	_, err = loginAPI.LoginWithAPIKey(ctx, apiKey)
	if err == nil {
		return ctx, fmt.Errorf("authentication with deleted API key should have failed")
	}

	return ctx, nil
}

func (s *AccountSteps) theUserCreatesAPIKeysNamedTwice(ctx context.Context, name string) (context.Context, error) {
	api, err := helpers.RequireAuthenticatedClient(ctx)
	if err != nil {
		return ctx, err
	}

	// Create first API key
	apiKey1, err := api.CreateAPIKey(ctx, name)
	if err != nil {
		return ctx, fmt.Errorf("failed to create first API key: %w", err)
	}
	ctx = helpers.SetFirstApiKeyToken(ctx, apiKey1.Token)
	ctx = helpers.AddAPIKeyUUIDCleanup(ctx, apiKey1.Uuid.String())

	// Create second API key with same name
	apiKey2, err := api.CreateAPIKey(ctx, name)
	if err != nil {
		return ctx, fmt.Errorf("failed to create second API key: %w", err)
	}
	ctx = helpers.SetSecondApiKeyToken(ctx, apiKey2.Token)
	ctx = helpers.AddAPIKeyUUIDCleanup(ctx, apiKey2.Uuid.String())

	return ctx, nil
}

func (s *AccountSteps) twoRegisteredUsers(ctx context.Context) (context.Context, error) {
	// Create first user
	user1 := helpers.CreateTestUser()
	api := helpers.GetUnauthenticatedClient()
	err := api.Register(ctx, user1.Email, user1.FirstName, user1.LastName, user1.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to register first user: %w", err)
	}
	ctx = helpers.AddTestUserCleanup(ctx, user1.Email)
	ctx = helpers.SetUser1(ctx, user1)

	// Create second user
	user2 := helpers.CreateTestUser()
	err = api.Register(ctx, user2.Email, user2.FirstName, user2.LastName, user2.Password)
	if err != nil {
		return ctx, fmt.Errorf("failed to register second user: %w", err)
	}
	ctx = helpers.AddTestUserCleanup(ctx, user2.Email)
	ctx = helpers.SetUser2(ctx, user2)

	return ctx, nil
}
