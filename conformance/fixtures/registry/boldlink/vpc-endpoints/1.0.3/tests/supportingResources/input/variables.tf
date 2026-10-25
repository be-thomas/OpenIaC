variable "name" {
  type        = string
  description = "Name of the stack"
  default     = "terraform-aws-vpc-endpoints"
}

variable "cidr_block" {
  type        = string
  description = "VPC CIDR"
  default     = "10.3.0.0/16"
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

variable "enable_internal_subnets" {
  type        = bool
  description = "Whether to enable internal subnets"
  default     = true
}
