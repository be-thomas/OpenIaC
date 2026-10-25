locals {
  assume_policy = var.assume_policy != null ? var.assume_policy : data.aws_iam_policy_document.default.json
  create_policy = var.create_policy != null ? var.create_policy : var.role_policy != null
}

data "aws_iam_policy_document" "default" {
  statement {
    actions = [
      "sts:AssumeRole"
    ]

    principals {
      type        = var.principal_type
      identifiers = var.principal_identifiers
    }
  }
}

resource "aws_iam_role" "default" {
  assume_role_policy    = local.assume_policy
  description           = var.description
  force_detach_policies = var.force_detach_policies
  max_session_duration  = var.max_session_duration
  name                  = var.name != null ? "${var.name}${var.postfix ? "Role" : ""}" : null
  name_prefix           = var.name_prefix != null ? "${var.name_prefix}${var.postfix ? "Role" : ""}" : null
  path                  = var.path
  permissions_boundary  = var.permissions_boundary
  tags                  = var.tags
}

resource "aws_iam_role_policy" "default" {
  count = local.create_policy ? 1 : 0

  name        = var.name != null ? "${var.name}${var.postfix ? "Policy" : ""}" : null
  name_prefix = var.name_prefix != null ? "${var.name_prefix}${var.postfix ? "Policy" : ""}" : null
  policy      = var.role_policy
  role        = aws_iam_role.default.id
}

resource "aws_iam_role_policy_attachment" "default" {
  for_each = var.policy_arns

  role       = aws_iam_role.default.name
  policy_arn = each.value
}
