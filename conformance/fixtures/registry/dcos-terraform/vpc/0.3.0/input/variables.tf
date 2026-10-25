variable "availability_zones" {
  description = "List of availability_zones to be used as the same format that are required by the platform/cloud providers. i.e `['RegionZone']`"
  type        = list(string)
  default     = []
}

variable "cluster_name" {
  description = "Name of the DC/OS cluster"
}

variable "subnet_range" {
  description = "Private IP space to be used in CIDR format"
  default     = "172.31.0.0/16"
}

variable "tags" {
  description = "Add custom tags to all resources"
  type        = map(string)
  default     = {}
}

