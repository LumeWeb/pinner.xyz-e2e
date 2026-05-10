Feature: Sia Access Control

  As an end-user of the LumeWeb Portal
  I want my Sia data to be protected by proper access controls
  So that only authorized access is allowed to my stored files

  Background:
    Given the admin is authenticated
    And the admin creates a new quota plan named "Default Test Plan"
    And the admin sets the plan as default
    And an existing registered user
    And the user is logged in
    And the user has connected their Sia account
    And the SIA account is ready

  @sia-share-link-valid
  Scenario: User shares a file with a valid time-limited link
    Given the user has uploaded a Sia file
    When the user shares a Sia file with a valid time-limited link
    Then the SIA shared file is accessible

  @sia-share-link-expired
  Scenario: User shares a file with an expired time-limited link
    Given the user has uploaded a Sia file
    When the user shares a Sia file with an expired time-limited link
    Then the SIA shared file access is denied

  @sia-unauthorized-access
  Scenario: Unauthenticated request to Sia is rejected
    When the user attempts to access a Sia file without authentication
    Then the SIA access is denied with authentication required

  @sia-invalid-request
  Scenario: Invalid request to Sia is rejected
    When the user sends an invalid Sia request
    Then the SIA response indicates bad request

  @sia-missing-file
  Scenario: Request for non-existent Sia file returns not found
    When the user requests a Sia file that does not exist
    Then the SIA response indicates file not found
