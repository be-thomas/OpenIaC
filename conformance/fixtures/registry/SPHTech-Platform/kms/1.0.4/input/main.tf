locals {
  account_id = data.aws_caller_identity.current.account_id
  region     = var.region != null ? var.region : data.aws_region.current.region
}

resource "aws_kms_key" "key" {
  region                  = local.region
  description             = var.key_description
  deletion_window_in_days = var.deletion_window_in_days
  enable_key_rotation     = var.enable_key_rotation
  policy                  = data.aws_iam_policy_document.combined_key_policy.json

  tags = var.tags
}

resource "aws_kms_alias" "key" {
  count = var.alias != null && var.alias != "" ? 1 : 0

  region        = local.region
  name          = var.alias
  target_key_id = aws_kms_key.key.key_id
}
