
variable "create_global_cluster" {
  description = "Whether you are creating a global cluster or not."
  type        = bool
  default     = false
}

variable "global_cluster_identifier" {
  description = "(Required, Forces new resources) Global cluster identifier."
  type        = string
  default     = ""
}

variable "database_name" {
  description = "(Optional, Forces new resources) Name for an automatically created database on cluster creation."
  type        = string
  default     = null
}

variable "deletion_protection" {
  description = "(Optional) If the Global Cluster should have deletion protection enabled. The database can't be deleted when this value is set to true. The default is false."
  type        = bool
  default     = false
}

variable "engine" {
  description = "(Optional, Forces new resources) Name of the database engine to be used for this DB cluster. Terraform will only perform drift detection if a configuration value is provided. Valid values: aurora, aurora-mysql, aurora-postgresql. Defaults to aurora. Conflicts with source_db_cluster_identifier."
  type        = string
  default     = "aurora"
}

variable "engine_version" {
  description = "(Optional) Engine version of the Aurora global database. The engine, engine_version, and instance_class (on the aws_rds_cluster_instance) must together support global databases."
  type        = string
  default     = null
}

variable "force_destroy" {
  description = "(Optional) Enable to remove DB Cluster members from Global Cluster on destroy. Required with source_db_cluster_identifier."
  type        = bool
  default     = true
}

variable "source_db_cluster_identifier" {
  description = "(Optional) Amazon Resource Name (ARN) to use as the primary DB Cluster of the Global Cluster on creation. "
  type        = string
  default     = null
}

variable "storage_encrypted" {
  description = "(Optional, Forces new resources) Specifies whether the DB cluster is encrypted. The default is false unless source_db_cluster_identifier is specified and encrypted."
  type        = bool
  default     = false
}
