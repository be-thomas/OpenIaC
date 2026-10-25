output "aws_api_gateway_rest_api_id" {
  value       = aws_api_gateway_rest_api.api.id
  description = "REST API id of the created api"
}

output "aws_api_gateway_execution_arn" {
  value       = aws_api_gateway_rest_api.api.execution_arn
  description = "The execution ARN part to be used in lambda_permission source_arn when allowing API Gateway to invoke a Lambda function"
}

output "aws_api_gateway_stage_name" {
  value       = aws_api_gateway_stage.stage.stage_name
  description = "Stage name of the deployed api gateway stage"
}

output "aws_api_gateway_log_group_arn" {
  value       = aws_cloudwatch_log_group.log_group.arn
  description = "ARN of the CloudWatch log group for API Gateway access logs"
}

output "aws_api_gateway_stage_execution_arn" {
  value       = aws_api_gateway_stage.stage.execution_arn
  description = "The execution ARN part to be used in lambda_permission source_arn when allowing API Gateway to invoke a Lambda function"
}

output "aws_api_gateway_stage_invoke_url" {
  value       = aws_api_gateway_stage.stage.invoke_url
  description = "The URL to invoke the API pointing to the stage"
}

output "aws_api_gateway_vpc_link_id" {
  value       = { for k, v in values(aws_api_gateway_vpc_link.vpc_link)[*] : v.name => v.id }
  description = "VPC link v1 IDs keyed by VPC link name"
}

output "aws_apigatewayv2_vpc_link_id" {
  value       = { for v in values(aws_apigatewayv2_vpc_link.vpc_link) : v.name => v.id }
  description = "VPC link v2 IDs keyed by VPC link name"
}

output "aws_api_gateway_rest_api_root_resource_id" {
  value       = aws_api_gateway_rest_api.api.root_resource_id
  description = "Root resource ID of the API Gateway REST API"
}

output "aws_api_gateway_stage_arn" {
  value       = aws_api_gateway_stage.stage.arn
  description = "ARN (for management) of apigw stage, use this when attaching Web ACL"
}

output "body_spec" {
  value       = local.body_spec
  description = "The computed OpenAPI body specification after merging resource policy if enabled."
}
