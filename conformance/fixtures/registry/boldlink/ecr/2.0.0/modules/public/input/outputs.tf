output "arn" {
  description = "Full ARN of the repository."
  value       = aws_ecrpublic_repository.this.*.arn
}

output "id" {
  description = "The repository name."
  value       = aws_ecrpublic_repository.this.*.id
}

output "registry_id" {
  description = "The registry ID where the repository was created."
  value       = aws_ecrpublic_repository.this.*.registry_id
}

output "repository_uri" {
  description = "The URI of the repository."
  value       = aws_ecrpublic_repository.this.*.repository_uri
}
