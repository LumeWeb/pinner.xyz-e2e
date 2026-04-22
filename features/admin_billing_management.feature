Feature: Admin Billing Management
  This feature tests admin billing operations including credits, price lines,
  pricing plans, pricing plan periods, user balance management, and subscription operations via separate feature files.

  Background:
    Given the admin is authenticated

  # Credit Management Tests
  @admin-list-credits
  Scenario: Admin lists all credits
    When the admin lists all credits
    Then the credits are returned successfully

  @admin-create-credit
  Scenario: Admin creates a credit for a test user
    Given a test user is registered
    When the admin creates a credit for the test user
    Then the credit is created successfully
    And the credit has the specified amount and manual_adjustment transaction type

  @admin-get-credit
  Scenario: Admin retrieves a specific credit
    Given a test user is registered
    And the admin has created a credit for the test user
    When the admin retrieves the credit by ID
    Then the credit details are returned successfully

  @admin-delete-credit
  Scenario: Admin deletes a credit
    Given a test user is registered
    And the admin has created a credit for the test user
    When the admin deletes the credit
    Then the credit is soft deleted successfully

  @admin-restore-credit
  Scenario: Admin restores a deleted credit
    Given a test user is registered
    And the admin has created and deleted a credit for the test user
    When the admin restores the credit
    Then the credit is restored successfully

  @admin-purge-credits
  Scenario: Admin purges soft-deleted credits older than specified duration
    Given a test user is registered
    And the admin has created and deleted a credit for the test user
    When the admin purges credits older than 1 second
    Then old soft-deleted credits are permanently removed
    And the purge count is recorded

  @admin-get-user-balance
  Scenario: Admin retrieves user's current balance
    Given a test user is registered
    When the admin retrieves the balance for the test user
    Then the user balance is returned successfully

  @admin-list-deleted-credits
  Scenario: Admin lists soft-deleted credits for a user
    Given a test user is registered
    And the admin has created and deleted a credit for the test user
    When the admin lists deleted credits for the test user
    Then the deleted credits are returned successfully

  # Price Line Management Tests
  @admin-list-price-lines
  Scenario: Admin lists all price lines
    When the admin lists all price lines
    Then the price lines are returned successfully

  @admin-create-price-line
  Scenario: Admin creates a new price line
    When the admin creates a new price line named "Storage Plus"
    Then the price line is created successfully
    And the price line has the specified name and description

  @admin-get-price-line
  Scenario: Admin retrieves a specific price line
    Given the admin has created a price line named "Test Line"
    When the admin retrieves the price line by ID
    Then the price line details are returned successfully

  @admin-update-price-line
  Scenario: Admin updates an existing price line
    Given the admin has created a price line named "Test Line"
    When the admin updates the price line with new details
    Then the price line is updated successfully
    And the updated details are reflected

  @admin-delete-price-line
  Scenario: Admin deletes a price line
    Given the admin has created a price line named "Line To Delete"
    When the admin deletes the price line
    Then the price line is deleted successfully

  # Pricing Plan Management Tests
  @admin-list-pricing-plans
  Scenario: Admin lists all pricing plans
    When the admin lists all pricing plans
    Then the pricing plans are returned successfully

  @admin-create-pricing-plan
  Scenario: Admin creates a new pricing plan
    When the admin creates a new pricing plan named "Enterprise"
    Then the pricing plan is created successfully
    And the pricing plan includes the specified periods

  @admin-update-pricing-plan
  Scenario: Admin updates an existing pricing plan
    Given the admin has created a pricing plan named "Test Plan"
    When the admin updates the pricing plan with new details
    Then the pricing plan is updated successfully

  @admin-delete-pricing-plan
  Scenario: Admin deletes a pricing plan
    Given the admin has created a pricing plan named "Plan To Delete"
    When the admin deletes the pricing plan
    Then the pricing plan is deleted successfully

  # Pricing Plan Period Management Tests
  @admin-list-pricing-plan-periods
  Scenario: Admin lists all pricing plan periods
    When the admin lists all pricing plan periods
    Then the pricing plan periods are returned successfully

  @admin-create-pricing-plan-period
  Scenario: Admin creates a new pricing plan period
    When the admin creates a new pricing plan period for a basic plan
    Then the pricing plan period is created successfully

  @admin-get-pricing-plan-period
  Scenario: Admin retrieves a specific pricing plan period
    Given the admin has created a pricing plan period
    When the admin retrieves the pricing plan period by ID
    Then the pricing plan period details are returned successfully

  @admin-update-pricing-plan-period
  Scenario: Admin updates an existing pricing plan period
    Given the admin has created a pricing plan period
    When the admin updates the pricing plan period with new details
    Then the pricing plan period is updated successfully

  @admin-delete-pricing-plan-period
  Scenario: Admin deletes a pricing plan period
    Given the admin has created a pricing plan period
    When the admin deletes the pricing plan period
    Then the pricing plan period is deleted successfully

  # Credit Filtering Tests
  @admin-filter-credits-by-type
  Scenario: Admin filters credits by transaction type
    Given a test user is registered
    And the admin has created multiple credits with different types
    When the admin filters credits by transaction type "manual_adjustment"
    Then only credits with that type are returned

  @admin-filter-credits-by-direction
  Scenario: Admin filters credits by direction
    Given a test user is registered
    And the admin has created multiple credits with different directions
    When the admin filters credits by direction "credit"
    Then only credits with that direction are returned

  # Price Line Plan Management Tests
  @admin-add-plan-to-price-line
  Scenario: Admin adds a plan to a price line
    Given the admin has created a pricing plan named "Test Plan For Price Line"
    And the admin has created a price line named "Test Price Line With Plan"
    When the admin adds the plan to the price line
    Then the plan is added to the price line successfully

  @admin-update-plan-position
  Scenario: Admin updates plan position in price line
    Given the admin has created a price line with multiple plans
    When the admin updates a plan position in the price line
    Then the plan position is updated successfully

  @admin-delete-plan-from-price-line
  Scenario: Admin removes a plan from a price line
    Given the admin has created a price line with a plan
    When the admin removes the plan from the price line
    Then the plan is removed from the price line successfully

