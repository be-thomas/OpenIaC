variable "name" {
  description = "The internal subnet name"
  type        = string
}

variable "cidrs" {
  description = "Internal subnet ipv4 CIDR list"
  type        = list(string)
}

variable "vpc_id" {
  description = "The VPC ID to associate with"
  type        = string
}

variable "ipv6_cidr_block" {
  description = "VPC ipv6 CIDR"
  type        = string
  default     = null
}

variable "assign_ipv6_address_on_creation" {
  description = "(Optional) Specify true to indicate that network interfaces created in the specified subnet should be assigned an IPv6 address. Default is false"
  type        = bool
  default     = false
}

variable "assign_generated_ipv6_cidr_block" {
  type        = bool
  description = "This variable here is used to enable routes and other IPv6 related resources, defaults to `false`"
  default     = false
}

variable "internal_subnet_ipv6_prefixes" {
  type        = list(string)
  description = "values to use for the ipv6_cidr_block when creating internal subnets. This is a list of integers between 0 and 255. The length of this list must be equal to the number of public subnets."
  default     = []
}

variable "tags" {
  description = "The map of tags to add to module resources"
  type        = map(string)
  default     = {}
}

variable "availability_zone" {
  description = "The AZs to use to assign each subnet to"
  type        = list(string)
  default     = []
}

variable "vpc_name" {
  description = "This value is used by the private subnets module to make it easier to search for the nat gateways through a specific tag"
  type        = string
  default     = null
}
