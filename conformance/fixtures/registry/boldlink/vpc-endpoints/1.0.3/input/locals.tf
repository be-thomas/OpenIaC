locals {
  vpc_cidr_block = data.aws_vpc.selected.cidr_block
  include_security_group = [for ep in var.vpc_endpoints :
    lookup(ep, "vpc_endpoint_type", "Interface") == "Interface" && var.create_endpoint_sg ? true : false
  ]
}
