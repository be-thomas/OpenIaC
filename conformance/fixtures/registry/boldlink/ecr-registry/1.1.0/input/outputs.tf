output "ecr_scanning_configuration_scan_type" {
  description = "The scan type for the ECR registry scanning configuration (if defined)."
  value       = length(var.scanning_configuration) > 0 ? aws_ecr_registry_scanning_configuration.this[0].scan_type : null
}
output "ecr_pull_through_cache_rule" {
  description = "The details of the ECR pull-through cache rule (if enabled)."
  value       = var.enable_pull_through_cache_rule ? aws_ecr_pull_through_cache_rule.this[0] : null
}

output "ecr_replication_configuration" {
  description = "The replication configuration details for ECR (if defined)."
  value       = length(var.replication_configuration) > 0 ? aws_ecr_replication_configuration.this[0].replication_configuration : null
}

output "ecr_repository_name" {
  description = "The name of the created ECR repository (if created)."
  value       = var.repository_name != "" ? aws_ecr_repository.this[0].name : null
}

output "ecr_repository_arn" {
  description = "The ARN of the created ECR repository (if created)."
  value       = var.repository_name != "" ? aws_ecr_repository.this[0].arn : null
}

output "ecr_repository_url" {
  description = "The URL of the created ECR repository (if created)."
  value       = var.repository_name != "" ? aws_ecr_repository.this[0].repository_url : null
}

output "ecr_repository_image_tag_mutability" {
  description = "The image tag mutability setting for the created repository (if created)."
  value       = var.repository_name != "" ? aws_ecr_repository.this[0].image_tag_mutability : null
}

output "ecr_repository_encryption_configuration" {
  description = "The encryption configuration of the created repository (if created)."
  value       = var.repository_name != "" ? aws_ecr_repository.this[0].encryption_configuration : null
}

output "ecr_repository_policy" {
  description = "The repository-level policy attached to the ECR repository (if defined)."
  value       = (var.repository_name != "" && var.repository_policy != null) ? aws_ecr_repository_policy.this[0].policy : null
}
