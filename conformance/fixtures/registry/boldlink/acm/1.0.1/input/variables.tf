variable "domain_name" {
  description = "(Required) Domain name for which the certificate should be issued"
  type        = string
}

variable "subject_alternative_names" {
  description = "(Optional) Set of domains that should be SANs in the issued certificate. To remove all elements of a previously configured list, set this value equal to an empty list ([]) "
  type        = list(string)
  default     = []
}

variable "validation_method" {
  description = "(Required) Which method to use for validation. DNS or EMAIL are valid, NONE can be used for certificates that were imported into ACM and then into Terraform."
  type        = string
}

variable "certificate_transparency_logging_preference" {
  description = "(Optional) Whether certificate details should be added to a certificate transparency log. Valid values are ENABLED or DISABLED"
  type        = string
  default     = "ENABLED"
}

variable "validation_option" {
  description = "(Optional) Configuration block used to specify information about the initial validation of each domain name."
  type        = any
  default     = {}
}

variable "tags" {
  description = "(Optional) Map of tags to assign to the resource. If configured with a provider default_tags configuration block present, tags with matching keys will overwrite those defined at the provider-level."
  type        = map(string)
  default     = {}
}

variable "allow_overwrite" {
  description = "(Optional) Allow creation of this record in Terraform to overwrite an existing record, if any. This does not affect the ability to update the record in Terraform and does not prevent other resources within Terraform or manual Route 53 changes outside Terraform from overwriting this record. false by default. This configuration is not recommended for most environments."
  type        = bool
  default     = true
}

variable "ttl" {
  description = "The TTL of the record."
  type        = number
  default     = 60
}

variable "wait_for_issue" {
  description = "Whether to wait for the certificate to be issued by ACM"
  type        = bool
  default     = false
}
