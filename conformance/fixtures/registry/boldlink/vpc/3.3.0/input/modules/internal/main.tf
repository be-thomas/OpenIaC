##############################
### Internal routes
##############################

resource "aws_route_table" "internal" {
  count  = length(var.cidrs) > 0 ? length(var.cidrs) : 0
  vpc_id = var.vpc_id
  tags = merge(
    {
      "Name" = "${var.vpc_name}.${var.name}.int.${count.index}"
    },
    var.tags,
  )
}

##########################################################################################
### Internal Subnets: Private subnets without direct access to the internet
##########################################################################################

resource "aws_subnet" "internal" {
  count                           = length(var.cidrs) > 0 ? length(var.cidrs) : 0
  vpc_id                          = var.vpc_id
  cidr_block                      = element(var.cidrs, count.index)
  availability_zone               = element(var.availability_zone, count.index)
  ipv6_cidr_block                 = var.assign_generated_ipv6_cidr_block && length(var.internal_subnet_ipv6_prefixes) > 0 ? cidrsubnet(var.ipv6_cidr_block, 8, var.internal_subnet_ipv6_prefixes[count.index]) : null
  assign_ipv6_address_on_creation = var.assign_ipv6_address_on_creation
  tags = merge(
    {
      "Name" = "${var.vpc_name}.${var.name}.int.${count.index}"
    },
    var.tags,
  )
}

##############################
### internal Route Association
##############################

resource "aws_route_table_association" "internal" {
  count          = length(var.cidrs) > 0 ? length(var.cidrs) : 0
  subnet_id      = element(aws_subnet.internal.*.id, count.index)
  route_table_id = element(aws_route_table.internal.*.id, count.index)
}
