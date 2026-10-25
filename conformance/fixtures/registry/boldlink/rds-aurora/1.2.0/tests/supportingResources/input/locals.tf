locals {
  cidr_block       = "10.1.0.0/16"
  database_subnets = [cidrsubnet(local.cidr_block, 8, 13), cidrsubnet(local.cidr_block, 8, 14), cidrsubnet(local.cidr_block, 8, 15)]
  name             = "terraform-aws-rds-aurora"

  tags = {
    Environment        = "examples"
    "user::CostCenter" = "terraform-registry"
    Department         = "operations"
    Project            = "aws-vpc"
    Owner              = "Boldlink"
    LayerName          = "cExample"
    LayerId            = "cExample"
  }
}
