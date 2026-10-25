
variable "primary_cluster" {
  description = "Confirms if this will be a primary cluster: If not, `database_name`, `master_username` and `master_password` are not required for `secondary` clusters with a `global_cluster_identifier` defined."
  type        = bool
  default     = true
}

variable "database_name" {
  description = "(Optional) Name for an automatically created database on cluster creation."
  type        = string
  default     = null
}

variable "master_password" {
  description = "(Required unless a snapshot_identifier or replication_source_identifier is provided or unless a global_cluster_identifier is provided when the cluster is the 'secondary' cluster of a global database) Password for the master DB user."
  type        = string
  default     = null
}

variable "manage_master_user_password" {
  description = "(Optional) Set to true to allow RDS to manage the master user password in Secrets Manager. Cannot be set if master_password is provided."
  type        = bool
  default     = true
}

variable "master_user_secret_kms_key_id" {
  description = " (Optional) Amazon Web Services KMS key identifier is the key ARN, key ID, alias ARN, or alias name for the KMS key. To use a KMS key in a different Amazon Web Services account, specify the key ARN or alias ARN. If not specified, the default KMS key for your Amazon Web Services account is used."
  type        = string
  default     = null
}

variable "master_username" {
  description = "(Required unless a snapshot_identifier or replication_source_identifier is provided or unless a global_cluster_identifier is provided when the cluster is the 'secondary' cluster of a global database) Username for the master DB user."
  type        = string
  default     = null
}

variable "allow_major_version_upgrade" {
  description = "(Optional) Enable to allow major engine version upgrades when changing engine versions. Defaults to false."
  type        = bool
  default     = false
}

variable "apply_immediately" {
  description = "(Optional) Specifies whether any cluster modifications are applied immediately, or during the next maintenance window. Default is `false`."
  type        = bool
  default     = false
}

variable "availability_zones" {
  description = "(Optional) A list of EC2 Availability Zones for the DB cluster storage where DB cluster instances can be created. RDS automatically assigns 3 AZs if less than 3 AZs are configured, which will show as a difference requiring resource recreation next Terraform apply. It is recommended to specify 3 AZs or use the lifecycle configuration block ignore_changes argument if necessary."
  type        = list(string)
  default     = []
}

variable "backtrack_window" {
  description = "(Optional) The target backtrack window, in seconds. Only available for aurora and aurora-mysql engines currently. To disable backtracking, set this value to 0. Defaults to 0. Must be between 0 and 259200 (72 hours)"
  type        = number
  default     = 0
}

variable "backup_retention_period" {
  description = "(Optional) The days to retain backups for. Default 1"
  type        = number
  default     = 1
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

variable "copy_tags_to_snapshot" {
  description = "(Optional, boolean) Copy all Cluster tags to snapshots. Default is `false`."
  type        = bool
  default     = true
}

variable "db_cluster_parameter_group_name" {
  description = "(Optional) A cluster parameter group to associate with the cluster."
  type        = string
  default     = null
}

variable "db_instance_parameter_group_name" {
  description = "(Optional) Instance parameter group to associate with all instances of the DB cluster. The db_instance_parameter_group_name parameter is only valid in combination with the allow_major_version_upgrade parameter."
  type        = string
  default     = null
}

variable "db_subnet_group_name" {
  description = "(Optional) A DB subnet group to associate with this DB instance. NOTE: This must match the db_subnet_group_name specified on every aws_rds_cluster_instance in the cluster."
  type        = string
  default     = null
}

variable "deletion_protection" {
  description = "(Optional) If the DB instance should have deletion protection enabled. The database can't be deleted when this value is set to true. The default is false."
  type        = bool
  default     = false
}

variable "enable_http_endpoint" {
  description = "(Optional) Enable HTTP endpoint (data API). Only valid when engine_mode is set to serverless."
  type        = bool
  default     = false
}

variable "enabled_cloudwatch_logs_exports" {
  description = "(Optional) Set of log types to export to cloudwatch. If omitted, no logs will be exported. The following log types are supported: audit, error, general, slowquery, postgresql (PostgreSQL)"
  type        = list(string)
  default     = []
}

variable "engine" {
  description = "(Required) The name of the database engine to be used for this DB cluster. Defaults to aurora. Valid Values: aurora, aurora-mysql, aurora-postgresql"
  type        = string
}

variable "engine_mode" {
  description = "(Optional) The database engine mode. Valid values: global (only valid for Aurora MySQL 1.21 and earlier), multimaster, parallelquery, provisioned, serverless. Defaults to: provisioned. "
  type        = string
  default     = null
}

variable "engine_version" {
  description = "(Optional) The database engine version. Updating this argument results in an outage. The actual engine version used is returned in the attribute engine_version_actual."
  type        = string
  default     = null
}

variable "final_snapshot_identifier" {
  description = "(Optional) The name of your final DB snapshot when this DB cluster is deleted. If omitted, no final snapshot will be made."
  type        = string
  default     = null
}

variable "global_cluster_identifier" {
  description = "(Optional) The global cluster identifier specified on `aws_rds_global_cluster`."
  type        = string
  default     = null
}

variable "enable_global_write_forwarding" {
  description = "(Optional) Whether cluster should forward writes to an associated global cluster."
  type        = bool
  default     = false
}

variable "iam_database_authentication_enabled" {
  description = "(Optional) Specifies whether or not mappings of AWS Identity and Access Management (IAM) accounts to database accounts is enabled."
  type        = bool
  default     = false
}

variable "iam_roles" {
  description = "(Optional) A List of ARNs for the IAM roles to associate to the RDS Cluster."
  type        = list(string)
  default     = []
}

variable "kms_key_id" {
  description = "(Optional) The ARN for the KMS encryption key. When specifying kms_key_id, storage_encrypted needs to be set to true."
  type        = string
  default     = null
}

variable "port" {
  description = "(Optional) The port on which the DB accepts connections"
  type        = number
  default     = null
}

variable "preferred_backup_window" {
  description = "(Optional) The daily time range during which automated backups are created if automated backups are enabled using the BackupRetentionPeriod parameter.Time in UTC. Default: A 30-minute window selected at random from an 8-hour block of time per region"
  type        = string
  default     = "00:00-01:00"
}

variable "preferred_maintenance_window" {
  description = "(Optional) The weekly time range during which system maintenance can occur, in (UTC) e.g., wed:04:00-wed:04:30"
  type        = string
  default     = "Sun:02:00-Sun:04:00"
}

variable "replication_source_identifier" {
  description = "(Optional) ARN of a source DB cluster or DB instance if this DB cluster is to be created as a Read Replica."
  type        = string
  default     = null
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

variable "source_region" {
  description = "(Optional) The source region for an encrypted replica DB cluster."
  type        = string
  default     = null
}

variable "storage_encrypted" {
  description = "(Optional) Specifies whether the DB cluster is encrypted. The default is false for provisioned engine_mode and true for serverless engine_mode. When restoring an unencrypted snapshot_identifier, the kms_key_id argument must be provided to encrypt the restored cluster. "
  type        = bool
  default     = false
}

variable "tags" {
  description = "(Optional)A map of tags to assign to the resources"
  type        = map(string)
  default     = {}
}

#S3 import

/*variable "s3_import" {
  description = " Requires that the S3 bucket be in the same region as the RDS cluster you're trying to create."
  type        = map(string)
  default     = null
}
*/

# Restore to point in time
variable "restore_to_point_in_time" {
  description = "(Optional) Nested attribute for point in time restore. "
  type        = map(any)
  default     = null
}

variable "cluster_timeouts" {
  description = "aws_rds_cluster provides the following Timeouts configuration options: create, update, delete"
  type        = map(string)
  default     = {}
}

variable "scaling_configuration" {
  description = "(Optional) Nested attribute with scaling properties. Only valid when engine_mode is set to serverless"
  type        = map(string)
  default     = {}
}

# Cluster Instance
variable "instance_count" {
  description = "Number of DocumentDB cluster instances to be created."
  type        = number
  default     = 2
}

variable "instance_class" {
  description = "(Required) The instance class to use. For details on CPU and memory, see Scaling Aurora DB Instances. Aurora uses db.* instance classes/types."
  type        = string
  default     = ""
}

variable "publicly_accessible" {
  description = "(Optional) Bool to control if instance is publicly accessible. Default false."
  type        = bool
  default     = false
}

variable "monitoring_role_arn" {
  description = "(Optional) The ARN for the IAM role that permits RDS to send enhanced monitoring metrics to CloudWatch Logs."
  type        = string
  default     = null
}

variable "monitoring_interval" {
  description = "(Optional) The interval, in seconds, between points when Enhanced Monitoring metrics are collected for the DB instance. To disable collecting Enhanced Monitoring metrics, specify 0. The default is 0. Valid Values: 0, 1, 5, 10, 15, 30, 60."
  type        = number
  default     = 0
}

variable "promotion_tier" {
  description = "(Optional) Default 0. Failover Priority setting on instance level. The reader who has lower tier has higher priority to get promoted to writer."
  type        = number
  default     = 0
}

variable "auto_minor_version_upgrade" {
  description = "(Optional) Indicates that minor engine upgrades will be applied automatically to the DB instance during the maintenance window. Default true."
  type        = bool
  default     = true
}

variable "availability_zone" {
  description = "(Optional, Computed, Forces new resource) The EC2 Availability Zone that the DB instance is created in."
  type        = string
  default     = null
}

variable "performance_insights_enabled" {
  description = "(Optional) Specifies whether Performance Insights is enabled or not."
  type        = bool
  default     = false
}

variable "performance_insights_kms_key_id" {
  description = "(Optional) ARN for the KMS key to encrypt Performance Insights data. When specifying performance_insights_kms_key_id, performance_insights_enabled needs to be set to true."
  type        = string
  default     = null
}

variable "performance_insights_retention_period" {
  description = "(Optional) Amount of time in days to retain Performance Insights data. Either 7 (7 days) or 731 (2 years). When specifying performance_insights_retention_period, performance_insights_enabled needs to be set to true. Defaults to '7'."
  type        = number
  default     = 7
}

variable "ca_cert_identifier" {
  description = "(Optional) The identifier of the CA certificate for the DB instance."
  type        = string
  default     = null
}

variable "instance_timeouts" {
  description = "aws_rds_cluster_instance provides the following Timeouts configuration options: create, update, delete"
  type        = map(string)
  default     = {}
}

# Subnet Group
variable "subnet_ids" {
  description = "(Required) A list of VPC subnet IDs."
  type        = list(string)
  default     = []
}

variable "create_db_subnet_group" {
  description = "Whether to create the Subnet Group or not"
  type        = bool
  default     = true
}

# Security Group

variable "vpc_id" {
  description = "(Optional, Forces new resource) VPC ID"
  type        = string
  default     = null
}

variable "ingress_rules" {
  description = "(Optional) Ingress rules to add to the security group"
  type        = any
  default     = {}
}

variable "egress_rules" {
  description = "(Optional) Egress rules to add to the security group"
  type        = any
  default     = {}
}

variable "create_security_group" {
  description = "Whether to create a Security Group for RDS cluster."
  default     = true
  type        = bool
}

# Cluster Parameter Group

variable "create_cluster_parameter_group" {
  description = "Whether to create a cluster_parameter_group or not"
  type        = bool
  default     = false
}

variable "name_prefix" {
  description = "(Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with name."
  type        = string
  default     = null
}

variable "family" {
  description = "(Required) The family of the DB cluster parameter group."
  type        = string
  default     = ""
}

variable "cluster_parameters" {
  description = "(Optional) A list of DB parameters to apply. Note that parameters may differ from a family to an other."
  type        = list(any)
  default     = []
}

# Enhanced Monitoring
variable "create_monitoring_role" {
  description = "Whether to create monitoring role or not"
  type        = bool
  default     = false
}

variable "assume_role_policy" {
  description = "(Required) Policy that grants an entity permission to assume the role."
  type        = string
  default     = ""
}

variable "policy_arn" {
  description = " (Required) - The ARN of the policy you want to apply"
  type        = string
  default     = ""
}

# Cluster Endpoint

variable "create_cluster_endpoint" {
  description = "Whether to create a cluster endpoint or not"
  type        = bool
  default     = false
}

variable "custom_endpoint_type" {
  description = "(Required) The type of the endpoint. One of: READER , ANY ."
  type        = string
  default     = "READER"
}

# Autoscaling
variable "enable_autoscaling" {
  description = "Whether to enable autoscaling or not"
  type        = bool
  default     = false
}

variable "max_capacity" {
  description = "Maximum number of replicas"
  type        = number
  default     = 1
}

variable "min_capacity" {
  description = "Minimum number of replicas"
  type        = number
  default     = 1
}

variable "scalable_dimension" {
  description = "(Required) The scalable dimension of the scalable target."
  type        = string
  default     = ""
}

variable "service_namespace" {
  description = "(Required) The AWS service namespace of the scalable target."
  type        = string
  default     = "rds"
}

variable "policy_type" {
  description = "Optional) The policy type. Valid values are StepScaling and TargetTrackingScaling. Defaults to StepScaling."
  type        = string
  default     = "StepScaling"
}

variable "predefined_metric_type" {
  description = "The type of metric to scale on"
  type        = string
  default     = ""
}

variable "target_value" {
  description = "(Required) The target value for the metric."
  type        = number
  default     = 75
}

variable "scale_in_cooldown" {
  description = " (Optional) The amount of time, in seconds, after a scale in activity completes before another scale in activity can start."
  type        = number
  default     = 300
}

variable "scale_out_cooldown" {
  description = "(Optional) The amount of time, in seconds, after a scale out activity completes before another scale out activity can start."
  type        = number
  default     = 300
}

variable "vpc_security_group_ids" {
  description = "List of VPC security groups to associate"
  type        = list(string)
  default     = []
}
