@dns-zone-management
Feature: DNS Zone Management

  Zone management provides CRUD operations for DNS zones in the IPFS Portal.
  Zones represent DNS domains that can contain DNS records.

  Background:
    Given the user has an authenticated API key
    And the user has an IPFS client connection

  @dns-list-zones
  Scenario: User lists all DNS zones
    When the user lists DNS zones
    Then the user receives a zone list
    And the zone list contains the zone count

  @dns-get-zone
  Scenario: User retrieves a specific DNS zone by ID
    Given the user has a DNS zone
    When the user gets DNS zone by ID
    Then the DNS zone is retrieved successfully

  @dns-create-zone
  Scenario: User creates a new DNS zone
    When the user creates a DNS zone with domain "test.example.com"
    Then the DNS zone is created successfully
    And the zone domain is stored in context

  @dns-delete-zone
  Scenario: User deletes a DNS zone
    Given the user has a DNS zone
    When the user deletes the DNS zone
    Then the zone cannot be retrieved
