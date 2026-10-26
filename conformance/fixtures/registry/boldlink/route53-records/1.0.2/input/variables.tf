variable "zone_id" {
  type        = string
  description = "(Required) The ID of the hosted zone to contain this record."
  default     = null
}

variable "name" {
  type        = string
  description = "(Required) The name of the record."
  default     = null
}

variable "type" {
  type        = string
  description = "(Required) The record type. Valid values are `A`, `AAAA`, `CAA`, `CNAME`, `DS`, `MX`, `NAPTR`, `NS`, `PTR`, `SOA`, `SPF`, `SRV` and `TXT`"
  default     = null
}

variable "ttl" {
  type        = number
  description = "(Required for non-alias records) The TTL of the record."
  default     = null
}

variable "records" {
  type        = list(string)
  description = "(Required for non-alias records) A string list of records. To specify a single record value longer than 255 characters such as a TXT record for DKIM, add \"\" inside the Terraform configuration string (e.g., \"first255characters`\"\"`morecharacters\")."
  default     = []
}

variable "allow_overwrite" {
  type        = bool
  description = "Allow creation of this record in Terraform to overwrite an existing record, if any. This does not affect the ability to update the record in Terraform and does not prevent other resources within Terraform or manual Route 53 changes outside Terraform from overwriting this record. `false` by default. This configuration is not recommended for most environments."
  default     = false
}

variable "set_identifier" {
  type        = string
  description = "(Optional) Unique identifier to differentiate records with routing policies from one another. Required if using `cidr_routing_policy`, `failover_routing_policy`, `geolocation_routing_policy`, `latency_routing_policy`, `multivalue_answer_routing_policy`, or `weighted_routing_policy`."
  default     = null
}

variable "health_check_id" {
  type        = string
  description = "(Optional) The health check the record should be associated with."
  default     = null
}

variable "alias" {
  type        = any
  description = "(Optional) An alias block. Conflicts with ttl & records."
  default     = {}
}

variable "cidr_routing_policy" {
  type        = map(string)
  description = "(Optional) A block indicating a routing policy based on the IP network ranges of requestors. Conflicts with any other routing policy."
  default     = {}
}

variable "failover_routing_policy" {
  type        = any
  description = "(Optional) A block indicating the routing behavior when associated health check fails. Conflicts with any other routing policy."
  default     = {}
}

variable "geolocation_routing_policy" {
  type        = any
  description = "(Optional) A block indicating a routing policy based on the geolocation of the requestor. Conflicts with any other routing policy."
  default     = {}
}

variable "latency_routing_policy" {
  type        = any
  description = "(Optional) A block indicating a routing policy based on the latency between the requestor and an AWS region. Conflicts with any other routing policy."
  default     = {}
}

variable "weighted_routing_policy" {
  type        = any
  description = "(Optional) A block indicating a weighted routing policy. Conflicts with any other routing policy."
  default     = {}
}

variable "multivalue_answer_routing_policy" {
  type        = bool
  description = "(Optional) Set to `true` to indicate a multivalue answer routing policy. Conflicts with any other routing policy."
  default     = null
}
