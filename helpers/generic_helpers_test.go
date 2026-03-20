package helpers

import (
	"context"
	"testing"
)

// TestGenericContextHelpers verifies that the new generic context helpers work correctly
func TestGenericContextHelpers(t *testing.T) {
	ctx := context.Background()

	// Test SetContextValue and GetContextValue
	ctx = SetContextValue(ctx, JWTTokenKey, "test-token-123")
	token, ok := GetContextValue[string](ctx, JWTTokenKey)
	if !ok {
		t.Error("Expected to retrieve token from context")
	}
	if token != "test-token-123" {
		t.Errorf("Expected token 'test-token-123', got '%s'", token)
	}

	// Test with non-existent key
	_, ok = GetContextValue[string](ctx, OperationIDKey)
	if ok {
		t.Error("Expected false for non-existent key")
	}

	// Test with different types
	testUser := &TestUser{Email: "test@example.com"}
	ctx = SetContextValue(ctx, TestUserKey, testUser)
	retrievedUser, ok := GetContextValue[*TestUser](ctx, TestUserKey)
	if !ok {
		t.Error("Expected to retrieve user from context")
	}
	if retrievedUser.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", retrievedUser.Email)
	}
}

// TestAddToCleanupList verifies the generic cleanup list helper works correctly
func TestAddToCleanupList(t *testing.T) {
	ctx := context.Background()

	// Test adding to cleanup list
	ctx = AddToCleanupList(ctx, APIKeysCleanupKey, "key1")
	ctx = AddToCleanupList(ctx, APIKeysCleanupKey, "key2")
	ctx = AddToCleanupList(ctx, APIKeysCleanupKey, "key3")

	keys := GetCleanupList[string](ctx, APIKeysCleanupKey)
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Verify order is preserved
	expected := []string{"key1", "key2", "key3"}
	for i, key := range expected {
		if keys[i] != key {
			t.Errorf("Expected key '%s' at index %d, got '%s'", key, i, keys[i])
		}
	}
}

// TestBackwardCompatibilityHelpers verifies that legacy helpers still work
func TestBackwardCompatibilityHelpers(t *testing.T) {
	ctx := context.Background()

	// Test SetJWTToken/GetJWTToken
	ctx = SetJWTToken(ctx, "legacy-token-456")
	token, ok := GetJWTToken(ctx)
	if !ok {
		t.Error("Legacy GetJWTToken failed")
	}
	if token != "legacy-token-456" {
		t.Errorf("Expected 'legacy-token-456', got '%s'", token)
	}

	// Test AddAPIKeyCleanup/GetAPIKeysCleanup
	ctx = AddAPIKeyCleanup(ctx, "api-key-1")
	ctx = AddAPIKeyCleanup(ctx, "api-key-2")
	keys := GetAPIKeysCleanup(ctx)
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys from legacy helpers, got %d", len(keys))
	}

	// Test SetTestUser/GetTestUser
	testUser := &TestUser{Email: "legacy@example.com"}
	ctx = SetTestUser(ctx, testUser)
	retrievedUser, ok := GetTestUser(ctx)
	if !ok {
		t.Error("Legacy GetTestUser failed")
	}
	if retrievedUser.Email != "legacy@example.com" {
		t.Errorf("Expected 'legacy@example.com', got '%s'", retrievedUser.Email)
	}
}

// TestAddToCleanupListWithReflection verifies the reflection-based cleanup list handling works
func TestAddToCleanupListWithReflection(t *testing.T) {
	ctx := context.Background()

	// When context is empty, should initialize new list
	ctx = AddToCleanupList(ctx, TestUsersCleanupKey, "user1@test.com")
	users := GetCleanupList[string](ctx, TestUsersCleanupKey)
	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	// When list already exists, should append
	ctx = AddToCleanupList(ctx, TestUsersCleanupKey, "user2@test.com")
	users = GetCleanupList[string](ctx, TestUsersCleanupKey)
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}
