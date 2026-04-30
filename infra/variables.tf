variable "aws_region" {
  description = "AWS region for uptime monitor infrastructure."
  type        = string
  default     = "ca-central-1"
}

variable "project_name" {
  description = "Project name used for IAM resource names and tags."
  type        = string
  default     = "uptime-monitor"
}

variable "history_bucket_name" {
  description = "S3 bucket name for uptime monitor latest and history JSON."
  type        = string
  default     = "uptime-monitor-history"
}

variable "tfstate_bucket_name" {
  description = "S3 bucket name for OpenTofu state."
  type        = string
  default     = "uptime-monitor-tfstate"
}

variable "lambda_function_name" {
  description = "Lambda function name for the uptime monitor backend."
  type        = string
  default     = "uptime-monitor"
}
