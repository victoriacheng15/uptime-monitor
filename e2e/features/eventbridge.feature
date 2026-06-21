Feature: EventBridge Cron Scheduled Trigger
  As a serverless runtime
  I want scheduled triggers from EventBridge to invoke target fleet checks
  So that monitoring coverage is maintained automatically without user intervention

  Scenario: Scheduled EventBridge trigger invokes the checks handler
    When an EventBridge scheduled trigger event is received
    Then the status check handler runs and returns HTTP status code 200

  Scenario: Missing monitoring targets prevents check execution
    Given the environment variable "MONITOR_TARGETS" is unset
    When an EventBridge scheduled trigger event is received
    Then the status check handler runs and returns HTTP status code 500
