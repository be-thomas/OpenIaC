resource "aws_lambda_function_event_invoke_config" "destination" {
  count         = var.destinations_on_failure_arn != "" || var.destinations_on_success_arn != "" ? 1 : 0
  function_name = aws_lambda_function.function.arn

  destination_config {

    dynamic "on_failure" {
      for_each = var.destinations_on_failure_arn != "" ? ["1"] : []
      content {
        destination = var.destinations_on_failure_arn
      }
    }
    dynamic "on_success" {
      for_each = var.destinations_on_success_arn != "" ? ["1"] : []
      content {
        destination = var.destinations_on_success_arn
      }
    }
  }

  depends_on = [
    aws_iam_role_policy.destination_on_success,
    aws_iam_role_policy.destination_on_failure,
  ]
}

data "aws_iam_policy_document" "destination_on_failure" {
  count = var.destinations_on_failure_arn == "" ? 0 : 1
  statement {
    actions = [
      "sns:Publish",
      "sqs:SendMessage",
      "lambda:InvokeFunction",
      "events:PutEvents"
    ]
    resources = [
      var.destinations_on_failure_arn
    ]
    sid = "AsynchronousInvocationOnFailure"
  }
}

data "aws_iam_policy_document" "destination_on_success" {
  count = var.destinations_on_success_arn == "" ? 0 : 1
  statement {
    actions = [
      "sns:Publish",
      "sqs:SendMessage",
      "lambda:InvokeFunction",
      "events:PutEvents"
    ]
    resources = [
      var.destinations_on_success_arn
    ]
    sid = "AsynchronousInvocationOnSuccess"
  }
}

resource "aws_iam_role_policy" "destination_on_failure" {
  count  = var.destinations_on_failure_arn == "" ? 0 : 1
  name   = "destination-failure"
  policy = data.aws_iam_policy_document.destination_on_failure.0.json
  role   = aws_iam_role.function.id
}

resource "aws_iam_role_policy" "destination_on_success" {
  count  = var.destinations_on_success_arn == "" ? 0 : 1
  name   = "destination-success"
  policy = data.aws_iam_policy_document.destination_on_success.0.json
  role   = aws_iam_role.function.id
}
