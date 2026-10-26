# Scanning configuration
resource "aws_ecr_registry_scanning_configuration" "this" {
  count     = length(var.scanning_configuration) > 0 ? 1 : 0
  scan_type = try(var.scanning_configuration["scan_type"], "ENHANCED")

  dynamic "rule" {
    for_each = lookup(var.scanning_configuration, "rules", null)
    content {
      scan_frequency = rule.value.scan_frequency

      repository_filter {
        filter      = rule.value.repository_filter.filter
        filter_type = rule.value.repository_filter.filter_type
      }
    }
  }
}

# Pull through cache rule
resource "aws_ecr_pull_through_cache_rule" "this" {
  count                 = var.enable_pull_through_cache_rule ? 1 : 0
  ecr_repository_prefix = var.ecr_repository_prefix
  upstream_registry_url = var.upstream_registry_url
}

# Replication configuration
resource "aws_ecr_replication_configuration" "this" {
  count = length(var.replication_configuration) > 0 ? 1 : 0
  replication_configuration {
    dynamic "rule" {
      for_each = try([var.replication_configuration["rule"]], [])
      content {
        dynamic "destination" {
          for_each = lookup(rule.value, "destinations", [])
          content {
            region      = destination.value.region
            registry_id = destination.value.registry_id
          }
        }
        dynamic "repository_filter" {
          for_each = lookup(rule.value, "repository_filters", [])
          content {
            filter      = try(repository_filter.value.filter, null)
            filter_type = try(repository_filter.value.filter_type, "PREFIX_MATCH")
          }
        }
      }
    }
  }
}

# ECR Repository with encryption always enabled (optional)
resource "aws_ecr_repository" "this" {
  count                = var.repository_name != "" ? 1 : 0
  name                 = var.repository_name
  image_tag_mutability = var.image_tag_mutability

  encryption_configuration {
    encryption_type = var.encryption_type
    # When using KMS, specify the kms_key; otherwise, use null.
    kms_key = var.encryption_type == "KMS" ? var.kms_key_id : null
  }

  tags = var.tags
}


# ECR Repository Policy (optional)
resource "aws_ecr_repository_policy" "this" {
  count      = (var.repository_name != "" && var.repository_policy != null) ? 1 : 0
  repository = aws_ecr_repository.this[0].name
  policy     = var.repository_policy
}
