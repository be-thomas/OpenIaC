output "id" {
  description = "ARN of the certificate"
  value       = aws_acm_certificate.main.id
}

output "arn" {
  description = "ARN of the certificate"
  value       = aws_acm_certificate.main.arn
}

output "domain_name" {
  description = "Domain name for which the certificate is issued"
  value       = aws_acm_certificate.main.domain_name
}

output "domain_validation_options" {
  description = "Set of domain validation objects which can be used to complete certificate validation. Can have more than one element, e.g., if SANs are defined. Only set if DNS-validation was used."
  value       = aws_acm_certificate.main.domain_validation_options
}

output "not_after" {
  description = "Expiration date and time of the certificate."
  value       = aws_acm_certificate.main.not_after
}

output "not_before" {
  description = "Expiration date and time of the certificate."
  value       = aws_acm_certificate.main.not_before
}

output "pending_renewal" {
  description = "true if a Private certificate eligible for managed renewal is within the early_renewal_duration period."
  value       = aws_acm_certificate.main.pending_renewal
}

output "renewal_eligibility" {
  description = "Whether the certificate is eligible for managed renewal."
  value       = aws_acm_certificate.main.renewal_eligibility
}

output "renewal_summary" {
  description = " Contains information about the status of ACM's managed renewal for the certificate."
  value       = aws_acm_certificate.main.renewal_summary
}

output "status" {
  description = " Status of the certificate."
  value       = aws_acm_certificate.main.status
}

output "type" {
  description = "Source of the certificate."
  value       = aws_acm_certificate.main.type
}

output "tags_all" {
  description = "Map of tags assigned to the resource, including those inherited from the provider default_tags configuration block."
  value       = aws_acm_certificate.main.tags_all
}

output "validation_emails" {
  description = "List of addresses that received a validation email. Only set if EMAIL validation was used."
  value       = aws_acm_certificate.main.validation_emails
}

output "record_fqdn" {
  description = "FQDN built using the zone domain and name."
  value       = [for record in aws_route53_record.main : record.fqdn]
}
