data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
    ]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "asg" {
  statement {
    sid    = "ECR"
    effect = "Allow"

    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
    ]

    resources = ["*"]
  }

  statement {
    sid    = "Logs"
    effect = "Allow"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:Describe*",
    ]

    resources = ["arn:aws:logs:*:*:*"]
  }
  statement {
    sid    = "ssmAgent"
    effect = "Allow"

    actions = [
      "ssmmessages:CreateControlChannel",
      "ssmmessages:CreateDataChannel",
      "ssmmessages:OpenControlChannel",
      "ssmmessages:OpenDataChannel",
    ]

    resources = ["arn:aws:logs:*:*:*"]
  }
  statement {
    sid    = "CWAgent"
    effect = "Allow"

    actions = [
      "cloudwatch:PutMetricData",
      "ec2:DescribeTags",
    ]

    resources = ["*"]
  }
}

data "template_cloudinit_config" "config" {
  gzip          = false
  base64_encode = true

  # Main cloud-config configuration file.
  part {
    filename     = "init.cfg"
    content_type = "text/cloud-config"
    content      = templatefile("${path.module}/scripts/init.cfg", {})
  }

  # Base Userdata
  part {
    content_type = "text/x-shellscript"
    content      = templatefile("${path.module}/scripts/userdata.sh", {})
  }

  # Cloudwatch config
  part {
    content_type = "text/x-shellscript"
    content = templatefile("${path.module}/scripts/cwldata.sh", {
      log_group = join("", aws_cloudwatch_log_group.main.*.name)
    })
  }

  # Additional script
  part {
    content_type = "text/x-shellscript"
    content      = var.extra_script
  }
}
