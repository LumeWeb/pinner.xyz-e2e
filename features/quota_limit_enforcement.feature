Feature: Quota Limit Enforcement
  This feature tests that operations are correctly denied when quota limits
  are exhausted. It ensures the EnableQuotaEnforcement flag properly blocks
  uploads, downloads, and content pinning operations.

  Background:
    Given the admin is authenticated
    And the admin creates a new quota plan named "Limited Test Plan"
    And the admin sets the plan as default
    And an existing registered user
    And the user is logged in
    And the user records their initial quota status

  @quota-storage-limit-exceeded
  Scenario: Content pinning is denied when storage limit is exceeded
    Given the admin sets the upload total limit to 1 GB
    And the admin sets the total download limit to 1 GB
    And the user has a 5MB file with unique content
    And the admin sets the storage limit to match the pending upload DAG size
    And the user records their initial quota status
    When the user uploads the IPFS file
    And the operation completes
    When the user pins the content to IPFS
    And the IPFS pin reaches pinned status within 5 minutes
    When the user attempts to upload another 5MB file to IPFS
    Then the upload operation is denied with a quota exceeded error

  @quota-download-total-limit-exceeded
  Scenario: HTTP download operation denied when total limit is exceeded
    Given the admin creates a new quota plan named "Total Limit Test Plan"
    And the admin sets the upload total limit to 1 GB
    And the admin sets the total download limit to 10 MB
    And the admin sets the storage limit to 1 GB
    And the admin sets the plan as default
    And the admin assigns the current plan to the authenticated user
    And the user records their initial quota status
    And the user uploads a 10MB file to IPFS
    And the operation completes
    And the IPFS pin reaches pinned status within 5 minutes
    And the admin sets the download limit to the uploaded file size
    When the user downloads the file via HTTP from IPFS gateway
    When the user attempts to download the file via HTTP from IPFS gateway
    Then the download operation is denied with a quota exceeded error

  @quota-multiple-limits-exceeded
  Scenario: Operations denied when multiple quota types are exhausted
    Given the admin sets the upload total limit to 5 MB
    And the admin sets the storage limit to 1 GB
    And the user has a 5MB file with unique content
    And the admin sets the upload total limit to match the pending upload DAG size
    When the user uploads the IPFS file
    And the operation completes
    When the user uploads another 5MB file to IPFS
    Then the upload operation is denied with a quota exceeded error
    When the user downloads the file via HTTP from IPFS gateway
    Then the download operation is denied with a quota exceeded error

  @quota-pin-denied-storage-exceeded
  Scenario: Network IPFS pin operation denied when storage limit is exceeded
    Given the admin sets the upload total limit to 1 GB
    And the admin sets the total download limit to 1 GB
    And the admin sets the storage limit to 5 MB
    And the user records their initial quota status
    And the user has an existing IPFS CID from the network that is 10MB
    When the user attempts to pin the IPFS CID
    Then the operation fails

  @quota-download-channel-denial
  Scenario: All download channels are denied when download quota is exhausted
    Given the admin sets the upload total limit to 1 GB
    And the admin sets the storage limit to 1 GB
    And the user records their initial quota status
    When the user uploads a 10MB file to IPFS
    And the operation completes
    And the IPFS pin reaches pinned status within 5 minutes
    And the admin sets the download limit to the uploaded file size
    When the user downloads the file via HTTP from IPFS gateway
    And the user checks their download quota
    And the download quota has reached 100% of the limit
    When the user attempts to download the file via HTTP from IPFS gateway
    Then the download operation is denied with a quota exceeded error
    When the user attempts to fetch content via IPFS via P2P network and fails


