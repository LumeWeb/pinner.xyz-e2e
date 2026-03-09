Feature: API Key Security and Management

  This feature validates security aspects of API key management including
  proper cleanup, expiration behavior, and session security.

  @api-key-unique-across-users
  Scenario: API keys are unique across different users
    Given two registered users
    When each user creates an API key named "my-api-key"
    Then both users receive different API key tokens
    And each user can only access their own API keys

  @api-key-delete-cleanup
  Scenario: Deleted API keys cannot be used for authentication
    Given an existing registered user with an API key
    When the user deletes the API key
    And the user attempts to authenticate using the deleted API key
    Then authentication fails with appropriate error

  @multiple-api-keys-name-conflict
  Scenario: Can create multiple API keys with same name
    Given an existing registered user
    And the user is logged in
    When the user creates API keys named "my-key" twice
    Then both API keys are created successfully
    And both API keys have unique tokens despite having the same name

  @api-key-list-empty-when-none
  Scenario: List API keys returns empty when none exist
    Given an existing registered user
    And the user is logged in
    When the user lists their API keys
    Then the API keys list is empty
