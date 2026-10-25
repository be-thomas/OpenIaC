locals {
  public_subnets  = [cidrsubnet(var.cidr_block, 8, 1), cidrsubnet(var.cidr_block, 8, 2), cidrsubnet(var.cidr_block, 8, 3)]
  private_subnets = [cidrsubnet(var.cidr_block, 8, 4), cidrsubnet(var.cidr_block, 8, 5), cidrsubnet(var.cidr_block, 8, 6)]
  region          = data.aws_region.current.id
  dns_suffix      = data.aws_partition.current.dns_suffix
  account_id      = data.aws_caller_identity.current.id
  partition       = data.aws_partition.current.partition
  tags            = merge({ "Name" = var.name }, var.tags)

  kms_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "Allow KMS Permissions to cluster",
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
