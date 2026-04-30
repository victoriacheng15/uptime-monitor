variable "history_bucket_name" {
  description = "S3 bucket name for uptime monitor latest and history JSON."
  type        = string
}

variable "tfstate_bucket_name" {
  description = "S3 bucket name for OpenTofu state."
  type        = string
}
