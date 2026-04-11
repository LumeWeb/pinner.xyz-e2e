Feature: Admin Quota Management
  This feature tests admin quota management operations including quota plans,
  allowances, configuration, and system statistics.

  Background:
    Given the admin is authenticated

  @admin-list-quota-plans
  Scenario: Admin lists all quota plans
    When the admin lists all quota plans
    Then the quota plans are returned successfully

  @admin-create-quota-plan
  Scenario: Admin creates a new quota plan
    When the admin creates a new quota plan named "Enterprise Plan"
    Then the quota plan is created with the specified limits
    And the admin retrieves the quota plan
    And the quota plan information is returned

  @admin-list-quota-allowances
  Scenario: Admin lists all quota allowances
    When the admin lists all quota allowances
    Then the quota allowances are returned successfully

  @admin-create-quota-allowance
  Scenario: Admin creates a quota allowance for a user
    When the admin creates a quota allowance for user 12345
    Then the quota allowance is created successfully

  @admin-view-system-stats
  Scenario: Admin views system-wide quota statistics
    When the admin retrieves system-wide quota statistics
    Then the system statistics include upload, download, and storage usage

  @admin-update-quota-plan
  Scenario: Admin updates an existing quota plan
    Given the admin has created a quota plan named "Test Plan"
    When the admin updates the quota plan with new limits
    Then the quota plan is updated successfully
    And the updated limits are reflected

  @admin-delete-quota-plan
  Scenario: Admin deletes a quota plan
    Given the admin has created a quota plan named "Plan To Delete"
    When the admin deletes the quota plan
    Then the quota plan is deleted successfully

  @admin-set-default-plan
  Scenario: Admin sets a quota plan as default
    Given the admin has created a quota plan named "Default Plan"
    When the admin sets the plan as default

  @admin-update-quota-allowance
  Scenario: Admin updates an existing quota allowance
    Given the admin has created a quota allowance for user 12345
    When the admin updates the allowance with new limits
    Then the allowance is updated successfully

  @admin-delete-quota-allowance
  Scenario: Admin deletes a quota allowance
    Given the admin has created a quota allowance for user 12345
    When the admin deletes the allowance
    Then the allowance is deleted successfully

  @admin-reconcile-quota
  Scenario: Admin reconciles quota for all users
    When the admin reconciles quota for all users
    Then the reconciliation completes successfully
    And the users processed count is recorded

  @admin-cleanup-quota
  Scenario: Admin performs quota cleanup
    When the admin performs quota cleanup with 90 day retention
    Then the cleanup completes successfully
    And the records deleted count is recorded
