@dns-bulk-operations
Feature: DNS Bulk Operations

  Bulk operations provide efficient ways to create or delete multiple DNS records
  in a single API call, improving performance for batch updates.

  Background:
    Given the user has an authenticated API key
    And the user has an IPFS client connection
    And the user has a DNS zone

  @dns-bulk-create-records
  Scenario: User creates multiple DNS records in bulk
    When the user creates "A" records for "www", "api", and "app" with value "192.0.2.1"
    And the DNS record count is 3

  @dns-bulk-delete-records
  Scenario: User deletes multiple DNS records in bulk
    Given the user has "A" records for "www", "api", and "app"
    When the user deletes DNS records for "www", "api", and "app"
    Then all DNS records are deleted successfully
    And the DNS record count is 0
