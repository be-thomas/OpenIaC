output "role_arn" {
  value       = var.secondary_region ? null : aws_iam_role.this[0].arn
  description = "AWS IAM Role ARN which Rad Security uses to connect"
}

output "eks_audit_logs_cloudwatch_iam_role_arn" {
  value       = var.enable_eks_audit_logs_pipeline && !var.secondary_region ? aws_iam_role.cloudwatch[0].arn : null
  description = "AWS IAM Role ARN for Cloudwatch to Firehose"
}

output "eks_audit_logs_firehose_iam_role_arn" {
  value       = var.enable_eks_audit_logs_pipeline && !var.secondary_region ? aws_iam_role.firehose[0].arn : null
  description = "AWS IAM Role ARN for Firehose to S3"
}

output "eks_audit_logs_filter_pattern" {
  value       = var.eks_audit_logs_filter_pattern
  description = "The Cloudwatch Log Subscription Filter pattern"
}

output "eks_audit_logs_firehose_arn" {
  value       = var.enable_eks_audit_logs_pipeline ? aws_kinesis_firehose_delivery_stream.firehose[0].arn : null
  description = "The Firehose delivery stream ARN"
}
