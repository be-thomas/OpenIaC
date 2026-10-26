output "endpoint" {
  description = "Redis primary or configuration endpoint, whichever is appropriate for the given cluster mode"
  value       = var.use_serverless ? try(aws_elasticache_serverless_cache.this[0].endpoint[0].address, null) : try(aws_elasticache_replication_group.this[0].primary_endpoint_address, null)
}

output "reader_endpoint_address" {
  description = "The address of the endpoint for the reader node in the replication group, if the cluster mode is disabled"
  value       = try(aws_elasticache_replication_group.this[0].reader_endpoint_address, null)
}

output "member_clusters" {
  description = "Redis cluster members"
  value       = try(aws_elasticache_replication_group.this[0].member_clusters, null)
}

output "arn" {
  description = "Elasticache Replication Group ARN"
  value       = var.use_serverless ? try(aws_elasticache_serverless_cache.this[0].arn, null) : try(aws_elasticache_replication_group.this[0].arn, null)
}

output "cluster_enabled" {
  description = "Indicates if cluster mode is enabled."
  value       = try(aws_elasticache_replication_group.this[0].cluster_enabled, null)
}

output "id" {
  description = "ID of the ElastiCache Replication Group."
  value       = try(aws_elasticache_replication_group.this[0].id, null)
}

output "configuration_endpoint_address" {
  description = "Address of the replication group configuration endpoint when cluster mode is enabled."
  value       = try(aws_elasticache_replication_group.this[0].configuration_endpoint_address, null)
}

output "engine_version_actual" {
  description = "Because ElastiCache pulls the latest minor or patch for a version, this attribute returns the running version of the cache engine."
  value       = try(aws_elasticache_replication_group.this[0].engine_version_actual, null)
}

output "subnet_group_name" {
  description = "The Name of the ElastiCache Subnet Group."
  value       = try(aws_elasticache_subnet_group.this[0].name, var.subnet_group_name)
}

output "subnet_group_subnet_ids" {
  description = "The Subnet IDs of the ElastiCache Subnet Group."
  value       = try(aws_elasticache_subnet_group.this[0].subnet_ids, var.subnets)
}

output "parameter_group_arn" {
  description = "The AWS ARN associated with the parameter group."
  value       = try(aws_elasticache_parameter_group.this[0].arn, null)
}

output "parameter_group_id" {
  description = "The ElastiCache parameter group name."
  value       = try(aws_elasticache_parameter_group.this[0].id, null)
}
