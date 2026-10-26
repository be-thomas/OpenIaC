locals {
  create_nat = var.nat_gateway_ids != [] && length(var.cidrs) > 0 ? true : false
}
