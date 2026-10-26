output "id" {
  description = "ARN of the secret."
  value       = aws_secretsmanager_secret.this.id
}

output "arn" {
  description = "ARN of the secret."
  value       = aws_secretsmanager_secret.this.arn
}

output "replica" {
  description = "Date that you last accessed the secret in the Region."
  value       = aws_secretsmanager_secret.this.replica
}

output "tags_all" {
  description = "Map of tags assigned to the resource, including those inherited from the provider default_tags"
  value       = aws_secretsmanager_secret.this.tags_all
}
