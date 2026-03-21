Feature: IPNS Key Management

  As an end-user of the LumeWeb Portal
  I want to manage IPNS keys
  So that I can publish content with consistent names and update them over time

  Background:
    Given an existing registered user
    And the user is logged in

  @ipns-create-key
  Scenario: User creates an IPNS key
    When the user creates an IPNS key named "my-key"
    Then an IPNS key with name "my-key" is created
    And the IPNS key has a valid peer ID

  @ipns-list-keys
  Scenario: User lists IPNS keys
    Given the user has 3 IPNS keys
    When the user lists their IPNS keys
    Then all 3 IPNS keys are returned
    And each key has a unique peer ID

  @ipns-delete-key
  Scenario: User deletes an IPNS key
    Given the user has an IPNS key named "test-key"
    When the user deletes the IPNS key
    Then the IPNS key is no longer in the list
