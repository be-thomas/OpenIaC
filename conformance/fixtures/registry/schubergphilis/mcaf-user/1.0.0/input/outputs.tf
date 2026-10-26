output "access_key_id" {
  value       = try(aws_iam_access_key.default[0].id, "")
  description = "The access key ID"
}

output "arn" {
  value       = aws_iam_user.default.arn
  description = "The user ARN"
}

output "name" {
  value       = aws_iam_user.default.name
  description = "The user name"
}

variable "region" {
  type        = string
  default     = null
  description = "The AWS region where resources will be created; if omitted the default provider region is used"
}

output "secret_access_key" {
  value       = try(aws_iam_access_key.default[0].secret, "")
  description = "The secret access key"
  sensitive   = true
}

output "ses_smtp_password_v4" {
  value       = try(aws_iam_access_key.default[0].ses_smtp_password_v4, "")
  description = "The SES SMTP password"
  sensitive   = true
}

output "ssm_access_key_id" {
  value       = try(aws_ssm_parameter.access_key_id[0].name, "")
  description = "The SSM access key ID parameter name"
}

output "ssm_secret_access_key" {
  value       = try(aws_ssm_parameter.secret_access_key[0].name, "")
  description = "The SSM secret access key parameter name"
}

output "ssm_ses_smtp_password_v4" {
  value       = try(aws_ssm_parameter.ses_smtp_password_v4[0].name, "")
  description = "The SSM SES SMTP password parameter name"
}
