output "service_account_namespace" {
  value       = local.enabled ? var.service_account_namespace : null
  description = "Kubernetes Service Account namespace"
}

output "service_account_name" {
  value       = local.enabled ? var.service_account_name : null
  description = "Kubernetes Service Account name"
}

output "service_account_role_name" {
  value       = one(aws_iam_role.service_account[*].name)
  description = "IAM role name"
}

output "service_account_role_unique_id" {
  value       = one(aws_iam_role.service_account[*].unique_id)
  description = "IAM role unique ID"
}

output "service_account_role_arn" {
  value       = one(aws_iam_role.service_account[*].arn)
  description = "IAM role ARN"
}

output "service_account_policy_name" {
  value       = one(aws_iam_policy.service_account[*].name)
  description = "IAM policy name"
}

output "service_account_policy_id" {
  value       = one(aws_iam_policy.service_account[*].id)
  description = "IAM policy ID"
}

output "service_account_policy_arn" {
  value       = one(aws_iam_policy.service_account[*].arn)
  description = "IAM policy ARN"
}
