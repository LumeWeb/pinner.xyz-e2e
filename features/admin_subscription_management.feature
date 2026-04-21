Feature: Admin Subscription Management

  Admin operations on user subscriptions including cancellation, plan changes,
  and subscriber management.

  Background:
    Given the admin is authenticated
    And the payment mock is reset
    And the billing infrastructure is set up
    And a test user is registered
    And the test user has an active subscription

  @admin-list-all-subscribers
  Scenario: Admin lists all subscribers
    When the admin lists all subscribers
    Then the subscribers are returned successfully

  @admin-retrieve-subscriber-by-id
  Scenario: Admin retrieves subscriber by ID
    Given there is an active subscriber
    When the admin retrieves the subscriber by ID
    Then the subscriber details are returned successfully

  @admin-list-gateway-subscribers
  Scenario: Admin lists subscribers for a specific gateway
    Given there is a gateway with subscribers
    When the admin lists subscribers for the gateway
    Then the gateway subscribers are returned successfully

  @admin-list-user-subscribers
  Scenario: Admin retrieves subscribers for a specific user
    Given a test user with subscriptions
    When the admin retrieves subscribers for the test user
    Then the user subscribers are returned successfully

  @admin-cancel-immediate
  Scenario: Admin cancels subscription immediately
    When the admin cancels the subscription immediately
    Then the subscription is cancelled successfully
    And the cancellation takes effect immediately

  @admin-cancel-end-period
  Scenario: Admin cancels subscription at end of period
    When the admin cancels the subscription at end of period
    Then the subscription is scheduled for cancellation
    And the cancellation will take effect at billing period end

  @admin-change-plan
  Scenario: Admin changes user plan
    Given there is an available pricing plan period
    When the admin changes the user's plan
    Then the plan change is processed successfully
    And the new pricing plan period is applied

  @admin-abort-cancellation
  Scenario: Admin aborts scheduled subscription cancellation
    Given there is a subscription scheduled for cancellation at period end
    When the admin aborts the subscription cancellation
    Then the scheduled cancellation is aborted successfully
    And the subscription status is "active"

  @admin-pause-subscription
  Scenario: Admin pauses user subscription
    When the admin pauses the user's subscription
    Then the pause operation is successful
    And the subscription status is "paused"

  @admin-resume-subscription
  Scenario: Admin resumes paused subscription
    Given the user's subscription is paused
    When the admin resumes the user's subscription
    Then the resume operation is successful
    And the subscription status is "active"