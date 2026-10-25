##############################
### Publiс routes
##############################

resource "aws_route_table" "public" {
  count            = length(var.cidrs) > 0 ? 1 : 0
  vpc_id           = var.vpc_id
  propagating_vgws = var.propagating_vgws
  tags = merge(
    {
      "Name" = "${var.vpc_name}.${var.name}.pub.${count.index}"
    },
    var.tags,
  )
}

resource "aws_route" "public_internet_gateway" {
  count                  = length(var.cidrs) > 0 ? 1 : 0
  route_table_id         = aws_route_table.public[0].id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = var.gateway_id
}

resource "aws_route" "public_internet_gateway_ipv6" {
  count                       = var.assign_generated_ipv6_cidr_block && length(var.cidrs) > 0 ? 1 : 0
  route_table_id              = aws_route_table.public[0].id
  destination_ipv6_cidr_block = "::/0"
  gateway_id                  = var.gateway_id
}

##############################
### Public Subnets
##############################
resource "aws_subnet" "public" {
  count                           = length(var.cidrs) > 0 ? length(var.cidrs) : 0
  vpc_id                          = var.vpc_id
  cidr_block                      = element(var.cidrs, count.index)
  availability_zone               = element(var.availability_zone, count.index)
  map_public_ip_on_launch         = var.map_public_ip_on_launch
  ipv6_cidr_block                 = var.assign_generated_ipv6_cidr_block && length(var.public_subnet_ipv6_prefixes) > 0 ? cidrsubnet(var.ipv6_cidr_block, 8, var.public_subnet_ipv6_prefixes[count.index]) : null
  assign_ipv6_address_on_creation = var.assign_ipv6_address_on_creation
  tags = merge(
    {
      "Name" = "${var.vpc_name}.${var.name}.pub.${count.index}"
    },
    var.tags,
  )
}

##############################
### Public Route Association
##############################
resource "aws_route_table_association" "public" {
  count          = length(var.cidrs)
  subnet_id      = element(aws_subnet.public.*.id, count.index)
  route_table_id = aws_route_table.public[0].id
}

##############################
### SingleAz NatG resources
##############################
resource "aws_eip" "single" {
  count  = local.set_nat_single == true ? 1 : 0
  domain = "vpc"
  tags = merge(
    {
      "Name" = "${var.vpc_name}.${var.name}.eip.${count.index}"
    },
    var.tags,
  )
  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_nat_gateway" "single" {
  count         = local.set_nat_single == true ? 1 : 0
  allocation_id = aws_eip.single[0].id
  subnet_id     = element(aws_subnet.public.*.id, count.index)
  tags = merge(
    {
      "Name" = "${var.vpc_name}.${var.name}.nat-gw.${count.index}"
    },
    var.tags,
  )
  lifecycle {
    create_before_destroy = true
  }
}

##############################
### MultiAz NatG resources
##############################
resource "aws_eip" "multi" {
  count  = local.set_nat_multi == true ? length(var.cidrs) : 0
  domain = "vpc"
  tags = merge(
    {
      "Name" = "${var.vpc_name}.${var.name}.eip.${count.index}"
    },
    var.tags,
  )
  lifecycle {
    create_before_destroy = true
  }
}
resource "aws_nat_gateway" "multi" {
  count         = local.set_nat_multi == true ? length(var.cidrs) : 0
  allocation_id = element(aws_eip.multi.*.id, count.index)
  subnet_id     = element(aws_subnet.public.*.id, count.index)
  tags = merge(
    {
      "Name" = "${var.vpc_name}.${var.name}.nat-gw.${count.index}"
    },
    var.tags,
  )
  lifecycle {
    create_before_destroy = true
  }
}
