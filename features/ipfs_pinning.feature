Feature: IPFS Pinning

  As an end-user of the LumeWeb Portal
  I want to pin content on IPFS
  So that my content remains available on the network

  Background:
    Given an existing registered user
    And the user is logged in

  @ipfs-pin-existing-cid
  Scenario: User pins an existing IPFS CID
    Given the user has an IPFS CID
    When the user starts pinning the IPFS CID
    And the IPFS pin reaches pinned status
    And the operation completes
    Then the IPFS pin is created successfully

  @ipfs-list-pins
  Scenario: User lists their IPFS pins
    Given the user has 3 pinned IPFS CIDs
    When the user lists their IPFS pins
    Then all 3 IPFS pins are returned
    And each pin has a unique CID

  @ipfs-unpin-content
  Scenario: User unpins IPFS content
    Given the user has a pinned IPFS CID
    When the user unpins the IPFS content
    Then the IPFS pin is no longer in the list

  @ipfs-replace-pin
  Scenario: User replaces an IPFS pin with new content
    Given the user has a pinned IPFS CID
    When the new IPFS CID is pinned
    And the IPFS pin reaches pinned status
    And the operation completes
    Then the old CID is no longer pinned

  @ipfs-pin-very-large
  Scenario: Pin operations complete within reasonable time for large content
    Given the user has a 3GB IPFS test file
    When the user uploads and pins the large IPFS test file
    And the IPFS pin reaches pinned status within 5 minutes
    Then the uploaded IPFS test file is available


