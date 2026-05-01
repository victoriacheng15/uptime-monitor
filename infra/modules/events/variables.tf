variable "function_name" {
  description = "Lambda function name invoked by the scheduled check."
  type        = string
}

variable "function_arn" {
  description = "Lambda function ARN invoked by the scheduled check."
  type        = string
}

variable "schedule_expression" {
  description = "EventBridge schedule expression for uptime checks."
  type        = string
}
