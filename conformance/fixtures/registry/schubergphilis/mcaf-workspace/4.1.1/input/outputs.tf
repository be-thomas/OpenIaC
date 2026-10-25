output "arn" {
  value       = try(module.auth[0].iam_user_arn, "")
  description = "The workspace IAM user ARN"
}

output "iam_user_arn" {
  description = "ARN of the IAM user (if auth_method is iam_user)"
  value       = try(module.auth[0].iam_user_arn, null)
}

output "iam_role_arn" {
  description = "ARN of the IAM role (if auth_method is iam_role)"
  value       = try(module.auth[0].iam_role_arn, null)
}

output "iam_role_oidc_arn" {
  description = "ARN of the IAM role for OIDC (if auth_method is iam_role_oidc)"
  value       = try(module.auth[0].iam_role_oidc_arn, null)
}

output "workspace_name" {
  value       = module.tfe-workspace.name
  description = "The Terraform Cloud workspace name"
}

output "workspace_id" {
  value       = module.tfe-workspace.id
  description = "The Terraform Cloud workspace ID"
}
