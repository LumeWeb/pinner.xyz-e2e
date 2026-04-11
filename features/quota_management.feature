Feature: Account Quota Management
  This feature tests quota consumption through actual service operations (upload/download files).
  Quota is only consumed when using portal services like IPFS upload, download, and pinning.

  Background:
    Given the admin is authenticated
    And the admin creates a new quota plan named "Default Test Plan"
    And the admin sets the plan as default
    And an existing registered user
    And the user is logged in
    And the user records their initial quota status

  @quota-upload-consumption
  Scenario: User uploads a file and upload quota is consumed
    When the user uploads a 10MB file to IPFS
    And the user checks their upload quota
    Then the upload quota usage has increased by approximately 10MB

  @quota-download-consumption
  Scenario: User downloads a file and download quota is consumed
    Given the user has uploaded a 5MB file with CID
    When the user downloads the file from IPFS
    And the user checks their download quota
    Then the download quota usage has increased by approximately 5MB

  @quota-storage-consumption
  Scenario: User pins content and storage quota is consumed
    Given the user has uploaded a 15MB file with CID
    When the user pins the content to IPFS
    And the user checks their storage quota
    Then the storage quota usage has increased by approximately 15MB

  @quota-bandwidth-tracking
  Scenario: Quota bandwidth is tracked with upload and download
    Given the user records their initial bandwidth status
    When the user uploads a 10MB file to IPFS
    And the user downloads the same file from IPFS
    And the user checks their bandwidth quota
    Then the bandwidth quota reflects both upload and download usage

  @quota-percentage-calculation
  Scenario: Quota percentage calculations are accurate
    When the user checks their account quota
    And the user uploads a 10MB file to IPFS
    And the user checks their account quota
    Then the upload quota percentage calculation matches the new usage

  @quota-history-tracking
  Scenario: User can retrieve historical quota usage data
    When the user uploads a 10MB file to IPFS
    And the user retrieves quota history for the last 24 hours
    Then the quota history shows usage data points

  @quota-download-via-ipfs-network
  Scenario: User fetches content via IPFS network and download quota is consumed
    Given the user has uploaded a 5MB file with CID
    And the user pins the content to IPFS
    And the operation completes
    When the user fetches content via IPFS network
    And the user checks their download quota
    Then the download quota usage has increased by approximately 5MB
