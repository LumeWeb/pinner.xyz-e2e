Feature: Stripe Subscription Lifecycle

  Stripe-specific subscription flow including checkout via Stripe,
  payment simulation, renewals, cancel_at_period_end behavior,
  pause/resume operations, plan upgrades/downgrades, and
  billing cadence switches via Stripe's subscription actions.

  Background:
    Given the admin is authenticated
    And the stripe gateway is active
    And the stripe mock is reset
    And the billing infrastructure is set up

  @stripe-checkout-complete
  Scenario: User creates checkout and Stripe session completes
    Given an existing registered user
    And the user is logged in
    And a pricing plan period exists
    When the user creates a checkout session for the plan
    And the Stripe checkout session completes
    Then the user has an active subscription
    And the Stripe subscription becomes active

  @stripe-cancel-at-period-end
  Scenario: Subscription is canceled at period end via Stripe
    Given an existing registered user
    And the user is logged in
    And the user has an active Stripe subscription
    And the user has a subscription scheduled for cancellation via Stripe
    When the Stripe subscription is canceled at period end
    Then the subscription status is "canceled"

  @stripe-pause
  Scenario: Subscription is paused via Stripe
    Given an existing registered user
    And the user is logged in
    And the user has an active Stripe subscription
    When the Stripe subscription is paused
    Then the subscription status is "paused"

  @stripe-resume
  Scenario: Subscription is resumed via Stripe
    Given an existing registered user
    And the user is logged in
    And the user has a paused Stripe subscription
    When the Stripe subscription is resumed
    Then the subscription status is "active"

  @stripe-renew-after-cancel
  Scenario: User renews subscription after cancellation
    Given an existing registered user
    And the user is logged in
    And the user has an active Stripe subscription
    And the user has a subscription scheduled for cancellation via Stripe
    And the user pays for the next billing period via Stripe
    When the Stripe subscription is renewed
    Then the subscription status is "active"

  @stripe-upgrade-basic-to-pro
  Scenario: User upgrades from Basic to Pro plan
    Given an existing registered user
    And the user is logged in
    And the user has an active subscription on the Basic plan
    When the user upgrades to the Pro plan
    And the user's current plan is Pro
    And the subscription status is "active"

  @stripe-upgrade-basic-to-enterprise
  Scenario: User upgrades from Basic to Enterprise plan
    Given an existing registered user
    And the user is logged in
    And the user has an active subscription on the Basic plan
    When the user upgrades to the Enterprise plan
    And the user's current plan is Enterprise
    And the subscription status is "active"

  @stripe-upgrade-pro-to-enterprise
  Scenario: User upgrades from Pro to Enterprise plan
    Given an existing registered user
    And the user is logged in
    And the user has an active subscription on the Pro plan
    When the user upgrades to the Enterprise plan
    And the user's current plan is Enterprise
    And the subscription status is "active"

  @stripe-downgrade-pro-to-basic
  Scenario: User downgrades from Pro to Basic plan
    Given an existing registered user
    And the user is logged in
    And the user has an active subscription on the Pro plan
    When the user downgrades to the Basic plan
    And the user's current plan is Basic
    And the subscription status is "active"

  @stripe-downgrade-enterprise-to-pro
  Scenario: User downgrades from Enterprise to Pro plan
    Given an existing registered user
    And the user is logged in
    And the user has an active subscription on the Enterprise plan
    When the user downgrades to the Pro plan
    And the user's current plan is Pro
    And the subscription status is "active"

  @stripe-downgrade-enterprise-to-basic
  Scenario: User downgrades from Enterprise to Basic plan
    Given an existing registered user
    And the user is logged in
    And the user has an active subscription on the Enterprise plan
    When the user downgrades to the Basic plan
    And the user's current plan is Basic
    And the subscription status is "active"

  @stripe-upgrade-basic-to-pro-monthly
  Scenario: User upgrades from Basic to Pro with monthly billing
    Given an existing registered user
    And the user is logged in
    And the user has an active monthly subscription on the Basic plan
    When the user upgrades to the Pro plan
    And the user's current plan is Pro
    And the subscription status is "active"

  @stripe-upgrade-basic-to-enterprise-yearly
  Scenario: User upgrades from Basic to Enterprise with yearly billing
    Given an existing registered user
    And the user is logged in
    And the user has an active yearly subscription on the Basic plan
    When the user upgrades to the Enterprise plan
    And the user's current plan is Enterprise
    And the subscription status is "active"

  @stripe-downgrade-pro-to-basic-monthly
  Scenario: User downgrades from Pro to Basic with monthly billing
    Given an existing registered user
    And the user is logged in
    And the user has an active monthly subscription on the Pro plan
    When the user downgrades to the Basic plan
    And the user's current plan is Basic
    And the subscription status is "active"

  @stripe-downgrade-enterprise-to-pro-yearly
  Scenario: User downgrades from Enterprise to Pro with yearly billing
    Given an existing registered user
    And the user is logged in
    And the user has an active yearly subscription on the Enterprise plan
    When the user downgrades to the Pro plan
    And the user's current plan is Pro
    And the subscription status is "active"

  @stripe-upgrade-with-cadence-switch
  Scenario: User upgrades with billing cadence change (monthly to yearly)
    Given an existing registered user
    And the user is logged in
    And the user has an active monthly subscription on the Basic plan
    When the user selects yearly billing
    And the user upgrades to the Pro plan
    And the user's current plan is Pro
    And the subscription status is "active"

  @stripe-downgrade-with-cadence-switch
  Scenario: User downgrades with billing cadence change (yearly to monthly)
    Given an existing registered user
    And the user is logged in
    And the user has an active yearly subscription on the Enterprise plan
    When the user selects monthly billing
    And the user downgrades to the Pro plan
    And the user's current plan is Pro
    And the subscription status is "active"
