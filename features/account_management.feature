Feature: Account Management
  This feature tests basic account management operations including API key CRUD
  operations, upload limits checking, and account deletion workflows.

  Background:
    Given an existing registered user
    And the user is logged in

  @create-api-key
  Scenario: User creates an API key
    When the user creates a new API key named "my-test-key"
    Then the API key is created successfully
    And the API key has a unique token

  @list-api-keys
  Scenario: User lists API keys
    Given an existing registered user with an API key
    When the user lists their API keys
    Then the API keys are returned successfully
    And the created API key is in the list

  @delete-api-key-basic
  Scenario: User deletes an API key
    Given an existing registered user with an API key
    When the user deletes the API key
    Then the API key is deleted successfully
    And the API key is no longer in the list

  @check-upload-limits
  Scenario: User checks upload limits
    When the user checks their upload limits
    Then the upload limits are returned successfully
    And the limits include total and used values

  @delete-account
  Scenario: User deletes their account
    When the user deletes their account
    Then the account is deleted successfully
    And the user can no longer login
