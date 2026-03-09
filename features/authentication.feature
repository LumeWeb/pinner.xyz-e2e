Feature: Authentication
  This feature tests authentication workflows including user registration, email/password
  login, API key authentication, and two-factor authentication enable/disable functionality.

  @register-with-email-password
  Scenario: User registers with email and password
    Given a new user registration request
    When the user submits valid registration data
    Then the user is successfully registered
    And the user can login with the registered credentials

  @login-with-email-password
  Scenario: User login with email and password
    Given an existing registered user
    When the user submits valid login credentials
    Then the user receives a valid JWT token
    And the JWT token can be used for authenticated requests

  @login-with-api-key
  Scenario: User login with API key
    Given an existing registered user with an API key
    When the user authenticates using the API key
    Then the user receives a valid JWT token
    And the JWT token can be used for authenticated requests

  @enable-two-factor-auth
  Scenario: User enables two-factor authentication
    Given an existing registered user
    And the user is logged in
    When the user requests to enable 2FA
    Then an OTP secret is generated
    And the user can verify OTP codes
    And 2FA is successfully enabled

  @login-with-two-factor-auth
  Scenario: User login with 2FA enabled
    Given an existing registered user with 2FA enabled
    When the user submits valid login credentials
    Then the user is prompted for OTP verification
    And the user can login with valid OTP code
    And invalid OTP codes are rejected

  @disable-two-factor-auth
  Scenario: User disables two-factor authentication
    Given an existing registered user with 2FA enabled
    And the user is logged in
    When the user requests to disable 2FA
    Then 2FA is successfully disabled
    And the user can login without OTP verification
