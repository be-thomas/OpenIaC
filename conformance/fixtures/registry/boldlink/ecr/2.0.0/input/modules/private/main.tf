# ECR Repository
resource "aws_ecr_repository" "this" {
  count                = var.create_private_repository ? 1 : 0
  name                 = var.name
  image_tag_mutability = var.image_tag_mutability
  image_scanning_configuration {
    scan_on_push = var.scan_on_push
  }

  dynamic "encryption_configuration" {
    for_each = [var.encryption_configuration]
    content {
      encryption_type = "KMS"
      kms_key         = encryption_configuration.value.kms_key
    }
  }

  tags = merge(
    {
      "Name" = "${var.name}-tag"
    },
    var.other_tags,
  )
}

resource "aws_ecr_repository_policy" "this" {
  count      = var.aws_ecr_repository_policy != "" && var.create_private_repository ? 1 : 0
  repository = aws_ecr_repository.this[count.index].name
  policy     = var.aws_ecr_repository_policy
}

resource "aws_ecr_lifecycle_policy" "this" {
  count      = var.aws_ecr_lifecycle_policy != "" && var.create_private_repository ? 1 : 0
  repository = aws_ecr_repository.this[count.index].name
  policy     = var.aws_ecr_lifecycle_policy
}
