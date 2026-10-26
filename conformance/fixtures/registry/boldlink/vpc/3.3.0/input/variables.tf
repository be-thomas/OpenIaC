#####################
## VPC
#####################

variable "cidr_block" {
  type        = string
  description = "(Optional) The IPv4 CIDR block for the VPC. CIDR can be explicitly set or it can be derived from IPAM using `ipv4_netmask_length`."
  default     = null
}

variable "instance_tenancy" {
  type        = string
  description = "(Optional) A tenancy option for instances launched into the VPC. Default is `default`, which makes your instances shared on the host. Using either of the other options (`dedicated` or `host`) costs at least `$2/hr`."
  default     = "default"
}

variable "ipv4_ipam_pool_id" {
  type        = string
  description = "(Optional) The ID of an IPv4 IPAM pool you want to use for allocating this VPC's CIDR. IPAM is a VPC feature that you can use to automate your IP address management workflows including assigning, tracking, troubleshooting, and auditing IP addresses across AWS Regions and accounts. Using IPAM you can monitor IP address usage throughout your AWS Organization."
  default     = null
}

variable "ipv4_netmask_length" {
  type        = number
  description = "(Optional) The netmask length of the IPv4 CIDR you want to allocate to this VPC. Requires specifying a `ipv4_ipam_pool_id`."
  default     = null
}

variable "ipv6_ipam_pool_id" {
  type        = string
  description = "(Optional) IPAM Pool ID for a IPv6 pool. Conflicts with `assign_generated_ipv6_cidr_block`."
  default     = null
}

variable "ipv6_netmask_length" {
  type        = number
  description = "(Optional) Netmask length to request from IPAM Pool. Conflicts with ipv6_cidr_block. This can be omitted if IPAM pool as a allocation_default_netmask_length set. Valid values: `56`."
  default     = null
}

variable "ipv6_cidr_block_network_border_group" {
  type        = string
  description = "(Optional) By default when an IPv6 CIDR is assigned to a VPC a default ipv6_cidr_block_network_border_group will be set to the region of the VPC. This can be changed to restrict advertisement of public addresses to specific Network Border Groups such as LocalZones."
  default     = null
}

variable "enable_dns_support" {
  type        = bool
  description = "(Optional) A boolean flag to enable/disable DNS support in the VPC. Defaults `true`."
  default     = true
}

variable "enable_dns_hostnames" {
  type        = bool
  description = "(Optional) A boolean flag to enable/disable DNS hostnames in the VPC. Defaults `false`."
  default     = false
}

variable "assign_generated_ipv6_cidr_block" {
  type        = bool
  description = "(Optional) Requests an Amazon-provided IPv6 CIDR block with a /56 prefix length for the VPC. You cannot specify the range of IP addresses, or the size of the CIDR block. Default is false. Conflicts with `ipv6_ipam_pool_id`"
  default     = false
}

#####################
## VPC flow logs
#####################

variable "traffic_type" {
  type        = string
  description = "(Required) The type of traffic to capture. Valid values: `ACCEPT`,`REJECT`, `ALL`."
  default     = "ALL"
}

variable "flow_log_eni_id" {
  type        = string
  description = "(Optional) Elastic Network Interface ID to attach to"
  default     = null
}

variable "log_group_name" {
  description = "Use this variable to override the default log group name `/aws/vpc/name`"
  type        = string
  default     = null
}

variable "retention_in_days" {
  description = "(Optional) Specifies the number of days you want to retain log events in the specified log group. Possible values are: 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653, and 0. If you select 0, the events in the log group are always retained and never expire"
  type        = number
  default     = 0
}

variable "logs_kms_key_id" {
  description = "(Optional) The ARN of the KMS Key to use when encrypting log data."
  type        = string
  default     = null
}

variable "log_destination_type" {
  type        = string
  description = "(Optional) The type of the logging destination. Valid values: `cloud-watch-logs`, `s3`. Default: `cloud-watch-logs`."
  default     = "cloud-watch-logs"
}

variable "flow_log_subnet_id" {
  type        = string
  description = "(Optional) Subnet ID to attach to"
  default     = null
}

variable "log_format" {
  type        = string
  description = "(Optional) The fields to include in the flow log record, in the order in which they should appear."
  default     = null
}

variable "max_aggregation_interval" {
  type        = number
  description = "(Optional) The maximum interval of time during which a flow of packets is captured and aggregated into a flow log record. Valid Values: `60 seconds (1 minute) or 600 seconds (10 minutes)`. Default: `600`."
  default     = 600
}

variable "destination_options" {
  type        = map(any)
  description = "(Optional) Describes the destination options for a flow log."
  default     = {}
}

#####################
## DHCP options Set
#####################

variable "dhcp_domain_name" {
  type        = string
  description = "(Optional) the suffix domain name to use by default when resolving non Fully Qualified Domain Names. In other words, this is what ends up being the `search` value in the `/etc/resolv.conf` file."
  default     = null
}

variable "domain_name_servers" {
  type        = list(string)
  description = "(Optional) List of name servers to configure in /etc/resolv.conf. If you want to use the default AWS nameservers you should set this to `AmazonProvidedDNS`."
  default     = ["AmazonProvidedDNS"]
}

variable "ntp_servers" {
  type        = list(string)
  description = "(Optional) List of NTP servers to configure."
  default     = []
}

variable "netbios_name_servers" {
  type        = list(string)
  description = "(Optional) List of NETBIOS name servers."
  default     = []
}

variable "netbios_node_type" {
  type        = number
  description = "(Optional) The NetBIOS node type (1, 2, 4, or 8). AWS recommends to specify 2 since broadcast and multicast are not supported in their network."
  default     = 2
}

##########################################
## Subnets & Other resources
##########################################

variable "enable_public_subnets" {
  description = "Activate public subnets module"
  type        = bool
  default     = false
}
variable "public_subnets" {
  type        = any
  description = "Public subnets settings map"
  default     = {}
}

variable "enable_private_subnets" {
  description = "Activate private subnets module"
  type        = bool
  default     = false
}

variable "private_subnets" {
  type        = any
  description = "Private subnets settings map"
  default     = {}
}

variable "enable_internal_subnets" {
  description = "Activate internal subnets module"
  type        = bool
  default     = false
}

variable "internal_subnets" {
  type        = any
  description = "Internal subnets settings map"
  default     = {}
}

variable "name" {
  type        = string
  description = "Input the name of stack"
}

variable "tags" {
  type        = map(string)
  description = "(Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level."
  default     = {}
}

variable "logs_bucket_arn" {
  description = "The ARN of the s3 bucket to store the logs. Provide this when the value of `log_destination_type` is `s3`"
  type        = string
  default     = null
}
