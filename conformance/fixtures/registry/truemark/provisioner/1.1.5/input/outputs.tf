output "provisioner_arn" {
  value = join("", aws_iam_user.provisioner.*.arn)
}

output "provisioner_id" {
  value = join("", aws_iam_user.provisioner.*.id)
}

output "s3_provisioner_policy_arn" {
  value = join("", aws_iam_policy.s3_provisioner.*.arn)
}

output "provisioner_policy_arn" {
  value = join("", aws_iam_policy.provisioner.*.arn)
}