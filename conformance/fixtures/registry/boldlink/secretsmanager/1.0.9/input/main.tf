
resource "aws_secretsmanager_secret" "this" {
  description                    = var.description
  kms_key_id                     = var.kms_key_id
  name_prefix                    = var.name_prefix
  name                           = var.name
  recovery_window_in_days        = var.recovery_window_in_days
  force_overwrite_replica_secret = var.force_overwrite_replica_secret
  tags = merge({
    "Name" = var.name },
  var.tags)
}

resource "aws_secretsmanager_secret_rotation" "this" {
  count               = var.enable_secretsmanager_secret_rotation ? 1 : 0
  secret_id           = aws_secretsmanager_secret.this.id
  rotation_lambda_arn = var.rotation_lambda_arn
  rotation_rules {
    automatically_after_days = var.automatically_after_days
  }
}

resource "aws_secretsmanager_secret_version" "this" {
  for_each       = var.enable_secretsmanager_secret_version ? var.secrets : null
  secret_id      = aws_secretsmanager_secret.this.id
  secret_string  = lookup(each.value, "secret_binary", null) == null ? lookup(each.value, "secret_string", null) : null
  secret_binary  = lookup(each.value, "secret_string", null) == null ? lookup(each.value, "secret_binary", null) : null
  version_stages = lookup(each.value, "version_stages", null)
}

resource "aws_secretsmanager_secret_policy" "name" {
  count               = var.secret_policy != null ? 1 : 0
  policy              = var.secret_policy
  secret_arn          = aws_secretsmanager_secret.this.arn
  block_public_policy = var.block_public_policy
}
