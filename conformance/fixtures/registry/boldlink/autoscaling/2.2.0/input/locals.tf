locals {
  region                  = data.aws_region.current.name
  partition               = data.aws_partition.current.partition
  account_id              = data.aws_caller_identity.current.account_id
  dns_suffix              = data.aws_partition.current.dns_suffix
  launch_template         = var.external_launch_template_name == null ? var.name : var.external_launch_template_name
  launch_template_id      = var.create_launch_template ? aws_launch_template.main[0].id : var.launch_template_id
  launch_template_version = coalesce(var.launch_template_version, try(aws_launch_template.main[0].latest_version, null), var.external_launch_template_version)

  kms_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "Allow Cloud Watch Logs",
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "logs.${local.region}.${local.dns_suffix}"
        },
        "Action" : [
          "kms:Encrypt*",
          "kms:Decrypt*",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:Describe*"
        ],
        "Resource" : "*"
      },
      {
        "Sid" : "Enable IAM User Permissions",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "arn:${local.partition}:iam::${local.account_id}:root"
        },
        "Action" : [
          "kms:*"
        ],
        "Resource" : "*",
      }
    ]
  })
}
