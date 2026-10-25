output "arn" {
  description = "The Amazon Resource Name (ARN) of the Hosted Zone."
  value       = aws_route53_zone.main.arn
}

output "zone_id" {
  description = "The Hosted Zone ID. This can be referenced by zone records."
  value       = aws_route53_zone.main.zone_id
}

output "name_servers" {
  description = " A list of name servers in associated (or default) delegation set."
  value       = aws_route53_zone.main.name_servers
}

output "tags_all" {
  description = "A map of tags assigned to the resource, including those inherited from the provider default_tags"
  value       = aws_route53_zone.main.tags_all
}
