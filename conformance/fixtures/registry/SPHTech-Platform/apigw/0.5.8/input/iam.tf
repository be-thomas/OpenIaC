module "apigw_cwl_role" {
  #checkov:skip=CKV_TF_1:Ensure Terraform module sources use a commit hash
  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = "~> 5.9.2"

  trusted_role_services = [
    "apigateway.amazonaws.com"
  ]

  create_role = var.enable_global_apigw_logging

  role_name_prefix  = "apigw-cwl-global"
  role_requires_mfa = false

  custom_role_policy_arns = [
    "arn:aws:iam::aws:policy/service-role/AmazonAPIGatewayPushToCloudWatchLogs",
  ]
}

resource "aws_api_gateway_account" "api_gateway_account" {
  count = var.enable_global_apigw_logging ? 1 : 0

  cloudwatch_role_arn = module.apigw_cwl_role.iam_role_arn
}
