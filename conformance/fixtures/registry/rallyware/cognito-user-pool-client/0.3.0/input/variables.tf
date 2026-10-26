variable "user_pool_id" {
  type        = string
  description = "User pool the client belongs to."
}

variable "generate_secret" {
  type        = bool
  default     = true
  description = "Should an application secret be generated."
}

variable "callback_urls" {
  type        = list(string)
  default     = []
  description = "List of allowed callback URLs for the identity providers."
}

variable "allowed_oauth_flows" {
  type        = list(string)
  default     = ["code", "implicit"]
  description = "List of allowed OAuth flows (code, implicit, client_credentials)."
}

variable "allowed_oauth_scopes" {
  type        = list(string)
  default     = ["email", "openid", "profile", "aws.cognito.signin.user.admin"]
  description = "List of allowed OAuth scopes (phone, email, openid, profile, and aws.cognito.signin.user.admin)."
}

variable "supported_identity_providers" {
  type        = list(string)
  default     = []
  description = "List of provider names for the identity providers that are supported on this client."
}

variable "allowed_oauth_flows_user_pool_client" {
  type        = bool
  default     = true
  description = "Whether the client is allowed to follow the OAuth protocol when interacting with Cognito user pools."
}

variable "explicit_auth_flows" {
  type        = list(string)
  default     = ["ALLOW_REFRESH_TOKEN_AUTH", "ALLOW_USER_SRP_AUTH", "ALLOW_CUSTOM_AUTH"]
  description = "List of authentication flows (ADMIN_NO_SRP_AUTH, CUSTOM_AUTH_FLOW_ONLY, USER_PASSWORD_AUTH, ALLOW_ADMIN_USER_PASSWORD_AUTH, ALLOW_CUSTOM_AUTH, ALLOW_USER_PASSWORD_AUTH, ALLOW_USER_SRP_AUTH, ALLOW_REFRESH_TOKEN_AUTH)."
}

variable "logout_urls" {
  type        = list(string)
  default     = []
  description = "List of allowed logout URLs for the identity providers."
}

variable "prevent_user_existence_errors" {
  type        = string
  default     = "ENABLED"
  description = "Choose which errors and responses are returned by Cognito APIs during authentication, account confirmation, and password recovery when the user does not exist in the user pool. Possible values: ENABLED, LEGACY"
}

variable "token_ttl" {
  type = object({
    id      = optional(string, "1h")
    refresh = optional(string, "30d")
    access  = optional(string, "1h")
  })
  default     = {}
  description = "The token TTL. Allowed values are number of time units (s, m, h, d)."
}
