Feature: IPFS Content List and Filtering

  As an end-user of the LumeWeb Portal
  I want to list and filter my IPFS content
  So that I can find and manage my files

  Background:
    Given the admin is authenticated
    And the quota plan is set up
    And an existing registered user
    And the user is logged in

  @ipfs-list-all-content
  Scenario: User lists all their IPFS content
    Given the user has 5 uploaded IPFS files
    When the user lists their IPFS content
    Then all 5 IPFS files are returned

  @ipfs-filter-content-by-name
  Scenario: User filters IPFS content by name
    Given the user has uploaded IPFS files named "test1.txt", "test2.txt", "example.txt"
    When the user filters IPFS content by name "test"
    Then only matching items are returned

  @ipfs-get-content-status
  Scenario: User checks IPFS content status
    Given the user has uploaded IPFS content
    When the user checks IPFS content status
    Then the status is returned successfully
