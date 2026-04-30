output "function_name" {
  description = "Lambda function name."
  value       = aws_lambda_function.backend.function_name
}

output "function_arn" {
  description = "Lambda function ARN."
  value       = aws_lambda_function.backend.arn
}

output "function_url" {
  description = "Lambda Function URL."
  value       = aws_lambda_function_url.backend.function_url
}
