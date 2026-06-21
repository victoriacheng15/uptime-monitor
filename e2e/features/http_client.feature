Feature: HTTP Polling Client
  As a website monitor checker
  I want the HTTP client to poll configured endpoints and correctly validate their status
  So that status changes in target services are accurately captured

  Scenario: Polling healthy and unhealthy endpoints
    Given the target endpoints are mocked to return:
      | path      | status |
      | /target-a | 200    |
      | /target-b | 500    |
    And the environment variable "MONITOR_TARGETS" is configured with target HTTP endpoints
    When the status check handler is executed
    Then the target HTTP endpoints "/target-a,/target-b" must be requested by the HTTP client

  Scenario: HTTP client handles redirect response codes as healthy status
    Given the target endpoints are mocked to return:
      | path      | status |
      | /redirect | 301    |
    And the environment variable "MONITOR_TARGETS" is configured with target HTTP endpoints
    When the status check handler is executed
    Then the target HTTP endpoints "/redirect" must be requested by the HTTP client
