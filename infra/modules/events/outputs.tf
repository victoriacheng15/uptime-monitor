output "rule_name" {
  description = "EventBridge rule name for scheduled uptime checks."
  value       = aws_cloudwatch_event_rule.scheduled_check.name
}
