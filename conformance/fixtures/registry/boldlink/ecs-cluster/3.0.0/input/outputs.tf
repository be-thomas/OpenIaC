
output "arn" {
  description = "ARN that identifies the cluster."
  value       = aws_ecs_cluster.this.arn
}

output "id" {
  description = "ARN that identifies the cluster."
  value       = aws_ecs_cluster.this.id
}

output "tags_all" {
  description = "Map of tags assigned to the resource, including those inherited from the provider default_tags"
  value       = aws_ecs_cluster.this.tags_all
}

output "key_arn" {
  description = "The Amazon Resource Name (ARN) of the key."
  value       = aws_kms_key.main.*.arn
}

output "key_id" {
  description = " The globally unique identifier for the key."
  value       = aws_kms_key.main.*.key_id
}

output "autoscaling_group_arn" {
  description = "ARN of the Auto Scaling Group"
  value       = try(aws_autoscaling_group.container_instance[0].arn, null)
}

output "autoscaling_group_name" {
  description = "Name of the Auto Scaling Group"
  value       = try(aws_autoscaling_group.container_instance[0].name, null)
}
