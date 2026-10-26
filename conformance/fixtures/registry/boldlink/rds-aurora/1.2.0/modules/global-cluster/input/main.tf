
resource "aws_rds_global_cluster" "this" {
  count                        = var.create_global_cluster ? 1 : 0
  global_cluster_identifier    = var.global_cluster_identifier
  database_name                = var.database_name
  deletion_protection          = var.deletion_protection
  engine                       = var.engine
  engine_version               = var.engine_version
  force_destroy                = var.force_destroy
  source_db_cluster_identifier = var.source_db_cluster_identifier
  storage_encrypted            = var.storage_encrypted
}
