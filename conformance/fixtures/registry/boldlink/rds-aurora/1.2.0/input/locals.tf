locals {
  vpc_security_group_ids = length(var.vpc_security_group_ids) == 0 && var.create_security_group == false ? null : var.vpc_security_group_ids
}
