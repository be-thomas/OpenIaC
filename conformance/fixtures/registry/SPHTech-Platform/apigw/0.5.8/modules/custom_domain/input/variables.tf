################################################################################
# variables for Custom domain mapping sub module
################################################################################

variable "domain_name" {
  description = "Custom domain name"
  type        = string
}

variable "security_policy" {
  description = "TLS Security Policy for the domain"
  type        = string
  default     = "TLS_1_2"
}

variable "endpoint_access_mode" {
  type        = string
  description = "Endpoint access mode for the custom domain. Required for domain names that use security policies that start with `SecurityPolicy_*`. Valid values: `BASIC`, `STRICT`."
  default     = null
  validation {
    condition     = var.endpoint_access_mode == null || contains(["BASIC", "STRICT"], coalesce(var.endpoint_access_mode, "BASIC"))
    error_message = "Valid values for endpoint_access_mode are BASIC, STRICT, or null."
  }
}

variable "create_acm_cert" {
  description = "Create ACM cert. create_acm_cert and cert_arn Mutually exclusive. "
  type        = bool
  default     = false
}

variable "cert_arn" {
  description = "Cert ARN. Create ACM cert. create_acm_cert and cert_arn Mutually exclusive."
  type        = string
  default     = ""
}

variable "key_algorithm" {
  description = "Key algorithm for the cert"
  type        = string
  default     = "EC_prime256v1"
}

variable "path_mappings" {
  description = "List of stages the usage plan can be used "
  type = map(
    object({
      api_id     = string
      stage_name = string
      base_path  = string
    })
  )
}

variable "endpoint_type" {
  description = "Endpoint type."
  type        = string
  default     = "REGIONAL"

  validation {
    condition     = var.endpoint_type == "EDGE" || var.endpoint_type == "REGIONAL"
    error_message = "endpoint_type must be 'EDGE' or 'REGIONAL'."
  }
}
