data "aws_caller_identity" "current" {}

locals {
  s3_bucket       = var.s3_bucket == null ? "${data.aws_caller_identity.current.account_id}-terraform" : var.s3_bucket
  trusted_account = var.trusted_account == null ? data.aws_caller_identity.current.account_id : var.trusted_account
}

resource "aws_iam_user" "provisioner" {
  count = var.create_user ? 1 : 0
  name  = var.name
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["arn:aws:iam::${local.trusted_account}:root"]
      type        = "AWS"
    }
  }
}

resource "aws_iam_role" "provisioner" {
  count                = var.create_role ? 1 : 0
  name                 = var.name
  description          = var.description
  assume_role_policy   = data.aws_iam_policy_document.assume_role_policy.json
  max_session_duration = var.max_session_duration
}

# This policy grants the provisioner user access to specific paths in the S3 bucket holding terraform state.
# This is needed to prevent different provisioner users from stepping on one another's changes. Additionally,
# there is sensitive information stored in the state files in these S3 buckets which should be restricted.

data "aws_iam_policy_document" "s3_provisioner" {
  statement {
    sid       = "AllowBucketList"
    effect    = "Allow"
    actions   = ["s3:ListAllMyBuckets", "s3:GetBucketLocation"]
    resources = ["arn:aws:s3:::*"]
  }
  statement {
    sid       = "AllowListBucket"
    effect    = "Allow"
    actions   = ["s3:ListBucket"]
    resources = ["arn:aws:s3:::${local.s3_bucket}"]
  }
  statement {
    sid       = "AllowPath"
    effect    = "Allow"
    actions   = ["s3:GetObject", "s3:PutObject", "s3:ListObjects"]
    resources = ["arn:aws:s3:::${local.s3_bucket}/${var.s3_prefix}/*"]
  }
  statement {
    sid       = "AllowWorkspacePath"
    effect    = "Allow"
    actions   = ["s3:GetObject", "s3:PutObject", "s3:ListObjects"]
    resources = ["arn:aws:s3:::${local.s3_bucket}/env:/*/${var.s3_prefix}/*"]
  }
  statement {
    sid       = "AllowDynamo"
    effect    = "Allow"
    actions   = ["dynamodb:PutItem", "dynamodb:GetItem", "dynamodb:DeleteItem"]
    resources = ["arn:aws:dynamodb:*:${data.aws_caller_identity.current.account_id}:table/${local.s3_bucket}"]
  }
}

resource "aws_iam_policy" "s3_provisioner" {
  count       = var.s3_prefix != null && (var.create_user || var.create_role) ? 1 : 0
  name        = "${var.name}-s3-terraform"
  path        = var.path
  description = "Allows access to ${local.s3_bucket}/${var.s3_prefix}"
  policy      = data.aws_iam_policy_document.s3_provisioner.json
}

resource "aws_iam_user_policy_attachment" "s3_provisioner" {
  count      = var.s3_prefix != null && var.create_user ? 1 : 0
  policy_arn = aws_iam_policy.s3_provisioner[count.index].arn
  user       = aws_iam_user.provisioner[count.index].name
}

resource "aws_iam_role_policy_attachment" "s3_provisioner" {
  count      = var.s3_prefix != null && var.create_role ? 1 : 0
  policy_arn = aws_iam_policy.s3_provisioner[count.index].arn
  role       = aws_iam_role.provisioner[count.index].name
}

resource "aws_iam_policy" "provisioner" {
  count       = var.policy != null ? 1 : 0
  name        = var.name
  path        = var.path
  description = "Access policy for IAM user ${var.name}. Created and attached by TrueMark terraform module terraform-aws-provisioner."
  policy      = var.policy
}

resource "aws_iam_user_policy_attachment" "provisioner" {
  count      = var.policy != null && var.create_user ? 1 : 0
  policy_arn = aws_iam_policy.provisioner[count.index].arn
  user       = aws_iam_user.provisioner[count.index].name
}

resource "aws_iam_role_policy_attachment" "provisioner" {
  count      = var.policy != null && var.create_role ? 1 : 0
  policy_arn = aws_iam_policy.provisioner[count.index].arn
  role       = aws_iam_role.provisioner[count.index].name
}

resource "aws_iam_policy" "provisioner_n" {
  count       = length(var.policies)
  name        = "${var.name}-${count.index}"
  path        = var.path
  description = "Access policy for IAM user ${var.name}"
  policy      = var.policies[count.index]
}

resource "aws_iam_user_policy_attachment" "provisioner_n" {
  count      = var.create_user ? length(var.policies) : 0
  policy_arn = aws_iam_policy.provisioner_n[count.index].arn
  user       = join("", aws_iam_user.provisioner.*.name)
}

resource "aws_iam_role_policy_attachment" "provisioner_n" {
  count      = var.create_role ? length(var.policies) : 0
  policy_arn = aws_iam_policy.provisioner_n[count.index].arn
  role       = join("", aws_iam_role.provisioner.*.name)
}

resource "aws_iam_user_policy_attachment" "attachments" {
  count      = var.create_user ? length(var.policy_arns) : 0
  policy_arn = var.policy_arns[count.index]
  user       = join("", aws_iam_user.provisioner.*.name)
}

resource "aws_iam_role_policy_attachment" "attachments" {
  count      = var.create_role ? length(var.policy_arns) : 0
  policy_arn = var.policy_arns[count.index]
  role       = join("", aws_iam_role.provisioner.*.name)
}
