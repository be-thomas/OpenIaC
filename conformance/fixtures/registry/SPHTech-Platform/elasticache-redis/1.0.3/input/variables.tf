variable "enabled" {
  description = "Set to false to prevent the module from creating any resources"
  type        = bool
  default     = true
}

variable "replication_enabled" {
  description = "Set to false to diable replication in redis cluster"
  type        = bool
  default     = false
}

variable "cluster_mode_enabled" {
  description = "Set to false to diable cluster module"
  type        = bool
  default     = false
}

variable "name" {
  description = "Name of the application"
  type        = string
  default     = "value"
}

variable "engine" {
  description = "Cache engine to use. Valid values are 'redis' and 'valkey'."
  type        = string
  default     = "redis"

  validation {
    condition     = contains(["redis", "valkey"], var.engine)
    error_message = "engine must be either 'redis' or 'valkey'."
  }
}

variable "tags" {
  description = "Additional tags (_e.g._ map(\"BusinessUnit\",\"ABC\")"
  type        = map(string)
  default     = {}
}

variable "subnets" {
  description = "AWS subnet ids"
  type        = list(string)
  default     = []
}

variable "maintenance_window" {
  description = "Maintenance window"
  type        = string
  default     = "wed:03:00-wed:04:00"
}

variable "cluster_size" {
  description = "Cluster size"
  type        = number
  default     = 1
}

variable "instance_type" {
  description = "Elastic cache instance type"
  type        = string
  default     = "cache.t2.micro"
}

variable "engine_version" {
  description = "Redis engine version. https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/supported-engine-versions.html"
  type        = string
  default     = "7.0"
}

variable "alarm_cpu_threshold_percent" {
  description = "CPU threshold alarm level"
  type        = number
  default     = 75
}

variable "alarm_ecpu_threshold_percent" {
  description = "ECPU threshold alarm level for elasticache serverless"
  type        = number
  default     = 75
}

variable "alarm_memory_threshold_bytes" {
  description = "Alarm memory threshold bytes"
  type        = number
  default     = 10000000 # 10MB
}

variable "alarm_data_threshold_percent" {
  description = "Data threshold alarm level for elasticache serverless"
  type        = number
  default     = 75
}

variable "alarm_engine_cpu_threshold_percent" {
  description = "Engine CPU threshold alarm level"
  type        = number
  default     = 80
}

variable "alarm_evictions_threshold" {
  description = "Evictions threshold alarm level"
  type        = number
  default     = 0
}

variable "alarm_replication_lag_threshold_seconds" {
  description = "Replication lag threshold alarm level in seconds"
  type        = number
  default     = 5
}

variable "alarm_curr_connections_threshold" {
  description = "Current connections threshold alarm level. Adjust to suit the connection limit of the chosen instance type."
  type        = number
  default     = null
}

variable "notification_topic_arn" {
  description = "ARN of an SNS topic to send ElastiCache notifications"
  type        = string
  default     = ""
}

variable "alarm_actions" {
  description = "The list of actions to execute when this alarm transitions into an ALARM state from any other state."
  type        = list(string)
  default     = []
}

variable "ok_actions" {
  description = "The list of actions to execute when this alarm transitions into an OK state from any other state."
  type        = list(string)
  default     = []
}

variable "apply_immediately" {
  description = "Specifies whether any database modifications are applied immediately, or during the next maintenance window"
  type        = bool
  default     = true
}

variable "port" {
  description = "Redis port"
  type        = number
  default     = 6379
}

variable "security_groups" {
  description = "List of  Security Group IDs to place the cluster into"
  type        = list(string)
  default     = []
}

variable "subnet_group_name" {
  description = "Subnet group name for the ElastiCache instance"
  type        = string
  default     = ""
}

variable "elasticache_parameter_group_family" {
  description = "ElastiCache parameter group family. For redis use e.g. 'redis7'; for valkey use 'valkey7' or 'valkey8'."
  type        = string
  default     = "redis7"
}

variable "auto_minor_version_upgrade" {
  description = "Whether minor engine upgrades are applied automatically during the maintenance window. Supported on Redis/Valkey 6+."
  type        = bool
  default     = null
}

variable "auth_token_update_strategy" {
  description = "Strategy to use when updating the auth_token. Valid values are 'SET', 'ROTATE', and 'DELETE'. Leave null to use the provider default."
  type        = string
  default     = null

  validation {
    condition     = var.auth_token_update_strategy == null || contains(["SET", "ROTATE", "DELETE"], var.auth_token_update_strategy)
    error_message = "auth_token_update_strategy must be one of 'SET', 'ROTATE', or 'DELETE'."
  }
}

variable "data_tiering_enabled" {
  description = "Whether data tiering is enabled. Requires a r6gd node type."
  type        = bool
  default     = null
}

variable "log_delivery_configuration" {
  description = "List of log delivery configurations for the replication group (slow-log and/or engine-log)."
  type = list(object({
    destination      = string # CloudWatch log group name or Kinesis Firehose stream name
    destination_type = string # "cloudwatch-logs" or "kinesis-firehose"
    log_format       = string # "text" or "json"
    log_type         = string # "slow-log" or "engine-log"
  }))
  default = []

  validation {
    condition     = alltrue([for c in var.log_delivery_configuration : contains(["cloudwatch-logs", "kinesis-firehose"], c.destination_type)])
    error_message = "log_delivery_configuration destination_type must be 'cloudwatch-logs' or 'kinesis-firehose'."
  }

  validation {
    condition     = alltrue([for c in var.log_delivery_configuration : contains(["text", "json"], c.log_format)])
    error_message = "log_delivery_configuration log_format must be 'text' or 'json'."
  }

  validation {
    condition     = alltrue([for c in var.log_delivery_configuration : contains(["slow-log", "engine-log"], c.log_type)])
    error_message = "log_delivery_configuration log_type must be 'slow-log' or 'engine-log'."
  }
}

variable "replication_group_id" {
  description = "ElastiCache replication_group_id"
  type        = string
  default     = ""
}

variable "parameters" {
  description = "A list of Redis parameters to apply. Note that parameters may differ from one Redis family to another"
  type = list(object({
    name  = string
    value = string
  }))
  default = []
}

variable "cluster_id" {
  description = "Cluster ID"
  type        = string
  default     = null
}

variable "create_elasticache_subnet_group" {
  description = "Create Elasticache Subnet Group"
  type        = bool
  default     = true
}

variable "preferred_cache_cluster_azs" {
  description = "List of EC2 availability zones in which the replication group's cache clusters will be created. The order of the availability zones in the list is considered. The first item in the list will be the primary node. Ignored when updating"
  type        = list(string)
  default = [
    "ap-southeast-1a",
    "ap-southeast-1b",
  ]
}

variable "parameter_group_name" {
  description = "Existing Parameter Group name"
  type        = string
  default     = ""
}

variable "snapshot_retention_limit" {
  description = "Number of days for which ElastiCache will retain automatic cache cluster snapshots before deleting them. For example, if you set SnapshotRetentionLimit to 5, then a snapshot that was taken today will be retained for 5 days before being deleted. If the value of snapshot_retention_limit is set to zero (0), backups are turned off. Please note that setting a snapshot_retention_limit is not supported on cache.t1.micro cache nodes"
  type        = number
  default     = 5
}

variable "transit_encryption_enabled" {
  description = "Whether to enable encryption in transit"
  type        = string
  default     = true
}

variable "transit_encryption_mode" {
  description = "Accepted values are preferred and required. Can be specified only if `transit_encryption_enabled = true`"
  type        = string
  default     = null
}

variable "auth_token" {
  description = "Password used to access a password protected server. Can be specified only if `transit_encryption_enabled = true`"
  type        = string
  default     = null
}

variable "at_rest_encryption_enabled" {
  description = "Whether to enable encryption at rest"
  type        = string
  default     = true
}

variable "kms_key_id" {
  description = "The ARN of the key that you wish to use if encrypting at rest. If not supplied, uses service managed encryption. Can be specified only if `at_rest_encryption_enabled = true`"
  type        = string
  default     = null
}

variable "num_node_groups" {
  description = "Number of node groups (shards) for this Redis replication group. Changing this number will trigger an online resizing operation before other settings modifications. Required unless `global_replication_group_id` is set"
  type        = number
  default     = 2
}

variable "replicas_per_node_group" {
  description = "Number of replica nodes in each node group. Valid values are 0 to 5. Changing this number will trigger an online resizing operation before other settings modifications."
  type        = number
  default     = 1
}

# ElastiCache Serverless
variable "use_serverless" {
  description = "Use serverless ElastiCache service"
  type        = bool
  default     = false
}

variable "max_data_storage" {
  type        = number
  description = "The maximun cached data capacity of the Serverless Cache in GB"
  default     = 10

  validation {
    condition     = var.max_data_storage >= 1 && var.max_data_storage <= 5000
    error_message = "The max_data_storage in GB value must be between 1 and 5,000."
  }
}

variable "max_ecpu_per_second" {
  type        = number
  description = "The maximum ECPU per second of the Serverless Cache"
  default     = 1000

  validation {
    condition     = var.max_ecpu_per_second >= 1000 && var.max_ecpu_per_second <= 15000000
    error_message = "The max_ecpu_per_second value must be between 1,000 and 15,000,000."
  }
}

variable "daily_snapshot_time" {
  type        = string
  description = "The daily time range (in UTC) during which the service takes automatic snapshot of the Serverless Cache"
  default     = "18:00"
}

variable "snapshot_arns_to_restore" {
  type        = list(string)
  description = "The ARN's of snapshot to restore Serverless Cache"
  default     = []
}

variable "user_group_id" {
  type        = string
  description = "The ID of the user group Elasticache"
  default     = ""
}

variable "snapshot_window" {
  type        = string
  description = "The daily time range (in UTC) during which ElastiCache begins taking a daily snapshot of the node group (shard) specified by SnapshottingClusterId"
  default     = "00:00-01:00"
}

variable "snapshot_arns" {
  type        = list(string)
  description = "The ARN of the snapshot from which to restore data into the new node group (shard)"
  default     = []
}

variable "snapshot_name" {
  type        = string
  description = "The name of the snapshot from which to restore data into the new node group (shard)"
  default     = ""
}
