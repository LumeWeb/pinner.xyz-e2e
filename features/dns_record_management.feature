@dns-record-management
Feature: DNS Record Management

  Record management provides CRUD operations for DNS records within DNS zones.
  Records map DNS names to IP addresses or other DNS resources.

  Background:
    Given the user has an authenticated API key
    And the user has an IPFS client connection
    And the user has a DNS zone

  @dns-list-records
  Scenario: User lists all DNS records in a zone
    When the user lists DNS records
    Then the user receives a record list
    And the record list contains the record count

  @dns-get-record
  Scenario: User retrieves a specific DNS record
    Given the user has a DNS record "www" of type "A"
    When the user gets DNS record by name and type
    Then the DNS record is retrieved successfully

  @dns-create-record
  Scenario: User creates a new DNS record
    When the user creates an "A" record named "www" with value "192.0.2.1"
    Then the DNS record is created successfully
    And the record name is "www"
    And the record type is "A"

  @dns-update-record
  Scenario: User updates an existing DNS record
    Given the user has an "A" record named "www" with value "192.0.2.1"
    When the user updates the DNS record value to "192.0.2.2"
    Then the DNS record is updated successfully
    And the record value is "192.0.2.2"

  @dns-delete-record
  Scenario: User deletes a DNS record
    Given the user has a DNS record "www" of type "A"
    When the user deletes the DNS record
    Then the record cannot be retrieved
