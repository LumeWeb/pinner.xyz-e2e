@website-management
Feature: Website Management

  Website management provides CRUD operations for websites in the IPFS Portal.
  Websites unify IPFS content, DNS hosting, and SSL certificate management into a single
  service. Setting dns_hosting_enabled to false keeps websites in pure IPFS mode without
  auto-conversion to IPNS.

  Background:
    Given the admin is authenticated
    And the admin creates a new quota plan named "Default Test Plan"
    And the admin sets the plan as default
    And the user has an authenticated API key
    And the user has an IPFS client connection
    And the DNS dev server is available

  @website-list-websites
  Scenario: User lists all websites
    When the user lists websites
    Then the website count is returned

  @website-get-website
  Scenario: User retrieves a specific website by ID
    Given the user has a website
    When the user gets the website by ID
    Then the website exists

  @website-create-website-pure-ipfs
  Scenario: User creates a website with pure IPFS target (DNS hosting disabled)
    Given we upload a website bundle
    And the upload operation completes
    When the user creates a pure IPFS website with domain "test-website.example.com"
    Then the website exists
    And the website domain matches "test-website.example.com"
    And the website target type is "ipfs"
    And the website has a target hash

  @website-create-website-ipfs-with-dns
  Scenario: User creates a website with IPFS CID target (DNS hosting enabled, auto-converts to IPNS)
    Given we upload a website bundle
    And the upload operation completes
    When the user creates a website with domain "test-website-dns.example.com" using the IPFS CID
    Then the website exists
    And the website domain matches "test-website-dns.example.com"
    And the website target type is "ipns"
    And the website has a target hash

  @website-create-website-ipns
  Scenario: User creates a website with IPNS target
    Given the user has an IPNS key
    When the user creates a website with domain "test-website-ipns.example.com" using the IPNS peer ID
    Then the website exists
    And the website domain matches "test-website-ipns.example.com"
    And the website target type is "ipns"
    And the website has a target hash

  @website-update-website
  Scenario: User updates an existing website
    Given the user has a website
    When the user updates the website
    Then the website is updated successfully

  @website-delete-website
  Scenario: User deletes a website
    Given the user has a website
    When the user deletes the website
    Then the website does not exist
