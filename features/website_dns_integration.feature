@website-dns-integration
Feature: Website DNS Integration

  Website DNS integration tests verify that websites can leverage DNS hosting
  capabilities, either automatically or by linking to existing DNS zones.

  Background:
    Given the user has an authenticated API key
    And the user has an IPFS client connection
    And the DNS dev server is available

  @website-dns-hosting-enabled
  Scenario: User creates a website with DNS hosting enabled
    When the user creates a website with DNS hosting enabled
    Then DNS hosting is enabled for the website
    And the website exists
    When the user validates the DNS records
    Then the website becomes active

  @website-dns-hosting-disabled
  Scenario: User creates a website without DNS hosting
    When the user creates a website with domain "test-website-nodns.example.com"
    Then DNS hosting is disabled for the website
    When the user validates the DNS records
    Then the website becomes active

  @website-dns-update-hosting
  Scenario: User updates website to enable DNS hosting
    Given the user has a website without DNS hosting
    When the user updates the website to enable DNS hosting
    Then DNS hosting is enabled for the website
    When the user validates the DNS records
    Then the website becomes active

  @website-dns-zone-linkage
  Scenario: Website links to existing DNS zone
    Given the user has a DNS zone
    When the user creates a website with DNS zone linking
    Then the website DNS zone ID is stored in context
    And DNS hosting is enabled for the website
    When the user validates the DNS records
    Then the website becomes active
