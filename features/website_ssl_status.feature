@website-ssl-status
Feature: Website SSL Status

  Website SSL status tests verify SSL certificate provisioning and status
  tracking for websites with DNS hosting enabled. SSL status is reported
  via Caddy webhook calls to the internal API with gateway secret authentication.

  Background:
    Given the admin is authenticated
    And the admin creates a new quota plan named "Default Test Plan"
    And the admin sets the plan as default
    And the user has an authenticated API key
    And the user has an IPFS client connection
    And the DNS dev server is available

  @website-get-ssl-status
  Scenario: User queries SSL status for a domain
    Given the user has a website
    When the user checks SSL status for domain "example.com"
    Then the SSL status is stored in context

  @website-ssl-provisioning-success
  Scenario: Website SSL provisioning succeeds
    Given the user has a website with DNS hosting enabled
    When the SSL status is updated to "issuing"
    And the SSL status is updated to "ready"
    Then the website SSL status is "ready"
    And the SSL certificate is issued

  @website-ssl-provisioning-pending
  Scenario: Website SSL status transitions through issuing
    Given the user has a website with DNS hosting enabled
    When the SSL status is updated to "issuing"
    Then the website SSL status is "issuing"

  @website-ssl-provisioning-fails
  Scenario: Website SSL provisioning fails
    Given the user has a website with DNS hosting enabled
    When the SSL status is updated to "issuing"
    And the SSL status update fails with error
    Then the website SSL status is "failed"
    And the SSL status contains error details

  @website-ssl-status-recovery
  Scenario: Website SSL status recovers from failure
    Given the user has a website with DNS hosting enabled
    When the SSL status is updated to "failed"
    And the SSL status is updated to "ready"
    Then the website SSL status is "ready"
