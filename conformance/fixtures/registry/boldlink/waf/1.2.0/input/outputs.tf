output "arn" {
  description = "The ARN of the WAF WebACL."
  value       = aws_wafv2_web_acl.main.arn
}

output "capacity" {
  description = "Web ACL capacity units (WCUs) currently being used by this web ACL."
  value       = aws_wafv2_web_acl.main.capacity
}

output "id" {
  description = "The ID of the WAF WebACL."
  value       = aws_wafv2_web_acl.main.id
}

output "tags_all" {
  description = "Map of tags assigned to the resource, including those inherited from the provider default_tags configuration block."
  value       = aws_wafv2_web_acl.main.tags_all
}

output "ipv4_set_id" {
  description = "A unique identifier for the IPv4 set."
  value       = try(aws_wafv2_ip_set.ipset_v4[0].id, null)
}

output "ipv4_set_arn" {
  description = "The Amazon Resource Name (ARN) of the IPv4 set."
  value       = try(aws_wafv2_ip_set.ipset_v4[0].arn, null)
}

output "ipv6_set_id" {
  description = "A unique identifier for the IPv6 set."
  value       = try(aws_wafv2_ip_set.ipset_v6[0].id, null)
}

output "ipv6_set_arn" {
  description = "The Amazon Resource Name (ARN) of the IPv6 set."
  value       = try(aws_wafv2_ip_set.ipset_v6[0].arn, null)
}

output "log_destination_configs" {
  description = "The logging destination configuration for the WebACL."
  value       = try(aws_wafv2_web_acl_logging_configuration.main[0].log_destination_configs, null)
}
