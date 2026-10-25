module "aurora_vpc" {
  source                  = "boldlink/vpc/aws"
  version                 = "3.0.4"
  name                    = local.name
  cidr_block              = local.cidr_block
  enable_internal_subnets = true

  internal_subnets = {
    databases = {
      cidrs = local.database_subnets
    }
  }
  tags = local.tags
}

module "kms_key" {
  source                  = "boldlink/kms/aws"
  description             = "A test kms key for rds"
  alias_name              = "alias/${local.name}-alias"
  create_kms_alias        = true
  enable_key_rotation     = true
  deletion_window_in_days = 7
}
