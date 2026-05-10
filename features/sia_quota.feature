Feature: Sia Storage Quota

  As an end-user of the LumeWeb Portal
  I want to manage my Sia storage quota
  So that I can control how much data I store and add capacity when needed

  Background:
    Given the admin is authenticated
    And the admin creates a new quota plan named "Default Test Plan"
    And the admin sets the plan as default
    And an existing registered user
    And the user is logged in
    And the user has connected their Sia account
    And the SIA account is ready

  @sia-quota-exceeded
  Scenario: Upload is blocked when Sia storage quota is exceeded
    Given the user has a Sia storage quota limit
    When the user uploads data exceeding the Sia quota
    Then the SIA upload is blocked with quota exceeded

  @sia-quota-increase
  Scenario: User can upload after Sia storage quota is increased
    Given the user has reached their Sia storage quota
    When the SIA account is credited with additional storage
    Then the user can resume Sia uploads up to the new quota
