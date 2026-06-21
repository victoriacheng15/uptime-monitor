Feature: S3 Snapshot Persistence
  As an S3 storage client
  I want check results to be correctly serialized and stored in S3
  So that site history can be read by the static site generator

  Scenario: Writing first status snapshots to empty S3 bucket
    Given a set of check results:
      | url                      | status | is_up |
      | https://site-a.com/check | 200    | true  |
      | https://site-b.com/check | 500    | false |
    And the S3 storage is mocked and empty
    When the check results are persisted to storage
    Then the mock S3 storage must contain "latest.json" and "history.json" matching the check results

  Scenario: Appending results to existing rolling history in S3
    Given the mock S3 storage contains an existing history for "https://site-a.com/check" with 5 entries
    And a set of check results:
      | url                      | status | is_up |
      | https://site-a.com/check | 200    | true  |
    When the check results are persisted to storage
    Then the mock S3 storage must contain "latest.json" and "history.json" matching the check results
    And the history entries for "https://site-a.com/check" must be trimmed to keep only 5 entries
