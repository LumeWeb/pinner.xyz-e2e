Feature: Sia Account

  As an end-user of the LumeWeb Portal
  I want to connect and manage my Sia account
  So that I can access decentralized storage through the portal

  Background:
    Given the admin is authenticated

  @sia-connect-account
  Scenario: User connects their Sia account
    Given an existing registered user
    And the user is logged in
    And the email verification email has been received with a token
    And the user verifies their email using the token
    And the email is verified successfully
    When the user initiates a SIA account connection
    Then the SIA connection is pending approval
    And the user approves the SIA connection
    And the SIA connection is established
    And the user saves their SIA recovery phrase
    And the SIA account key is stored

  @sia-connect-reject-retry
  Scenario: User rejects then retries Sia connection
    Given an existing registered user
    And the user is logged in
    And the email verification email has been received with a token
    And the user verifies their email using the token
    And the email is verified successfully
    When the user initiates a SIA account connection
    Then the SIA connection is pending approval
    And the user rejects the SIA connection
    Then the SIA connection fails
    When the user retries the SIA connection
    Then the SIA connection is established and an account key is created

  @sia-view-account
  Scenario: User views their Sia account details
    Given an existing registered user
    And the user is logged in
    And the email verification email has been received with a token
    And the user verifies their email using the token
    And the email is verified successfully
    And the user has connected their Sia account
    When the user views their Sia account
    Then the Sia account shows usage and quota information

  @sia-register-login
  Scenario: New user registers and logs into Sia
    Given a new Sia user registration
    When the Sia user signs in
    Then the SIA account summary is displayed

  @sia-create-app-account
  Scenario: User creates a Sia app account
    Given an existing registered user
    And the user is logged in
    And the email verification email has been received with a token
    And the user verifies their email using the token
    And the email is verified successfully
    And the user has connected their Sia account
    When the user creates a new Sia app account
    Then the SIA app account is linked to the user

  @sia-connect-unauthenticated-redirect
  Scenario: Unauthenticated user is redirected to login from connect page
    Given an existing registered user
    And the email verification email has been received with a token
    And the user verifies their email using the token
    And the email is verified successfully
    And the user initiates a SIA account connection
    When an unauthenticated request is made to the SIA connect page
    Then the response is a redirect to the login page with a return URL

  @sia-connect-ui-page
  Scenario: Authenticated user sees connect approval page
    Given an existing registered user
    And the user is logged in
    And the email verification email has been received with a token
    And the user verifies their email using the token
    And the email is verified successfully
    And the user initiates a SIA account connection
    When an authenticated request is made to the SIA connect page
    Then the response is an HTML page with approval controls
