Feature: Basic API Key Management
  This feature tests advanced API key management features including bulk creation,
  pagination, filtering by name, and error handling for invalid operations.

  Background:
    Given an existing registered user
    And the user is logged in

  @create-multiple-api-keys
  Scenario: User creates multiple API keys
    When the user creates two API keys named "key-one" and "key-two"
    Then both API keys are created successfully
    And the API keys list contains both keys
    And each API key has a unique token

  @list-api-keys-pagination
  Scenario: User lists API keys with pagination
    Given an existing registered user with 5 API keys
    When the user lists API keys with page size 2
    Then the API keys are returned successfully
    And only 2 API keys are returned on the first page

  @filter-api-keys-by-name
  Scenario: User filters API keys by name
    Given an existing registered user with API keys named "test-key" and "prod-key"
    When the user lists API keys filtering by name "test"
    Then the API keys are returned successfully
    And only "test-key" is in the results

  @delete-api-key
  Scenario: User deletes an API key
    Given an existing registered user with an API key
    When the user deletes the API key
    Then the API key is deleted successfully
    And the API key is no longer in the list

  @delete-nonexistent-api-key
  Scenario: User cannot delete non-existent API key
    When the user attempts to delete a non-existent API key
    Then the deletion fails
