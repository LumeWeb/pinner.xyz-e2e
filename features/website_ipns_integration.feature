@website-ipns-integration
Feature: Website IPNS Integration

  Website IPNS integration tests verify that websites can use IPNS names
  as targets instead of direct CIDs, allowing for content updates without
  changing the website configuration.

  Background:
    Given the user has an authenticated API key
    And the user has an IPFS client connection
    And the DNS dev server is available

  @website-ipns-target
  Scenario: User creates a website with IPNS target
    Given we upload a website bundle
    And the upload operation completes
    And the user creates an IPNS key
    And the user publishes the uploaded CID to the IPNS key
    When the user has an IPNS website
    Then the website exists
    And the website target type is "ipns"
    And the website has an IPNS target
    When the janitor validates the website
    Then the website becomes active

  @website-ipns-update-content
  Scenario: Website with IPNS target can be updated without changing config
    Given the user has an IPNS website
    When we upload a website bundle
    And the upload operation completes
    And the user publishes the uploaded CID to the existing IPNS key
    Then the website target hash is the IPNS name
    And the website target type is still "ipns"
    When the janitor validates the website
    Then the resolved CID is pinned
    And the website becomes active

  @website-convert-ipfs-to-ipns
  Scenario: User converts website from IPFS to IPNS target
    Given the user has a website
    And the user has an IPNS key
    When the user publishes the website CID to the IPNS key
    And the user updates the website to use the IPNS target
    Then the website target type is "ipns"
    And the website target hash is the IPNS name
    When the janitor validates the website
    Then the website becomes active
