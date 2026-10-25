locals {
  set_nat_single = (var.nat == "single" && var.enable_private_subnets) ? true : null
  set_nat_multi  = (var.nat == "multi" && var.enable_private_subnets) ? true : null
}
