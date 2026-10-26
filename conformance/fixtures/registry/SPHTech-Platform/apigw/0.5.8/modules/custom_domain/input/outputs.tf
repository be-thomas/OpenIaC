output "arn" {
  value       = var.endpoint_type == "EDGE" ? aws_api_gateway_domain_name.domain_edge[0].arn : aws_api_gateway_domain_name.domain_regional[0].arn
  description = "ARN of domain name."
}

output "domain_name" {
  value       = var.endpoint_type == "EDGE" ? aws_api_gateway_domain_name.domain_edge[0].cloudfront_domain_name : aws_api_gateway_domain_name.domain_regional[0].regional_domain_name
  description = "hostname for custom domain regional/cloudfront endpoint"
}

output "hosted_zone_id" {
  value       = var.endpoint_type == "EDGE" ? aws_api_gateway_domain_name.domain_edge[0].cloudfront_zone_id : aws_api_gateway_domain_name.domain_regional[0].regional_zone_id
  description = "hosted zone id of custom domain"
}
