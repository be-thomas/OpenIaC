data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
    ]
    principals {
      type        = "Service"
      identifiers = ["lambda.${local.dns_suffix}"]
    }
  }
}

data "aws_iam_policy_document" "lambda" {
  statement {
    sid    = "AllowLambdaLogGroupPermissions"
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:aws:logs:*:*:*"]
  }
  dynamic "statement" {
    for_each = var.additional_lambda_permissions
    content {
      sid = try(statement.value.sid, null)

      actions = try(statement.value.actions, null)

      effect    = try(statement.value.effect, null)
      resources = statement.value.resources

      dynamic "condition" {
        for_each = try([statement.value.condition], [])
        content {
          test     = condition.value.test
          variable = condition.value.variable
          values   = condition.value.values
        }
      }
    }
  }
}
