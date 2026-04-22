Feature: Atlos Subscription Lifecycle

  Atlos-specific subscription flow including checkout via simulated widget API calls,
  payment simulation, renewals, cancellation behavior, plan upgrades/downgrades, and
  billing cadence switches via Atlos crypto payments with postback notifications.
  Since Atlos uses crypto payments with postback notifications instead of
  Stripe-style webhooks, these tests verify the portal correctly handles
  Atlos postbacks and subscription lifecycle.

  Background:
    Given the admin is authenticated
    And the atlos gateway is active
    And the payment mock is reset
    And the billing infrastructure is set up

  @atlos-checkout-complete
  Scenario: User creates checkout and Atlos payment completes
    Given an existing registered user
    And the user is logged in
    And a pricing plan period exists
    When the user creates a checkout session for the plan
    And the Atlos checkout session completes
    Then the user has an active subscription
    And the Atlos subscription becomes active

  @atlos-cancel-scheduled
  Scenario: User cancels subscription at end of period
    Given an existing registered user
    And the user is logged in
    And the user has an active Atlos subscription
    When the user cancels the subscription
    Then the subscription status is "cancel_at_period_end"

  @atlos-abort-cancellation
  Scenario: User aborts scheduled cancellation
    Given an existing registered user
    And the user is logged in
    And the user has an active Atlos subscription
    And the user has a subscription scheduled for cancellation via Atlos
    When the user aborts the subscription cancellation
    Then the subscription status is "active"

  @atlos-renewal
  Scenario: User renews subscription via new Atlos payment
    Given an existing registered user
    And the user is logged in
    And the user has an active Atlos subscription
    And the user pays for the next billing period via Atlos
    When the Atlos subscription is renewed
    Then the subscription status is "active"

  @atlos-subscription-renewal
  Scenario: Subscription renews automatically with Atlos payment
    Given an existing registered user
    And the user is logged in
    And the user has an active Atlos subscription
    When the Atlos subscription is renewed
    Then the subscription status is "active"

  @atlos-upgrade-basic-to-pro
  Scenario: User upgrades from Basic to Pro plan via Atlos
    Given an existing registered user
    And the user is logged in
    And the user has an active subscription on the Basic plan
    When the user upgrades to the Pro plan
    And the user's current plan is Pro
    And the subscription status is "active"

  @atlos-upgrade-basic-to-enterprise
  Scenario: User upgrades from Basic to Enterprise plan via Atlos
    Given an existing registered user
    And the user is logged in
    And the user has an active subscription on the Basic plan
    When the user upgrades to the Enterprise plan
    And the user's current plan is Enterprise
    And the subscription status is "active"

  @atlos-upgrade-pro-to-enterprise
  Scenario: User upgrades from Pro to Enterprise plan via Atlos
    Given an existing registered user
    And the user is logged in
    And the user has an active subscription on the Pro plan
    When the user upgrades to the Enterprise plan
    And the user's current plan is Enterprise
    And the subscription status is "active"

  @atlos-downgrade-pro-to-basic
  Scenario: User downgrades from Pro to Basic plan via Atlos
    Given an existing registered user
    And the user is logged in
    And the user has an active subscription on the Pro plan
    When the user downgrades to the Basic plan
    And the user's current plan is Basic
    And the subscription status is "active"

  @atlos-downgrade-enterprise-to-pro
  Scenario: User downgrades from Enterprise to Pro plan via Atlos
    Given an existing registered user
    And the user is logged in
    And the user has an active subscription on the Enterprise plan
    When the user downgrades to the Pro plan
    And the user's current plan is Pro
    And the subscription status is "active"

  @atlos-downgrade-enterprise-to-basic
  Scenario: User downgrades from Enterprise to Basic plan via Atlos
    Given an existing registered user
    And the user is logged in
    And the user has an active subscription on the Enterprise plan
    When the user downgrades to the Basic plan
    And the user's current plan is Basic
    And the subscription status is "active"

  @atlos-upgrade-basic-to-pro-monthly
  Scenario: User upgrades from Basic to Pro with monthly billing via Atlos
    Given an existing registered user
    And the user is logged in
    And the user has an active monthly subscription on the Basic plan
    When the user upgrades to the Pro plan
    And the user's current plan is Pro
    And the subscription status is "active"

  @atlos-upgrade-basic-to-enterprise-yearly
  Scenario: User upgrades from Basic to Enterprise with yearly billing via Atlos
    Given an existing registered user
    And the user is logged in
    And the user has an active yearly subscription on the Basic plan
    When the user upgrades to the Enterprise plan
    And the user's current plan is Enterprise
    And the subscription status is "active"

  @atlos-downgrade-pro-to-basic-monthly
  Scenario: User downgrades from Pro to Basic with monthly billing via Atlos
    Given an existing registered user
    And the user is logged in
    And the user has an active monthly subscription on the Pro plan
    When the user downgrades to the Basic plan
    And the user's current plan is Basic
    And the subscription status is "active"

  @atlos-downgrade-enterprise-to-pro-yearly
  Scenario: User downgrades from Enterprise to Pro with yearly billing via Atlos
    Given an existing registered user
    And the user is logged in
    And the user has an active yearly subscription on the Enterprise plan
    When the user downgrades to the Pro plan
    And the user's current plan is Pro
    And the subscription status is "active"

  @atlos-upgrade-with-cadence-switch
  Scenario: User upgrades with billing cadence change (monthly to yearly) via Atlos
    Given an existing registered user
    And the user is logged in
    And the user has an active monthly subscription on the Basic plan
    When the user selects yearly billing
    And the user upgrades to the Pro plan
    And the user's current plan is Pro
    And the subscription status is "active"

  @atlos-downgrade-with-cadence-switch
  Scenario: User downgrades with billing cadence change (yearly to monthly) via Atlos
    Given an existing registered user
    And the user is logged in
    And the user has an active yearly subscription on the Enterprise plan
    When the user selects monthly billing
    And the user downgrades to the Pro plan
    And the user's current plan is Pro
    And the subscription status is "active"
