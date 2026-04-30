output "history_bucket_name" {
  description = "Runtime data bucket name."
  value       = aws_s3_bucket.history.bucket
}

output "history_bucket_arn" {
  description = "Runtime data bucket ARN."
  value       = aws_s3_bucket.history.arn
}

output "tfstate_bucket_name" {
  description = "OpenTofu state bucket name."
  value       = aws_s3_bucket.tfstate.bucket
}

output "tfstate_bucket_arn" {
  description = "OpenTofu state bucket ARN."
  value       = aws_s3_bucket.tfstate.arn
}
