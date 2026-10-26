output "cluster_arn" {
  description = "Amazon Resource Name (ARN) of cluster"
  value       = aws_docdb_cluster.this.arn
}
output "instance_arn" {
  description = "Amazon Resource Name (ARN) of cluster instance"
  value       = aws_docdb_cluster_instance.this.*.arn
}
output "cluster_members" {
  description = "List of DocDB Instances that are a part of this cluster"
  value       = aws_docdb_cluster.this.cluster_members
}

output "cluster_resource_id" {
  description = "The DocDB Cluster Resource ID"
  value       = aws_docdb_cluster.this.cluster_resource_id
}

output "endpoint" {
  description = "The DNS address of the DocDB instance"
  value       = aws_docdb_cluster.this.endpoint
}

output "hosted_zone_id" {
  description = "The Route53 Hosted Zone ID of the endpoint"
  value       = aws_docdb_cluster.this.hosted_zone_id
}

output "id" {
  description = "The DocDB Cluster Identifier"
  value       = aws_docdb_cluster.this.id
}

output "reader_endpoint" {
  description = "A read-only endpoint for the DocDB cluster, automatically load-balanced across replicas"
  value       = aws_docdb_cluster.this.reader_endpoint
}
output "engine_version" {
  description = "The database engine version"
  value       = aws_docdb_cluster_instance.this.*.engine
}

output "kms_key_id" {
  description = "The ARN for the KMS encryption key if one is set to the cluster."
  value       = aws_docdb_cluster_instance.this.*.kms_key_id
}

output "port" {
  description = "The database port"
  value       = aws_docdb_cluster_instance.this.*.port
}

output "preferred_backup_window" {
  description = "The daily time range during which automated backups are created if automated backups are enabled."
  value       = aws_docdb_cluster_instance.this.*.preferred_backup_window
}

output "storage_encrypted" {
  description = "Specifies whether the DB cluster is encrypted."
  value       = aws_docdb_cluster_instance.this.*.storage_encrypted
}

output "writer" {
  description = "Boolean indicating if this instance is writable. False indicates this instance is a read replica."
  value       = aws_docdb_cluster_instance.this.*.writer
}

output "ca_cert_identifier" {
  description = "(Optional) The identifier of the CA certificate for the DB instance."
  value       = aws_docdb_cluster_instance.this.*.ca_cert_identifier
}
output "tags_all" {
  description = "A map of tags assigned to the resource, including those inherited from the provider default_tags"
  value       = aws_docdb_cluster.this.tags_all
}
