output "arn" {
  description = "RDS Global Cluster Amazon Resource Name (ARN)"
  value       = aws_rds_global_cluster.this.*.arn
}

output "global_cluster_members" {
  description = "Set of objects containing Global Cluster members."
  value       = aws_rds_global_cluster.this.*.global_cluster_members
}

output "global_cluster_resource_id" {
  description = "AWS Region-unique, immutable identifier for the global database cluster. This identifier is found in AWS CloudTrail log entries whenever the AWS KMS key for the DB cluster is accessed"
  value       = aws_rds_global_cluster.this.*.global_cluster_resource_id
}

output "id" {
  description = "RDS Global Cluster identifier"
  value       = aws_rds_global_cluster.this.*.id
}
