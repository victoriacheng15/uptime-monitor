resource "aws_cloudwatch_event_rule" "scheduled_check" {
  name                = "${var.function_name}-scheduled-check"
  description         = "Runs the uptime monitor check on a schedule."
  schedule_expression = var.schedule_expression
}

resource "aws_cloudwatch_event_target" "scheduled_check" {
  rule      = aws_cloudwatch_event_rule.scheduled_check.name
  target_id = "${var.function_name}-check"
  arn       = var.function_arn

  input = jsonencode({
    source = "uptime-monitor.schedule"
    action = "check"
  })
}

resource "aws_lambda_permission" "eventbridge" {
  statement_id  = "AllowEventBridgeScheduledCheck"
  action        = "lambda:InvokeFunction"
  function_name = var.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.scheduled_check.arn
}
