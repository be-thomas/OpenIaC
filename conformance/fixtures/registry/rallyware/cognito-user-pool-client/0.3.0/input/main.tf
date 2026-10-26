locals {
  enabled = module.this.enabled

  time_units = {
    s = "seconds"
    m = "minutes"
    h = "hours"
    d = "days"
  }
}

resource "aws_cognito_user_pool_client" "default" {
  count = local.enabled ? 1 : 0

  name                                 = module.this.id
  user_pool_id                         = var.user_pool_id
  generate_secret                      = var.generate_secret
  callback_urls                        = var.callback_urls
  allowed_oauth_flows_user_pool_client = var.allowed_oauth_flows_user_pool_client
  allowed_oauth_flows                  = var.allowed_oauth_flows
  allowed_oauth_scopes                 = var.allowed_oauth_scopes
  explicit_auth_flows                  = var.explicit_auth_flows
  logout_urls                          = var.logout_urls
  supported_identity_providers         = var.supported_identity_providers
  prevent_user_existence_errors        = var.prevent_user_existence_errors
  access_token_validity                = regex("(\\d+)([smhd])", var.token_ttl["access"])[0]
  id_token_validity                    = regex("(\\d+)([smhd])", var.token_ttl["id"])[0]
  refresh_token_validity               = regex("(\\d+)([smhd])", var.token_ttl["refresh"])[0]

  token_validity_units {
    access_token  = local.time_units[regex("(\\d+)([smhd])", var.token_ttl["access"])[1]]
    id_token      = local.time_units[regex("(\\d+)([smhd])", var.token_ttl["id"])[1]]
    refresh_token = local.time_units[regex("(\\d+)([smhd])", var.token_ttl["refresh"])[1]]
  }
}
