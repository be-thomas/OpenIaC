output "iam_user_arn" {
  description = "ARN of the IAM user (if auth_method is iam_user)"
  value       = var.auth_method == "iam_user" ? module.workspace_iam_user[0].arn : null
}

output "iam_role_arn" {
  description = "ARN of the IAM role (if auth_method is iam_role)"
  value       = var.auth_method == "iam_role" ? module.workspace_iam_role[0].arn : null
}

output "iam_role_oidc_arn" {
  description = "ARN of the IAM role for OIDC (if auth_method is iam_role_oidc)"
  value       = local.enable_oidc ? module.workspace_iam_role_oidc[0].arn : null
}
