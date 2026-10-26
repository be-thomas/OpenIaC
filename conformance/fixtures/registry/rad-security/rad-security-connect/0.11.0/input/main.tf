# Policy
data "aws_iam_policy_document" "assume_role" {
  count = var.secondary_region ? 0 : 1
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
    ]

    principals {
      type        = "AWS"
      identifiers = [var.rad-security_assumed_role_arn]
    }

    dynamic "condition" {
      for_each = var.aws_external_id != "" ? [var.aws_external_id] : []
      content {
        test     = "StringEquals"
        variable = "sts:ExternalId"
        values   = [condition.value]
      }
    }
  }

  statement {
    effect = "Allow"
    actions = [
      "sts:TagSession"
    ]

    principals {
      type        = "AWS"
      identifiers = [var.rad-security_assumed_role_arn]
    }
  }
}

# Role
resource "aws_iam_role" "this" {
  count                = var.secondary_region ? 0 : 1
  name                 = "rad-security-connect"
  path                 = "/"
  max_session_duration = 3600
  description          = "IAM role for rad-security-connect"
  tags                 = var.tags

  assume_role_policy = data.aws_iam_policy_document.assume_role[0].json
}

# Instance profile
resource "aws_iam_instance_profile" "this" {
  count = var.secondary_region ? 0 : 1

  name = aws_iam_role.this[0].name
  role = aws_iam_role.this[0].name
  tags = var.tags
}

resource "rad-security_aws_register" "this" {
  count                         = var.secondary_region ? 0 : 1
  rad_security_assumed_role_arn = var.rad-security_assumed_role_arn
  aws_account_id                = data.aws_caller_identity.current.account_id
}

resource "aws_s3_bucket_lifecycle_configuration" "audit_logs" {
  count  = var.enable_eks_audit_logs_pipeline ? 1 : 0
  bucket = aws_s3_bucket.audit_logs[0].bucket

  rule {
    id     = "delete-old-objects"
    status = "Enabled"

    filter {}

    expiration {
      days = var.eks_audit_logs_bucket_object_age
    }

    noncurrent_version_expiration {
      noncurrent_days = var.eks_audit_logs_bucket_object_age
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = var.eks_audit_logs_bucket_object_age
    }
  }
}

locals {
  readonly_s3_buckets_enabled = !var.secondary_region && length(var.readonly_s3_buckets) > 0
}

data "aws_iam_policy_document" "readonly_s3_buckets_access" {
  count = local.readonly_s3_buckets_enabled ? 1 : 0

  statement {
    effect = "Allow"
    actions = [
      "s3:ListBucket",
      "s3:GetBucketLocation",
    ]
    resources = [for bucket in var.readonly_s3_buckets : "arn:aws:s3:::${bucket}"]
  }

  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject",
    ]
    resources = [for bucket in var.readonly_s3_buckets : "arn:aws:s3:::${bucket}/*"]
  }
}

resource "aws_iam_policy" "readonly_s3_buckets_access" {
  count = local.readonly_s3_buckets_enabled ? 1 : 0

  name        = "rad-security-readonly-s3-buckets-access"
  path        = "/"
  description = "Policy to allow Rad Security to read logs from specified S3 buckets"
  policy      = data.aws_iam_policy_document.readonly_s3_buckets_access[0].json
  tags        = var.tags
}

resource "aws_iam_role_policy_attachment" "readonly_s3_buckets_access" {
  count = local.readonly_s3_buckets_enabled ? 1 : 0

  role       = aws_iam_role.this[0].name
  policy_arn = aws_iam_policy.readonly_s3_buckets_access[0].arn
}

locals {
  ecr_image_pull_enabled = !var.secondary_region && var.enable_ecr_image_pull
}

data "aws_iam_policy_document" "ecr_image_pull" {
  count = local.ecr_image_pull_enabled ? 1 : 0

  statement {
    effect = "Allow"
    actions = [
      "ecr:GetAuthorizationToken",
    ]
    resources = ["*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "ecr:BatchGetImage",
      "ecr:GetDownloadUrlForLayer",
    ]
    resources = var.ecr_repository_arns
  }
}

resource "aws_iam_policy" "ecr_image_pull" {
  count = local.ecr_image_pull_enabled ? 1 : 0

  name        = "rad-security-ecr-image-pull"
  path        = "/"
  description = "Policy to allow Rad Security to pull container images from ECR for SBOM generation and vulnerability scanning"
  policy      = data.aws_iam_policy_document.ecr_image_pull[0].json
  tags        = var.tags
}

resource "aws_iam_role_policy_attachment" "ecr_image_pull" {
  count = local.ecr_image_pull_enabled ? 1 : 0

  role       = aws_iam_role.this[0].name
  policy_arn = aws_iam_policy.ecr_image_pull[0].arn
}
