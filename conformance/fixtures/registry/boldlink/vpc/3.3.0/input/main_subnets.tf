module "public_subnets" {
  source                           = "./modules/public/"
  for_each                         = { for k, v in var.public_subnets : k => v if var.enable_public_subnets == true }
  name                             = try(each.value.name, each.key)
  vpc_id                           = aws_vpc.main.id
  vpc_name                         = var.name
  enable_private_subnets           = var.enable_private_subnets
  propagating_vgws                 = try(each.value.propagating_vgws, null)
  cidrs                            = try(each.value.cidrs, null)
  ipv6_cidr_block                  = try(aws_vpc.main.ipv6_cidr_block, null)
  assign_generated_ipv6_cidr_block = try(var.assign_generated_ipv6_cidr_block, null)
  public_subnet_ipv6_prefixes      = try(each.value.public_subnet_ipv6_prefixes, [])
  nat                              = try(each.value.nat, null)
  availability_zone                = try(each.value.override_azs, []) == [] ? data.aws_availability_zones.available.names : each.value.override_azs
  map_public_ip_on_launch          = try(each.value.map_public_ip_on_launch, null)
  gateway_id                       = aws_internet_gateway.main[0].id
  tags = merge(
    var.tags,
    try(each.value.tags, null)
  )

}

module "private_subnets" {
  source                           = "./modules/private/"
  for_each                         = { for k, v in var.private_subnets : k => v if var.enable_private_subnets == true }
  name                             = try(each.value.name, each.key)
  vpc_name                         = var.name
  vpc_id                           = aws_vpc.main.id
  cidrs                            = try(each.value.cidrs)
  ipv6_cidr_block                  = try(aws_vpc.main.ipv6_cidr_block, null)
  assign_generated_ipv6_cidr_block = try(var.assign_generated_ipv6_cidr_block, null)
  private_subnet_ipv6_prefixes     = try(each.value.private_subnet_ipv6_prefixes, [])
  availability_zone                = try(each.value.override_azs, []) == [] ? data.aws_availability_zones.available.names : each.value.override_azs
  nat_gateway_ids                  = flatten(local.nat_gateways)
  tags = merge(
    var.tags,
    try(each.value.tags, null)
  )
  depends_on = [module.public_subnets]
}

module "internal_subnets" {
  source                           = "./modules/internal/"
  for_each                         = { for k, v in var.internal_subnets : k => v if var.enable_internal_subnets == true }
  name                             = try(each.value.name, each.key)
  vpc_name                         = var.name
  vpc_id                           = aws_vpc.main.id
  cidrs                            = try(each.value.cidrs)
  ipv6_cidr_block                  = try(aws_vpc.main.ipv6_cidr_block, null)
  assign_generated_ipv6_cidr_block = try(var.assign_generated_ipv6_cidr_block, null)
  internal_subnet_ipv6_prefixes    = try(each.value.internal_subnet_ipv6_prefixes, [])
  availability_zone                = try(each.value.override_azs, []) == [] ? data.aws_availability_zones.available.names : each.value.override_azs
  tags = merge(
    var.tags,
    try(each.value.tags, null)
  )
}
