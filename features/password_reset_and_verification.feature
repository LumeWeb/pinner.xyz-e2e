Feature: Password Reset and Email Verification
  This feature tests password reset workflows and email verification processes including
  error handling for invalid tokens and email resend functionality.

  @request-password-reset
  Scenario: User requests password reset
    Given an existing registered user
    When the user requests a password reset for their email
    Then the password reset request is successful

  @complete-password-reset-with-valid-token
  Scenario: User completes password reset with valid token
    Given an existing registered user
    And the user has requested a password reset
    And the password reset email has been received with a token
    When the user resets their password using the token
    Then the password is updated successfully
    And the user can login with the new password

  @cannot-reset-password-with-invalid-token
  Scenario: User cannot reset password with invalid token
    Given an existing registered user
    And the user has requested a password reset
    When the user attempts to reset their password with an invalid token
    Then the password reset fails

  @verify-email-address
  Scenario: User verifies their email address
    Given a newly registered user
    And the email verification email has been received with a token
    When the user verifies their email using the token
    Then the email is verified successfully

  @cannot-verify-email-with-invalid-token
  Scenario: User cannot verify email with invalid token
    Given a newly registered user
    When the user attempts to verify their email with an invalid token
    Then the email verification fails

  @resend-email-verification
  Scenario: User resends email verification
    Given a newly registered user
    When the user requests a new verification email
    Then the verification email is sent successfully
    And the verification email is received
