variable "name" {
  type        = string
  description = "Name of the stack"
  default     = "complete-event-bridge-example"
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

variable "image" {
  type        = string
  description = "Name of image to pull from dockerhub"
  default     = "boldlink/flaskapp:latest"
}

variable "cpu" {
  type        = number
  description = "The number of cpu units to allocate"
  default     = 10
}

variable "memory" {
  type        = number
  description = "The size of memory to allocate in MiBs"
  default     = 512
}

variable "essential" {
  type        = bool
  description = "Whether this container is essential"
  default     = true
}

variable "containerport" {
  type        = number
  description = "Specify container port"
  default     = 5000
}

variable "hostport" {
  type        = number
  description = "Specify host port"
  default     = 5000
}

variable "network_mode" {
  type        = string
  description = "Docker networking mode to use for the containers in the task. Valid values are none, bridge, awsvpc, and host."
  default     = "awsvpc"
}

variable "service_ingress_rules" {
  description = "Ingress rules to add to the service security group."
  type        = list(any)
  default = [
    {
      from_port   = 5000
      to_port     = 5000
      protocol    = "tcp"
      description = "Allow traffic on port 5000. The app is configured to use this port"
      cidr_blocks = ["0.0.0.0/0"]
    }
  ]
}
