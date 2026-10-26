output "address" {
  value       = aws_db_instance.this.address
  description = "The hostname of the RDS instance. See also endpoint and port"
}

output "arn" {
  value       = aws_db_instance.this.arn
  description = "The ARN of the RDS instance."
}

output "allocated_storage" {
  value       = aws_db_instance.this.allocated_storage
  description = "The amount of allocated storage."
}


output "availability_zone" {
  value       = aws_db_instance.this.availability_zone
  description = "The availability zone of the instance."
}

output "backup_retention_period" {
  value       = aws_db_instance.this.backup_retention_period
  description = "The backup retention period."
}

output "backup_window" {
  value       = aws_db_instance.this.backup_window
  description = "The backup window."
}

output "ca_cert_identifier" {
  value       = aws_db_instance.this.ca_cert_identifier
  description = " Specifies the identifier of the CA certificate for the DB instance."
}

output "endpoint" {
  value       = aws_db_instance.this.endpoint
  description = "The connection endpoint in address:port format."
}

output "identifier" {
  description = "The RDS instance identifier"
  value       = aws_db_instance.this.identifier
}

output "engine" {
  value       = aws_db_instance.this.engine
  description = "The database engine."
}

output "engine_version_actual" {
  value       = aws_db_instance.this.engine_version
  description = "The running version of the database."
}

output "hosted_zone_id" {
  value       = aws_db_instance.this.hosted_zone_id
  description = "The canonical hosted zone ID of the DB instance (to be used in a Route 53 Alias record)."
}

output "id" {
  value       = aws_db_instance.this.id
  description = "The RDS instance ID."
}

output "instance_class" {
  value       = aws_db_instance.this.instance_class
  description = "The RDS instance class."
}

output "latest_restorable_time" {
  description = "The latest time, in UTC RFC3339 format, to which a database can be restored with point-in-time restore."
  value       = aws_db_instance.this.latest_restorable_time
}

output "maintenance_window" {
  value       = aws_db_instance.this.maintenance_window
  description = "The instance maintenance window."
}

output "multi_az" {
  value       = aws_db_instance.this.multi_az
  description = "If the RDS instance is multi AZ enabled."
}

output "db_name" {
  value       = aws_db_instance.this.db_name
  description = "The database name."
}

output "domain" {
  description = "The ID of the Directory Service Active Directory domain the instance is joined to"
  value       = aws_db_instance.this.domain
}

output "port" {
  value       = aws_db_instance.this.port
  description = "The database port."
}

output "resource_id" {
  value       = aws_db_instance.this.resource_id
  description = "The RDS Resource ID of this instance."
}

output "status" {
  value       = aws_db_instance.this.status
  description = "The RDS instance status."
}

output "storage_encrypted" {
  value       = aws_db_instance.this.storage_encrypted
  description = "Specifies whether the DB instance is encrypted."
}

output "username" {
  value       = aws_db_instance.this.username
  description = "The master username for the database."
}

output "tags_all" {
  description = "A map of tags assigned to the resource, including those inherited from the provider default_tags"
  value       = aws_db_instance.this.tags_all
}

# Additional output for MsSQL
output "character_set_name" {
  description = "The character set (collation) used on Oracle and Microsoft SQL instances."
  value       = aws_db_instance.this.character_set_name
}

# Security Group
output "sg_arn" {
  description = "ARN of the security group."
  value       = aws_security_group.this.*.arn
}

output "sg_id" {
  description = " ID of the security group.."
  value       = aws_security_group.this.*.id
}
