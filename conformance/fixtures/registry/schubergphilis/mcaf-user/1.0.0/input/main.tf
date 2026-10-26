locals {
  create_policy = var.create_policy != null ? var.create_policy : var.policy != null
  groups        = local.create_policy ? setunion(var.groups, [aws_iam_group.default[0].name]) : [aws_iam_group.default[0].name]
  ssm_name      = replace(var.name, "@", "_")
}

# Our IAM users are not real users so not going to have MFA configured. Real users
# should instead use AWS SSO and assume a role.
#
# tfsec:ignore:aws-iam-enforce-group-mfa
resource "aws_iam_user" "default" {
  # checkov:skip=CKV_AWS_273
  name                 = "${var.name}${var.postfix ? "Account" : ""}"
  force_destroy        = var.force_destroy
  path                 = var.path
  permissions_boundary = var.permissions_boundary
  tags                 = var.tags
}

resource "aws_iam_access_key" "default" {
  count = var.create_iam_access_key ? 1 : 0
  user  = aws_iam_user.default.name
}

resource "aws_iam_group" "default" {
  count = local.create_policy || length(var.policy_arns) > 0 ? 1 : 0

  name = "${var.name}${var.postfix ? "Group" : ""}"
}

resource "aws_iam_group_policy" "default" {
  count = local.create_policy ? 1 : 0

  name   = "${var.name}${var.postfix ? "Policy" : ""}"
  group  = aws_iam_group.default[0].name
  policy = var.policy
}

resource "aws_iam_group_policy_attachment" "default" {
  for_each = var.policy_arns

  group      = aws_iam_group.default[0].name
  policy_arn = each.value
}

resource "aws_iam_user_group_membership" "default" {
  user   = aws_iam_user.default.name
  groups = local.groups
}

resource "aws_ssm_parameter" "access_key_id" {
  count = var.create_iam_access_key ? 1 : 0

  region = var.region
  name   = "/${lower(local.ssm_name)}${var.postfix ? "account" : ""}/credentials/access_key_id"
  key_id = var.kms_key_id
  type   = "SecureString"
  value  = aws_iam_access_key.default[0].id
  tags   = var.tags
}

resource "aws_ssm_parameter" "secret_access_key" {
  count = var.create_iam_access_key ? 1 : 0

  region = var.region
  name   = "/${lower(local.ssm_name)}${var.postfix ? "account" : ""}/credentials/secret_access_key"
  key_id = var.kms_key_id
  type   = "SecureString"
  value  = aws_iam_access_key.default[0].secret
  tags   = var.tags
}

resource "aws_ssm_parameter" "ses_smtp_password_v4" {
  count = var.create_iam_access_key && var.ssm_ses_smtp_password_v4 ? 1 : 0

  region = var.region
  name   = "/${lower(local.ssm_name)}${var.postfix ? "account" : ""}/credentials/ses_smtp_password_v4"
  key_id = var.kms_key_id
  type   = "SecureString"
  value  = aws_iam_access_key.default[0].ses_smtp_password_v4
  tags   = var.tags
}
