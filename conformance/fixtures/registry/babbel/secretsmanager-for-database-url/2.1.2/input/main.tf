locals {
  username        = var.db_instance != null ? var.db_instance.username : var.rds_cluster.master_username
  password        = var.db_instance != null ? var.db_instance.password : var.rds_cluster.master_password
  endpoint        = var.db_instance != null ? var.db_instance.address : var.rds_cluster.endpoint
  reader_endpoint = var.db_instance != null ? var.db_instance.address : var.rds_cluster.reader_endpoint
}

resource "aws_secretsmanager_secret" "endpoint" {
  name        = "${var.name_prefix}.database_url"
  description = "Secret value is managed via Terraform"

  tags = merge(var.default_tags, var.secretsmanager_secret_tags)
}

resource "aws_secretsmanager_secret_version" "endpoint" {
  secret_id = aws_secretsmanager_secret.endpoint.arn

  secret_string = "${var.protocol}://${urlencode(local.username)}:${urlencode(local.password)}@${local.endpoint}/${urlencode(var.database_name)}"
}

resource "aws_secretsmanager_secret" "reader_endpoint" {
  name        = "${var.name_prefix}.database_replica_url"
  description = "Secret value is managed via Terraform"

  tags = merge(var.default_tags, var.secretsmanager_secret_tags)
}

resource "aws_secretsmanager_secret_version" "reader_endpoint" {
  secret_id = aws_secretsmanager_secret.reader_endpoint.arn

  secret_string = "${var.protocol}://${urlencode(local.username)}:${urlencode(local.password)}@${local.reader_endpoint}/${urlencode(var.database_name)}"
}

moved {
  from = aws_secretsmanager_secret.this
  to   = aws_secretsmanager_secret.endpoint
}

moved {
  from = aws_secretsmanager_secret_version.this
  to   = aws_secretsmanager_secret_version.endpoint
}
