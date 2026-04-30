output "history_bucket_name" {
  description = "S3 bucket used for uptime monitor runtime data."
  value       = module.s3.history_bucket_name
}

output "tfstate_bucket_name" {
  description = "S3 bucket intended for OpenTofu remote state."
  value       = module.s3.tfstate_bucket_name
}

output "lambda_function_name" {
  description = "Uptime monitor Lambda function name."
  value       = module.lambda.function_name
}

output "lambda_function_url" {
  description = "Public Lambda Function URL for the uptime monitor backend."
  value       = module.lambda.function_url
}
