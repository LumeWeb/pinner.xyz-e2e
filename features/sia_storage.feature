Feature: Sia Storage

  As an end-user of the LumeWeb Portal
  I want to store and retrieve files on Sia
  So that my data is persisted on the decentralized storage network

  Background:
    Given the admin is authenticated
    And the admin creates a new quota plan named "Default Test Plan"
    And the admin sets the plan as default
    And an existing registered user
    And the user is logged in
    And the user has connected their Sia account
    And the SIA account is ready

  @sia-upload-small-file
  Scenario: User uploads a small file to Sia
    Given the user has a small Sia test file
    When the user uploads the file to Sia
    Then the SIA file is stored successfully

  @sia-download-file
  Scenario: User downloads a stored file from Sia
    Given the user has a stored Sia file
    When the user downloads the file from Sia
    Then the downloaded Sia content matches the original

  @sia-view-file-details
  Scenario: User views details of a stored file
    Given the user has a stored Sia file
    When the user views the Sia file details
    Then the SIA file information is displayed

  @sia-delete-file
  Scenario: User deletes a stored file from Sia
    Given the user has a stored Sia file
    When the user removes the file from Sia
    Then the SIA file is no longer listed

  @sia-pin-file
  Scenario: User pins a file to keep it available
    Given the user has uploaded a file to Sia
    When the user pins the SIA file
    Then the SIA file reaches pinned status

  @sia-unpin-file
  Scenario: User unpins and prunes a stored file
    Given the user has a pinned Sia file
    When the user unpins the SIA file
    And the user prunes unpinned Sia files
    Then the SIA file list reflects the changes

  @sia-file-metadata
  Scenario: User views metadata of a pinned file
    Given the user has a pinned Sia file
    When the user views the Sia file metadata
    Then the SIA file metadata is complete

  @sia-store-metadata
  Scenario: User stores and retrieves custom metadata
    Given the user uploads a file with custom SIA metadata
    When the user retrieves the SIA object metadata
    Then the SIA object metadata matches the stored metadata
