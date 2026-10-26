output "iam_role_arn_role" {
  value       = [aws_iam_role.task_role[*].arn]
  description = "The ARN of IAM task role"
}

output "iam_role_id_role" {
  value       = [aws_iam_role.task_role[*].id]
  description = "ID of IAM task role"
}

output "iam_role_name_role" {
  value       = [aws_iam_role.task_role[*].name]
  description = "Name of IAM task role"
}

output "iam_role_create_date_role" {
  value       = [aws_iam_role.task_role[*].create_date]
  description = "Creation date of IAM task role"
}

output "iam_role_unique_id_role" {
  value       = [aws_iam_role.task_role[*].unique_id]
  description = "Unique ID of IAM task role "
}

# Cloudwatch
output "cloudwatch_log_group_arn" {
  value       = aws_cloudwatch_log_group.main.arn
  description = "ARN of cloudwatch log group"
}

output "lb_dns_name" {
  value       = local.lb_dns_name
  description = "DNS name of load balancer"
}

output "lb_dns_zone_id" {
  value       = local.lb_dns_zone_id
  description = "DNS zone id of load balancer"
}

output "lb_arn" {
  value       = [aws_lb.main[*].arn]
  description = "The ARN of the load balancer (matches `id`)"
}

# task definition
output "task_definition_arn" {
  value       = [aws_ecs_task_definition.this[*].arn]
  description = "Full ARN of the Task Definition (including both family and revision)"
}

output "task_definition_arn_without_revision" {
  value       = [aws_ecs_task_definition.this[*].arn_without_revision]
  description = "ARN of the Task Definition with the trailing `revision` removed. This may be useful for situations where the latest task definition is always desired. If a revision isn't specified, the latest ACTIVE revision is used."
}

output "task_definition_revision" {
  value       = [aws_ecs_task_definition.this[*].revision]
  description = "Revision of the task in a particular family."
}

# Security groups
output "lb_sg_id" {
  value       = [aws_security_group.lb[*].id]
  description = "ARN of the load balancer security group."
}

output "lb_sg_arn" {
  value       = [aws_security_group.lb[*].arn]
  description = "ID of the load balancer security group."
}

output "service_sg_id" {
  value       = aws_security_group.service.id
  description = "ARN of the service security group."
}

output "service_sg_arn" {
  value       = aws_security_group.service.arn
  description = "ID of the service security group."
}
