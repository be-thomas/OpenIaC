output "zone_id" {
  description = "The ID of the hosted zone to contain this record."
  value       = aws_route53_record.main.zone_id
}

output "name" {
  description = "The name of the record."
  value       = aws_route53_record.main.name
}

output "type" {
  description = "The record type."
  value       = aws_route53_record.main.type
}

output "ttl" {
  description = "The TTL of the record."
  value       = aws_route53_record.main.ttl
}

output "fqdn" {
  description = "FQDN built using the zone domain and name."
  value       = aws_route53_record.main.fqdn
}
