variable "function_name" {
  description = "Lambda function name."
  type        = string
}

variable "package_file" {
  description = "Path to the Lambda deployment package zip."
  type        = string
}

variable "history_bucket_arn" {
  description = "ARN of the S3 bucket storing runtime monitor data."
  type        = string
}

variable "history_bucket_name" {
  description = "Name of the S3 bucket storing runtime monitor data."
  type        = string
}
