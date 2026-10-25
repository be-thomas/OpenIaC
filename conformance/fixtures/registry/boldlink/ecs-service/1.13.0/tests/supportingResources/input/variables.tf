variable "name" {
  type        = string
  description = "Name of the stack"
  default     = "terraform-aws-ecs-service"
}

variable "cidr_block" {
  type        = string
  description = "VPC CIDR"
  default     = "10.1.0.0/16"
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply to the created resources"
  default = {
    Environment        = "examples"
    "user::CostCenter" = "terraform-registry"
    Department         = "DevOps"
    InstanceScheduler  = true
    Project            = "Examples"
    Owner              = "Boldlink"
    LayerName          = "cExample"
    LayerId            = "cExample"
  }
}

variable "enable_dns_hostnames" {
  type        = bool
  description = "Whether to enable dns hostnames"
  default     = true
}

variable "enable_dns_support" {
  type        = bool
  description = "Whether to enable dns support for the vpc"
  default     = true
}

variable "enable_public_subnets" {
  type        = bool
  description = "Whether to enable public subnets"
  default     = true
}

variable "enable_private_subnets" {
  type        = bool
  description = "Whether to enable private subnets"
  default     = true
}

variable "map_public_ip_on_launch" {
  type        = bool
  description = "Whether assign public IPs by default to instances launched on subnet"
  default     = true
}

variable "nat" {
  type        = string
  description = "Choose `single` or `multi` for NATs"
  default     = "single"
}

variable "kms_description" {
  type        = string
  description = "Description of what kms key does"
  default     = "A test kms key for ecs cluster"
}

variable "create_kms_alias" {
  type        = bool
  description = "Whether to create kms alias"
  default     = true
}

variable "enable_key_rotation" {
  type        = bool
  description = "Whether to create kms alias"
  default     = true
}

variable "deletion_window_in_days" {
  type        = number
  description = "Number of days before key is deleted"
  default     = 7
}

variable "retention_in_days" {
  type        = number
  description = "Period to retain logs in log group"
  default     = 1
}

variable "cloud_watch_encryption_enabled" {
  type        = bool
  description = "Whether to enable cloudwatch encryption"
  default     = true
}

variable "s3_bucket_encryption_enabled" {
  type        = bool
  description = "Whether encryption is enabled for the s3 bucket"
  default     = false
}

variable "logging_type" {
  type        = string
  description = "Specify the type of ecs cluster logging"
  default     = "OVERRIDE"
}
