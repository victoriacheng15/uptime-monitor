output "history_bucket_name" {
  description = "S3 bucket used for uptime monitor runtime data."
  value       = module.s3.history_bucket_name
}

output "tfstate_bucket_name" {
  description = "S3 bucket intended for OpenTofu remote state."
  value       = module.s3.tfstate_bucket_name
}
