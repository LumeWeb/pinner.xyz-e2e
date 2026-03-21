Feature: IPNS Publishing

  As an end-user of the LumeWeb Portal
  I want to publish content to IPNS
  So that I can use consistent names to reference content that may change over time

  Background:
    Given an existing registered user
    And the user is logged in

  @ipns-publish-cid
  Scenario: User publishes a CID to IPNS
    Given the user has an IPNS key named "website"
    When the user publishes CID "bafkreiggnartitziefjx2kvf555ku6nj7mnhvo7tyli4w6bx32wa57tak4" to the IPNS key
    Then the IPNS name is resolvable by Portal
    And Kubo can resolve the IPNS name
    And both resolve to "bafkreiggnartitziefjx2kvf555ku6nj7mnhvo7tyli4w6bx32wa57tak4"

  @ipns-republish-all
  Scenario: User republishes all IPNS entries
    Given the user has an IPNS key named "test-key"
    When the user republishes all IPNS entries
    Then the republish operation succeeds
