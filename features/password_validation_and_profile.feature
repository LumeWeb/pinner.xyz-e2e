Feature: Password Validation and Profile Management
  This feature tests password validation rules, password change functionality,
  session management, and user profile operations including first name and
  last name updates.

  @register-rejects-too-short-password
  Scenario: Registration rejects passwords that are too short
    Given a new user registration attempt with password "Abc1"
    When the user submits registration data
    Then the registration fails with password validation error

  @register-rejects-invalid-email-format
  Scenario: Registration rejects invalid email format
    Given a new user registration attempt with email "not-an-email"
    When the user submits registration data
    Then the registration fails with email validation error

  @change-password-with-valid-current-password
  Scenario: User changes password with valid current password
    Given an existing registered user
    And the user is logged in
    When the user changes their password to "NewPassword123"
    Then the password is changed successfully
    And the user can login with the new password

  @change-password-with-invalid-current-password
  Scenario: User cannot change password with invalid current password
    Given an existing registered user
    And the user is logged in
    When the user attempts to change password with invalid current password
    Then the password change fails

  @change-password-too-short
  Scenario: User cannot change password to short password
    Given an existing registered user
    And the user is logged in
    When the user attempts to change password to "Abc1"
    Then the password change fails with validation error

  @logout-invalidates-session
  Scenario: Logout invalidates user session
    Given an existing registered user
    And the user is logged in
    When the user logs out
    Then the session is invalidated
    And the JWT token can no longer be used for authenticated requests

  @update-profile-first-name
  Scenario: User updates profile first name
    Given an existing registered user
    And the user is logged in
    When the user updates their first name to "NewFirstName"
    Then the profile is updated successfully
    And the updated profile reflects the new first name

  @update-profile-last-name
  Scenario: User updates profile last name
    Given an existing registered user
    And the user is logged in
    When the user updates their last name to "NewLastName"
    Then the profile is updated successfully
    And the updated profile reflects the new last name

  @get-profile
  Scenario: User retrieves their profile information
    Given an existing registered user
    And the user is logged in
    When the user retrieves their profile
    Then the profile information is returned successfully
    And the profile contains email
    And the profile contains first name and last name
