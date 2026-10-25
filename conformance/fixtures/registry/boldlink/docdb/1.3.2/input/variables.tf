
# DOcDB Cluster
variable "apply_immediately" {
  description = "(Optional) Specifies whether any cluster or database modifications are applied immediately, or during the next maintenance window. Default is false."
  type        = string
  default     = false
}

variable "availability_zones" {
  description = "(Optional) A list of EC2 Availability Zones that instances in the DB cluster can be created in."
  type        = list(string)
  default     = []
}

variable "backup_retention_period" {
  description = "(Optional) The days to retain backups for. Default 1"
  type        = number
  default     = 7
}

variable "cluster_identifier_prefix" {
  description = "(Optional, Forces new resource) Creates a unique cluster identifier beginning with the specified prefix. Conflicts with cluster_identifier."
  type        = string
  default     = null
}

variable "cluster_identifier" {
  description = "(Optional, Forces new resources) The cluster identifier. If omitted, Terraform will assign a random, unique identifier."
  type        = string
  default     = null
}

variable "db_cluster_parameter_group_name" {
  description = "(Optional) A cluster parameter group to associate with the cluster."
  type        = string
  default     = null
}

variable "deletion_protection" {
  description = "(Optional) A value that indicates whether the DB cluster has deletion protection enabled. The database can't be deleted when deletion protection is enabled. By default, deletion protection is disabled."
  type        = bool
  default     = false
}

variable "enabled_cloudwatch_logs_exports" {
  description = "(Optional) List of log types to export to cloudwatch. If omitted, no logs will be exported. The following log types are supported: audit, profiler"
  type        = list(string)
  default     = ["audit", "profiler"]
}

variable "engine_version" {
  description = "(Optional) The database engine version. Updating this argument results in an outage."
  type        = string
  default     = null
}

variable "engine" {
  description = "(Optional) The name of the database engine to be used for this DB cluster and instance. Defaults to docdb. Valid Values: docdb"
  type        = string
  default     = "docdb"
}

variable "kms_key_id" {
  description = "(Optional) The ARN for the KMS encryption key. When specifying kms_key_id, storage_encrypted needs to be set to true."
  type        = string
  default     = null
}
variable "master_password" {
  description = "(Required unless a snapshot_identifier or unless a global_cluster_identifier is provided when the cluster is the 'secondary' cluster of a global database) Password for the master DB user. Note that this may show up in logs, and it will be stored in the state file."
  type        = string
}

variable "master_username" {
  description = "(Required unless a snapshot_identifier or unless a global_cluster_identifier is provided when the cluster is the 'secondary' cluster of a global database) Username for the master DB user"
  type        = string
}
variable "port" {
  description = "(Optional) The port on which the DB accepts connections"
  type        = number
  default     = 27017
}

variable "preferred_backup_window" {
  description = "(Optional) The daily time range during which automated backups are created if automated backups are enabled using the BackupRetentionPeriod parameter.Time in UTC Default: A 30-minute window selected at random from an 8-hour block of time per regionE.g., 04:00-09:00"
  type        = string
  default     = "04:00-05:00"
}

variable "preferred_maintenance_window" {
  description = "(Optional) The weekly time range during which system maintenance can occur, in (UTC) e.g., wed:04:00-wed:04:30"
  type        = string
  default     = "sun:01:00-sun:03:30"
}

variable "skip_final_snapshot" {
  description = "(Optional) Determines whether a final DB snapshot is created before the DB cluster is deleted. If true is specified, no DB snapshot is created. If false is specified, a DB snapshot is created before the DB cluster is deleted, using the value from final_snapshot_identifier. Default is false."
  type        = bool
  default     = false
}

variable "snapshot_identifier" {
  description = "(Optional) Specifies whether or not to create this cluster from a snapshot. You can use either the name or ARN when specifying a DB cluster snapshot, or the ARN when specifying a DB snapshot."
  type        = string
  default     = null
}

variable "storage_encrypted" {
  description = "(Optional) Specifies whether the DB cluster is encrypted. The default is false."
  type        = bool
  default     = true
}

variable "vpc_security_group_ids" {
  description = "(Optional) List of VPC security groups to associate with the Cluster"
  type        = list(string)
  default     = []

}

#Tags
variable "tags" {
  description = " (Optional) A map of tags to assign to the DB cluster. If configured with a provider default_tags configuration block present, tags with matching keys will overwrite those defined at the provider-level."
  type        = map(string)
  default     = {}
}

variable "cluster_timeouts" {
  description = "aws_docdb_cluster provides the following Timeouts configuration options: create, update, delete"
  type        = map(string)
  default     = {}
}

# DocDB SubnetGroup
variable "subnet_name_prefix" {
  description = "(Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with name."
  type        = string
  default     = null
}

variable "subnet_ids" {
  description = "(Required) A list of VPC subnet IDs."
  default     = []
  type        = list(string)
}

# Cluster Instance
variable "instance_count" {
  description = "Number of DocumentDB cluster instances to be created."
  type        = number
  default     = 3
}
variable "instance_class" {
  description = "(Required) The instance class to use. For details on CPU and memory, see Scaling for DocDB Instances. db.r5.large, db.r5.xlarge ,db.r5.2xlarge, db.r5.4xlarge, db.r5.12xlarge, db.r5.24xlarge, db.r4.large, db.r4.xlarge, db.r4.2xlarge, db.r4.4xlarge, db.r4.8xlarge, db.r4.16xlarge, db.t3.medium"
  type        = string
  default     = "db.t3.medium"
}

variable "auto_minor_version_upgrade" {
  description = "(Optional) Indicates that minor engine upgrades will be applied automatically to the DB instance during the maintenance window. Default true"
  type        = bool
  default     = true
}

variable "availability_zone" {
  description = " (Optional, Computed) The EC2 Availability Zone that the DB instance is created in."
  type        = string
  default     = null
}

variable "identifier" {
  description = "(Optional, Forces new resource) The identifier for the DocDB instance, if omitted, Terraform will assign a random, unique identifier."
  type        = string
  default     = null
}

variable "identifier_prefix" {
  description = "(Optional, Forces new resource) Creates a unique identifier beginning with the specified prefix. Conflicts with identifier"
  type        = string
  default     = null
}

variable "promotion_tier" {
  description = "(Optional) Default 0. Failover Priority setting on instance level. The reader who has lower tier has higher priority to get promoter to writer."
  type        = number
  default     = 0
}

variable "instance_timeouts" {
  description = "aws_docdb_cluster_instance provides the following Timeouts configuration options: create, update, delete"
  type        = map(string)
  default     = {}
}

# Cluster Parameter Group
variable "create_cluster_parameter_group" {
  description = "Whether to create cluster parameter group"
  type        = bool
  default     = false
}

variable "name" {
  description = "(Optional, Forces new resource) The name of the documentDB cluster parameter group. If omitted, Terraform will assign a random, unique name."
  type        = string
  default     = null
}

variable "name_prefix" {
  description = "(Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with name."
  type        = string
  default     = null
}

variable "family" {
  description = "(Required, Forces new resource) The family of the documentDB cluster parameter group."
  type        = string
  default     = "docdb4.0"
}

variable "cluster_parameters" {
  description = "(Optional) A list of documentDB parameters to apply. Setting parameters to system default values may show a difference on imported resources."
  type = list(object({
    name         = string
    value        = string
    apply_method = string
  }))
  default = []
}

# Security Group
variable "vpc_id" {
  description = "(Optional, Forces new resource) VPC ID. Defaults to the region's default VPC."
  type        = string
  default     = null
}

variable "create_security_group" {
  description = "Whether to create a Security Group for DocDB cluster."
  default     = true
  type        = bool
}

variable "security_group_ingress_rules" {
  description = "(Optional) Ingress rules to add to the security group"
  type        = any
  default     = {}
}
variable "security_group_egress_rules" {
  description = "(Optional) Egress rules to add to the security group"
  type        = any
  default     = {}
}
