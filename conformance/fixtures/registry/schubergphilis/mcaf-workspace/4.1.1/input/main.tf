locals {
  oidc_project_filter   = try(var.oidc_settings.project_scope, false) ? var.oidc_settings.project_name : "*"
  oidc_workspace_filter = !try(var.oidc_settings.project_scope, false) ? var.name : "*"
}

################################################################################
# Workspace
################################################################################

module "tfe-workspace" {
  source  = "schubergphilis/mcaf-workspace/tfe"
  version = "~> 3.0.0"

  name                                         = var.name
  agent_pool_id                                = var.execution_mode == "agent" ? var.agent_pool_id : null
  allow_destroy_plan                           = var.allow_destroy_plan
  assessments_enabled                          = var.assessments_enabled
  auto_apply                                   = var.auto_apply
  auto_apply_run_trigger                       = var.auto_apply_run_trigger
  auto_destroy_activity_duration               = var.auto_destroy_activity_duration
  auto_destroy_at                              = var.auto_destroy_at
  branch                                       = var.branch
  clear_text_env_variables                     = var.region != null ? merge(var.clear_text_env_variables, { AWS_DEFAULT_REGION = var.region }) : var.clear_text_env_variables
  clear_text_hcl_variables                     = var.clear_text_hcl_variables
  clear_text_terraform_variables               = var.clear_text_terraform_variables
  description                                  = var.description
  execution_mode                               = var.execution_mode
  file_triggers_enabled                        = var.file_triggers_enabled
  force_delete                                 = var.force_delete
  github_app_installation_id                   = var.repository_identifier != null ? var.github_app_installation_id : null
  global_remote_state                          = var.global_remote_state
  notification_configuration                   = var.notification_configuration
  oauth_token_id                               = var.repository_identifier != null ? var.oauth_token_id : null
  project_id                                   = var.project_id
  queue_all_runs                               = var.queue_all_runs
  remote_state_consumer_ids                    = var.remote_state_consumer_ids
  repository_identifier                        = var.repository_identifier
  sensitive_env_variables                      = var.sensitive_env_variables
  sensitive_hcl_variables                      = var.sensitive_hcl_variables
  sensitive_terraform_variables                = var.sensitive_terraform_variables
  speculative_enabled                          = var.speculative_enabled
  ssh_key_id                                   = var.ssh_key_id
  terraform_organization                       = var.terraform_organization
  terraform_version                            = var.terraform_version
  trigger_patterns                             = var.trigger_patterns
  trigger_patterns_working_directory_recursive = var.trigger_patterns_working_directory_recursive
  variable_set_ids                             = var.variable_set_ids
  variable_set_names                           = var.variable_set_names
  working_directory                            = var.working_directory
  workspace_tags                               = var.workspace_tags
}

################################################################################
# RBAC
################################################################################

data "tfe_team" "default" {
  for_each = toset(keys(var.team_access))

  name         = each.value
  organization = var.terraform_organization
}

resource "tfe_team_access" "default" {
  for_each = var.team_access

  access       = each.value.access
  team_id      = data.tfe_team.default[each.key].id
  workspace_id = module.tfe-workspace.workspace_id

  dynamic "permissions" {
    for_each = each.value.permissions != null ? { create = true } : {}

    content {
      run_tasks         = each.value.permissions["run_tasks"]
      runs              = each.value.permissions["runs"]
      sentinel_mocks    = each.value.permissions["sentinel_mocks"]
      state_versions    = each.value.permissions["state_versions"]
      variables         = each.value.permissions["variables"]
      workspace_locking = each.value.permissions["workspace_locking"]
    }
  }
}

################################################################################
# Authentication TFE Workspace <> AWS IAM Role/User
################################################################################

module "auth" {
  count = var.enable_authentication ? 1 : 0

  source = "./modules/auth"

  agent_role_arns          = var.agent_role_arns
  auth_method              = var.auth_method
  path                     = var.path
  permissions_boundary_arn = var.permissions_boundary_arn
  policy                   = var.policy
  policy_arns              = var.policy_arns
  postfix                  = var.postfix
  role_name                = var.role_name
  tags                     = var.tags
  terraform_organization   = var.terraform_organization
  username                 = var.username
  workspace_id             = module.tfe-workspace.workspace_id

  oidc_settings = var.oidc_settings != null ? {
    audience              = var.oidc_settings.audience
    oidc_project_filter   = local.oidc_project_filter
    oidc_workspace_filter = local.oidc_workspace_filter
    provider_arn          = var.oidc_settings.provider_arn
    site_address          = var.oidc_settings.site_address
  } : null
}
