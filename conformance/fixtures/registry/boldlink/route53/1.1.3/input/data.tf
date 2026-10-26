data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "route53_query_logging_policy" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:${local.partition}:logs:us-east-1:${local.account_id}:log-group:/aws/route53/${var.name}:*"]

    principals {
      identifiers = ["route53.${local.dns_suffix}"]
      type        = "Service"
    }
  }
}
