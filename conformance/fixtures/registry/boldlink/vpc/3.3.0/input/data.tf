data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_nat_gateways" "all" {
  count  = var.enable_private_subnets == true ? 1 : 0
  vpc_id = aws_vpc.main.id
  filter {
    name   = "state"
    values = ["available"]
  }
  filter {
    name   = "tag:Name"
    values = ["${var.name}.*.nat-gw.*"]
  }

  depends_on = [module.public_subnets]
}

data "aws_iam_policy_document" "assume_policy" {
  statement {
    sid     = "TrustVPCFlowLogs"
    actions = ["sts:AssumeRole"]
    effect  = "Allow"
    principals {
      type        = "Service"
      identifiers = ["vpc-flow-logs.${local.dns_suffix}"]
    }
  }
}

data "aws_iam_policy_document" "role_policy" {
  version = "2012-10-17"
  statement {
    sid    = "LogsGroupDelivery"
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeLogGroups",
      "logs:DescribeLogStreams"
    ]
    resources = [
      "arn:${local.partition}:logs:${local.region}:${local.account_id}:log-group:/aws/vpc/${var.name}:*"
    ]
  }
}

# data "aws_iam_policy_document" "s3_role_policy" {
#   version = "2012-10-17"
#   statement {
#     sid    = "LogsS3Delivery"
#     effect = "Allow"
#     actions = [
#       "logs:CreateLogDelivery",
#       "logs:DeleteLogDelivery"
#     ]
#     resources = ["*"]
#   }
# }

# data "aws_iam_policy_document" "s3_bucket_policy" {
#   version = "2012-10-17"
#   statement {
#     sid    = "AWSLogDeliveryWrite"
#     effect = "Allow"
#     principals {
#       type        = "Service"
#       identifiers = ["delivery.logs.${local.dns_suffix}"]
#     }
#     actions = ["s3:PutObject"]
#     resources = [
#       "arn:${local.partition}:s3:::${local.logs_bucket_name}/*",
#       "arn:${local.partition}:s3:::${local.logs_bucket_name}/AWSLogs/${local.account_id}/*"
#     ]
#     condition {
#       test     = "StringEquals"
#       variable = "s3:x-amz-acl"
#       values   = ["bucket-owner-full-control"]
#     }
#     condition {
#       test     = "StringEquals"
#       variable = "aws:SourceAccount"
#       values   = [local.account_id]
#     }
#     condition {
#       test     = "ArnLike"
#       variable = "aws:SourceArn"
#       values   = ["arn:${local.partition}:logs:${local.region}:${local.account_id}:*"]
#     }
#   }
#   statement {
#     sid    = "AWSLogDeliveryCheck"
#     effect = "Allow"
#     principals {
#       type        = "Service"
#       identifiers = ["delivery.logs.${local.dns_suffix}"]
#     }
#     actions = [
#       "s3:ListBucket",
#       "s3:GetBucketAcl"
#     ]
#     resources = ["arn:${local.partition}:s3:::${local.logs_bucket_name}"]
#     condition {
#       test     = "StringEquals"
#       variable = "aws:SourceAccount"
#       values   = [local.account_id]
#     }
#     condition {
#       test     = "ArnLike"
#       variable = "aws:SourceArn"
#       values   = ["arn:${local.partition}:logs:${local.region}:${local.account_id}:*"]
#     }
#   }
# }
