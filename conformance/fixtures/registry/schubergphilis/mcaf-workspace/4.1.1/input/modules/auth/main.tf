locals {
  enable_oidc = var.auth_method == "iam_role_oidc" && var.oidc_settings != null
}

################################################################################
# Auth - IAM User
################################################################################

module "workspace_iam_user" {
  count = var.auth_method == "iam_user" ? 1 : 0

  source  = "schubergphilis/mcaf-user/aws"
  version = "~> 0.4.0"

  name                 = var.username
  path                 = var.path
  policy               = var.policy
  policy_arns          = var.policy_arns
  postfix              = var.postfix
  permissions_boundary = var.permissions_boundary_arn
  tags                 = var.tags
}

resource "tfe_variable" "aws_access_key_id" {
  count = var.auth_method == "iam_user" ? 1 : 0

  key             = "AWS_ACCESS_KEY_ID"
  value           = module.workspace_iam_user[0].access_key_id
  category        = "env"
  variable_set_id = var.variable_set_id
  workspace_id    = var.workspace_id
}

resource "tfe_variable" "aws_secret_access_key" {
  count = var.auth_method == "iam_user" ? 1 : 0

  key             = "AWS_SECRET_ACCESS_KEY"
  value           = module.workspace_iam_user[0].secret_access_key
  category        = "env"
  sensitive       = true
  variable_set_id = var.variable_set_id
  workspace_id    = var.workspace_id
}

################################################################################
# Auth - IAM Role - External ID & Agent
################################################################################

resource "random_uuid" "external_id" {
  count = var.auth_method == "iam_role" ? 1 : 0
}

module "workspace_iam_role" {
  count = var.auth_method == "iam_role" ? 1 : 0

  source  = "schubergphilis/mcaf-role/aws"
  version = "~> 0.4.0"

  name                 = var.role_name
  path                 = var.path
  permissions_boundary = var.permissions_boundary_arn
  policy_arns          = var.policy_arns
  postfix              = var.postfix
  role_policy          = var.policy
  tags                 = var.tags

  assume_policy = templatefile("${path.module}/templates/assume_role_policy.tftpl", {
    external_id    = random_uuid.external_id[0].result,
    role_arns_json = jsonencode(var.agent_role_arns)
  })
}

resource "tfe_variable" "aws_assume_role" {
  count = var.auth_method == "iam_role" ? 1 : 0

  key             = "aws_assume_role"
  value           = module.workspace_iam_role[0].arn
  category        = "terraform"
  variable_set_id = var.variable_set_id
  workspace_id    = var.workspace_id
}

resource "tfe_variable" "aws_assume_role_external_id" {
  count = var.auth_method == "iam_role" ? 1 : 0

  key             = "aws_assume_role_external_id"
  value           = random_uuid.external_id[0].result
  category        = "terraform"
  sensitive       = true
  variable_set_id = var.variable_set_id
  workspace_id    = var.workspace_id
}

################################################################################
# Auth - IAM Role - OIDC
################################################################################
data "tfe_organization" "default" {
  count = local.enable_oidc ? 1 : 0
}

module "workspace_iam_role_oidc" {
  count = local.enable_oidc ? 1 : 0

  source  = "schubergphilis/mcaf-role/aws"
  version = "~> 0.5.3"

  name                 = var.role_name
  path                 = var.path
  permissions_boundary = var.permissions_boundary_arn
  policy_arns          = var.policy_arns
  postfix              = var.postfix
  role_policy          = var.policy
  tags                 = var.tags

  assume_policy = templatefile("${path.module}/templates/assume_role_policy_oidc.tftpl", {
    audience         = var.oidc_settings.audience,
    org_name         = var.terraform_organization != null ? var.terraform_organization : data.tfe_organization.default[0].name,
    project_filter   = var.oidc_settings.oidc_project_filter,
    provider_arn     = var.oidc_settings.provider_arn,
    site_address     = var.oidc_settings.site_address,
    workspace_filter = var.oidc_settings.oidc_workspace_filter,
  })
}

resource "tfe_variable" "tfc_aws_provider_auth" {
  count = local.enable_oidc ? 1 : 0

  key             = "TFC_AWS_PROVIDER_AUTH"
  value           = "true"
  category        = "env"
  variable_set_id = var.variable_set_id
  workspace_id    = var.workspace_id
}

resource "tfe_variable" "tfc_aws_run_role_arn" {
  count = local.enable_oidc ? 1 : 0

  key             = "TFC_AWS_RUN_ROLE_ARN"
  value           = module.workspace_iam_role_oidc[0].arn
  category        = "env"
  variable_set_id = var.variable_set_id
  workspace_id    = var.workspace_id
}

resource "tfe_variable" "tfc_aws_workload_identity_audience" {
  count = local.enable_oidc ? 1 : 0

  key             = "TFC_AWS_WORKLOAD_IDENTITY_AUDIENCE"
  value           = var.oidc_settings.audience
  category        = "env"
  variable_set_id = var.variable_set_id
  workspace_id    = var.workspace_id
}
