terraform {
  required_providers {
    rad-security = {
      source  = "rad-security/rad-security"
      version = "1.0.3"
    }
  }
}

provider "aws" {
  alias  = "us-west-2"
  region = "us-west-2"
}

provider "aws" {
  alias  = "us-east-1"
  region = "us-east-1"
}

provider "rad-security" {
  access_key_id = var.ksoc_access_key_id
  secret_key    = var.ksoc_secret_key
}

module "rad-security-connect" {
  providers = {
    aws = aws.us-west-2
  }

  # https://registry.terraform.io/modules/rad-security/rad-security-connect/aws/latest
  source  = "rad-security/rad-security-connect/aws"
  version = "<version>"

  enable_eks_audit_logs_pipeline = true
  eks_audit_logs_multi_region    = true
}

data "aws_cloudwatch_log_groups" "ksoc-cw-log-groups-us-west-2" {
  provider              = aws.us-west-2
  log_group_name_prefix = "/aws/eks/"
}

data "aws_cloudwatch_log_groups" "ksoc-cw-log-groups-us-east-1" {
  provider              = aws.us-east-1
  log_group_name_prefix = "/aws/eks/"
}

resource "aws_cloudwatch_log_subscription_filter" "ksoc-subscription-filter-us-west-2" {
  for_each = { for name in data.aws_cloudwatch_log_groups.ksoc-cw-log-groups-us-west-2.log_group_names : name => name }
  provider = aws.us-west-2

  name            = "ksoc-audit-logs"
  role_arn        = module.rad-security-connect.eks_audit_logs_cloudwatch_iam_role_arn
  log_group_name  = each.value
  filter_pattern  = module.rad-security-connect.eks_audit_logs_filter_pattern
  destination_arn = module.rad-security-connect.eks_audit_logs_firehose_arn
  depends_on      = [module.rad-security-connect]
}

resource "aws_cloudwatch_log_subscription_filter" "ksoc-subscription-filter-us-east-1" {
  for_each = { for name in data.aws_cloudwatch_log_groups.ksoc-cw-log-groups-us-east-1.log_group_names : name => name }
  provider = aws.us-east-1

  name            = "ksoc-audit-logs"
  role_arn        = module.rad-security-connect.eks_audit_logs_cloudwatch_iam_role_arn
  log_group_name  = each.value
  filter_pattern  = module.rad-security-connect.eks_audit_logs_filter_pattern
  destination_arn = module.rad-security-connect.eks_audit_logs_firehose_arn
  depends_on      = [module.rad-security-connect]
}
