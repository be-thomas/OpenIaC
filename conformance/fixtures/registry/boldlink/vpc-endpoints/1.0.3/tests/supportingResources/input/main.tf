module "endpoints_vpc" {
  #checkov:skip=CKV_TF_1: "Ensure Terraform module sources use a commit hash"
  source                  = "boldlink/vpc/aws"
  version                 = "3.0.4"
  name                    = var.name
  cidr_block              = var.cidr_block
  enable_dns_support      = var.enable_dns_support
  enable_dns_hostnames    = var.enable_dns_hostnames
  enable_internal_subnets = var.enable_internal_subnets
  tags                    = var.tags

  internal_subnets = {
    apps = {
      cidrs = local.internal_subnets
    }
  }
}
