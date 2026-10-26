data "aws_iam_policy_document" "tachyon_bucket" {
  statement {
    sid       = "AllowListBucketContents"
    effect    = "Allow"
    actions   = ["s3:ListBucket"]
    resources = [var.bucket_arn]
  }
  statement {
    sid    = "AllowModificationsToBucket"
    effect = "Allow"
    actions = [
      "s3:AbortMultipartUpload",
      "s3:DeleteObject",
      "s3:GetBucketAcl",
      "s3:GetBucketLocation",
      "s3:GetBucketPolicy",
      "s3:GetObject",
      "s3:GetObjectAcl",
      "s3:ListBucket",
      "s3:ListBucketMultipartUploads",
      "s3:ListMultipartUploadParts",
      "s3:PutObject",
      "s3:PutObjectAcl"
    ]
    resources = ["${var.bucket_arn}/*"]
  }
}

module "tachyon" {
  source        = "terraform-aws-modules/lambda/aws"
  version       = "8.7.0"
  region        = "us-east-1"
  function_name = "tachyon-${var.bucket_name}"
  description   = "Lambda@Edge Tachyon for ${var.bucket_name}"
  handler       = "lambda-handler.handler"
  runtime       = "nodejs22.x"
  # 30s is the hard ceiling AWS allows for an origin-request Lambda@Edge, so this cannot be raised.
  # Resizing must fit inside it - see tachyon_memory_size, which is the lever that actually helps.
  timeout                           = 30
  memory_size                       = var.tachyon_memory_size
  lambda_at_edge                    = true
  publish                           = true
  create_package                    = false
  local_existing_package            = "${path.module}/lambda/tachyon/tachyon-r53.zip"
  cloudwatch_logs_retention_in_days = 30
  attach_policy_jsons               = true
  number_of_policy_jsons            = 1
  policy_jsons                      = [data.aws_iam_policy_document.tachyon_bucket.json]

  # CloudFront keeps replicas of an edge function for hours after its distribution is gone, and
  # the function cannot be deleted until they are, so destroying the stack only removes it from the
  # state. Once the replicas are gone it can be deleted by hand
  # (`aws lambda delete-function --region us-east-1 --function-name tachyon-<bucket_name>`), or with
  # the AWS account. Until then, applying the stack again with the same bucket_name fails with
  # "Function already exist".
  skip_destroy = true
}
