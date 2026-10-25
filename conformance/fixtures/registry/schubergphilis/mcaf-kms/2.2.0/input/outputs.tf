output "arn" {
  value       = aws_kms_key.default.arn
  description = "ARN of the key"
}

output "id" {
  value       = aws_kms_key.default.key_id
  description = "ID of the key"
}

output "policy" {
  value       = local.policy
  description = "The key policy applied to the key, as a JSON-encoded string"
}
