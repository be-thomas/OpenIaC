output "arn" {
  description = "The arn of the Lambda function"
  value       = aws_lambda_function.function.arn
}

output "invoke_arn" {
  description = "The invocation arn of this lambda function"
  value       = aws_lambda_function.function.invoke_arn
}

output "role_arn" {
  description = "The IAM role arn of this lambda function"
  value       = aws_iam_role.function.arn
}


output "invoke_policy_arn" {
  description = "The arn of the invocation policy for this Lambda function"
  value       = aws_iam_policy.invoke_function.arn
}

output "name" {
  description = "The name of the Lambda function"
  value       = aws_lambda_function.function.function_name
}

output "qualified_arn" {
  description = "The qualified arn of the Lambda function"
  value       = aws_lambda_function.function.qualified_arn
}

output "qualified_invoke_arn" {
  description = "The qualified invocation arn of the lambda function"
  value       = "arn:aws:apigateway:${data.aws_region.current.name}:lambda:path/2015-03-31/functions/${aws_lambda_function.function.qualified_arn}/invocations"
}

output "log_group_name" {
  description = "the Cloudwatch Log group name"
  value       = aws_cloudwatch_log_group.log_group.name
}

output "version" {
  description = "The version of the Lambda function"
  value       = aws_lambda_function.function.version
}
