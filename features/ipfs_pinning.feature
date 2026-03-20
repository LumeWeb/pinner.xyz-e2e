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
    And the operation completes
    And the IPFS pin reaches pinned status
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
    And the operation completes
    And the IPFS pin reaches pinned status
    Then the old CID is no longer pinned




