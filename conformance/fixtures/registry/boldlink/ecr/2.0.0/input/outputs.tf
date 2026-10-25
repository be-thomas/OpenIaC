output "private_arn" {
  description = "Full ARN of the repository."
  value       = module.private.arn
}

output "private_registry_id" {
  description = "The registry ID where the repository was created."
  value       = module.private.registry_id
}

output "repository_url" {
  description = "The URL of the repository (in the form aws_account_id.dkr.ecr.region.amazonaws.com/repositoryName)."
  value       = module.private.repository_url
}

output "tags_all" {
  description = "A map of tags assigned to the resource, including those inherited from the provider default_tags"
  value       = module.private.tags_all
}

output "public_arn" {
  description = "Full ARN of the repository."
  value       = module.public.arn
}

output "id" {
  description = "The repository name."
  value       = module.public.id
}

output "public_registry_id" {
  description = "The registry ID where the repository was created."
  value       = module.public.registry_id
}

output "repository_uri" {
  description = "The URI of the repository."
  value       = module.public.repository_uri
}
