## SG for Interface endpoint
resource "aws_security_group" "allow_443" {
  count       = length(var.vpc_endpoints) > 0 && var.create_endpoint_sg ? 1 : 0
  name        = try(var.security_group_name, "vpce.interface_endpoint.allow_443")
  description = "Allow VPC Endpoints SSL/TLS inbound traffic"
  vpc_id      = var.vpc_id
  tags = merge(
    { "Name" = "Interface-Endpoint-SG" },
    var.tags
  )

  ingress {
    description = "Allow traffic for Interface VPC Endpoints"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [local.vpc_cidr_block]
  }
}

resource "aws_vpc_endpoint" "endpoint" {
  count               = length(var.vpc_endpoints) > 0 ? length(var.vpc_endpoints) : 0
  vpc_id              = var.vpc_id
  service_name        = try(var.vpc_endpoints[count.index]["service_name"], null)
  auto_accept         = try(var.vpc_endpoints[count.index]["auto_accept"], null)
  policy              = try(var.vpc_endpoints[count.index]["policy"], null)
  private_dns_enabled = try(var.vpc_endpoints[count.index]["private_dns_enabled"], false)
  ip_address_type     = try(var.vpc_endpoints[count.index]["ip_address_type"], null)
  route_table_ids     = try(var.vpc_endpoints[count.index]["route_table_ids"], [])
  subnet_ids          = try(var.vpc_endpoints[count.index]["subnet_ids"], [])
  vpc_endpoint_type   = try(var.vpc_endpoints[count.index]["vpc_endpoint_type"], "Interface")

  dynamic "dns_options" {
    for_each = try(var.vpc_endpoints[count.index]["dns_options"], {})
    content {
      dns_record_ip_type                             = try(dns_options.value.dns_record_ip_type, null)
      private_dns_only_for_inbound_resolver_endpoint = try(dns_options.value.private_dns_only_for_inbound_resolver_endpoint, null)
    }
  }

  dynamic "timeouts" {
    for_each = try([var.vpc_endpoints[count.index]["timeouts"]], [])
    content {
      create = try(timeouts.value.create, null)
      update = try(timeouts.value.update, null)
      delete = try(timeouts.value.delete, null)
    }
  }

  tags = merge(
    try(var.vpc_endpoints[count.index]["name"], null) != null ? { "Name" = var.vpc_endpoints[count.index]["name"] } : {},
    var.tags
  )

  security_group_ids = local.include_security_group[count.index] ? [aws_security_group.allow_443[0].id] : []
}
