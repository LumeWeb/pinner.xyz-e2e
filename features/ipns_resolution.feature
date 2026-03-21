Feature: IPNS Resolution

  As an end-user of the LumeWeb Portal
  I want to resolve IPNS names to CIDs
  So that I can find the current content published under a consistent name

  Background:
    Given an existing registered user
    And the user is logged in

  @ipns-resolve-kubo-published
  Scenario: User resolves IPNS name published via Kubo
    Given the user has an IPNS CID "bafkreihsg6m2x7u2g5v7ffufx3eb5vkcjaxqtohnwxizovtcqnpm3xtp2m" in Kubo
    And the CID is published to the IPNS key "blog" via Kubo
    When the user resolves IPNS name published by Kubo via Portal
    Then the resolved CID is "bafkreihsg6m2x7u2g5v7ffufx3eb5vkcjaxqtohnwxizovtcqnpm3xtp2m"

  @ipns-create-key-portal-resolve
  Scenario: User creates IPNS key and resolves via Portal
    Given the user has an IPNS key named "myblog" in the Portal
    And the user has an IPNS CID "bafkreihsg6m2x7u2g5v7ffufx3eb5vkcjaxqtohnwxizovtcqnpm3xtp2m"
    And the user publishes CID "bafkreihsg6m2x7u2g5v7ffufx3eb5vkcjaxqtohnwxizovtcqnpm3xtp2m" to the IPNS key
    When the user resolves the IPNS name via Portal
    Then the IPNS name exists in the Portal
