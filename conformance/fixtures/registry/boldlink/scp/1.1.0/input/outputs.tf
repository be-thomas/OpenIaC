output "id" {
  description = "The unique identifier (ID) of the policy."
  value       = aws_organizations_policy.main.id
}

output "arn" {
  description = "Amazon Resource Name (ARN) of the policy."
  value       = aws_organizations_policy.main.arn
}

output "tags_all" {
  description = "A map of tags assigned to the resource"
  value       = aws_organizations_policy.main.tags_all
}
