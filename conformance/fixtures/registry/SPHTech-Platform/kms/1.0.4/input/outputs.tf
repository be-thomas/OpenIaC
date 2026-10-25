output "key_arn" {
  description = "ARN of the key created"
  value       = aws_kms_key.key.arn
}

output "key_id" {
  description = "Created key ID"
  value       = aws_kms_key.key.key_id
}

output "key_alias" {
  description = "Created key alias"
  value       = var.alias != null && var.alias != "" ? aws_kms_alias.key[0].name : null
}
