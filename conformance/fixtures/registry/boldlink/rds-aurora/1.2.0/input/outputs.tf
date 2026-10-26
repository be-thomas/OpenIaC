output "arn" {
  description = "Amazon Resource Name (ARN) of cluster"
  value       = aws_rds_cluster.this.arn
}

output "id" {
  description = "The RDS Cluster Identifier"
  value       = aws_rds_cluster.this.id
}

output "cluster_identifier" {
  description = "The RDS Cluster Identifier"
  value       = aws_rds_cluster.this.cluster_identifier
}

output "cluster_resource_id" {
  description = "The RDS Cluster Resource ID"
  value       = aws_rds_cluster.this.cluster_resource_id
}

output "availability_zones" {
  description = "The availability zone of the instance"
  value       = aws_rds_cluster.this.availability_zones
}

output "backup_retention_period" {
  description = "The backup retention period"
  value       = aws_rds_cluster.this.backup_retention_period
}

output "preferred_backup_window" {
  description = "The daily time range during which the backups happen"
  value       = aws_rds_cluster.this.preferred_backup_window
}

output "preferred_maintenance_window" {
  description = "The maintenance window"
  value       = aws_rds_cluster.this.preferred_maintenance_window
}

output "endpoint" {
  description = "The DNS address of the RDS instance"
  value       = aws_rds_cluster.this.endpoint
}

output "reader_endpoint" {
  description = " A read-only endpoint for the Aurora cluster, automatically load-balanced across replicas"
  value       = aws_rds_cluster.this.reader_endpoint
}

output "engine" {
  description = "The database engine"
  value       = aws_rds_cluster.this.engine
}

output "engine_version_actual" {
  description = " The running version of the database."
  value       = aws_rds_cluster.this.engine_version_actual
}

output "database_name" {
  description = "The database name"
  value       = aws_rds_cluster.this.database_name
}

output "port" {
  description = "The database port"
  value       = aws_rds_cluster.this.port
}

output "master_username" {
  description = " The master username for the database"
  value       = aws_rds_cluster.this.master_username
}

output "storage_encrypted" {
  description = "Specifies whether the DB cluster is encrypted"
  value       = aws_rds_cluster.this.storage_encrypted
}

output "replication_source_identifier" {
  description = "ARN of the source DB cluster or DB instance if this DB cluster is created as a Read Replica."
  value       = aws_rds_cluster.this.replication_source_identifier
}

output "hosted_zone_id" {
  description = "The Route53 Hosted Zone ID of the endpoint"
  value       = aws_rds_cluster.this.hosted_zone_id
}

output "tags_all" {
  description = "A map of tags assigned to the resource, including those inherited from the provider default_tags"
  value       = aws_rds_cluster.this.tags_all
}
