Feature: IPFS Upload

  As an end-user of the LumeWeb Portal
  I want to upload files to IPFS
  So that I can store and share my content on the decentralized web

  Background:
    Given an existing registered user
    And the user is logged in

  @ipfs-upload-small-file
  Scenario: User uploads a small file to IPFS
    When the user uploads a small file to IPFS "test.txt" with content "Hello, IPFS!"
    And the IPFS pin reaches pinned status
    And the operation completes
    Then a valid IPFS CID is returned

  @ipfs-upload-large-file-tus
  Scenario: User uploads a large file to IPFS
    When the user uploads a 100MB file to IPFS
    And the IPFS pin reaches pinned status
    And the operation completes
    Then the file is available on IPFS

  @ipfs-upload-directory
  Scenario: User uploads a directory structure to IPFS
    Given the user has an IPFS directory with multiple files
    When the user uploads the IPFS directory
    And the IPFS pin reaches pinned status
    And the operation completes
    Then all files are uploaded as an IPFS directory CID
    And the directory structure is preserved

  @ipfs-concurrent-upload-stress
  Scenario: Multiple concurrent uploads complete successfully
    When the user starts 10 concurrent file uploads
    And the IPFS pin reaches pinned status
    And the operation completes
    Then all 10 files are available

  @ipfs-upload-content-integrity
  Scenario: Uploaded content matches original
    Given the user has a file with known content
    When the user uploads the file
    And the IPFS pin reaches pinned status
    And the operation completes
    Then the retrieved content matches original

  @ipfs-large-file-integrity-tus
  Scenario: Large file uploaded via TUS maintains integrity
    Given the user has a 200MB file with unique content
    When the user uploads the file via TUS
    And the IPFS pin reaches pinned status
    And the operation completes
    Then the retrieved file CID matches original

