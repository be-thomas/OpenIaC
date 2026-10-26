variable "security_group_name" {
  type        = string
  description = "The name to assign to the interfaces security group"
  default     = null
}

variable "vpc_endpoints" {
  type        = any
  description = "Configuration lists for vpc endpoints"
  default     = []
}

variable "vpc_id" {
  type        = string
  description = "The ID of the VPC in which the endpoint will be used."
}

variable "tags" {
  type        = map(string)
  description = "A map of tags to assign to the resources"
  default     = {}
}

variable "create_endpoint_sg" {
  type        = bool
  description = "Specify whether to create Security Group for Interface endpoints"
  default     = false
}
