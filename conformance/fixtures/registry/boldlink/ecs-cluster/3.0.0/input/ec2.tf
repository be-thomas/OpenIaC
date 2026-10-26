## Security group
resource "aws_security_group" "this" {
  count       = var.create_security_group && var.create_ec2_instance ? 1 : 0
  name        = var.name
  vpc_id      = var.vpc_id
  description = var.security_group_description
  tags        = var.tags
}

resource "aws_security_group_rule" "ingress" {
  for_each          = var.ingress_rules
  type              = "ingress"
  description       = var.security_group_ingress_description
  from_port         = lookup(each.value, "from_port")
  to_port           = lookup(each.value, "to_port")
  protocol          = var.security_group_default_protocol
  cidr_blocks       = lookup(each.value, "cidr_blocks", null)
  security_group_id = join("", aws_security_group.this.*.id)
}

resource "aws_security_group_rule" "egress" {
  for_each          = var.egress_rules
  type              = "egress"
  description       = var.security_group_egress_description
  from_port         = lookup(each.value, "from_port")
  to_port           = lookup(each.value, "to_port")
  protocol          = var.security_group_default_protocol
  cidr_blocks       = lookup(each.value, "cidr_blocks", null)
  security_group_id = join("", aws_security_group.this.*.id)
}

# Iam role
resource "aws_iam_instance_profile" "this" {
  count = var.create_ec2_instance ? 1 : 0
  name  = var.name
  role  = aws_iam_role.cluster_instance[0].name
  tags  = var.tags
}

resource "aws_iam_role" "cluster_instance" {
  count              = var.create_ec2_instance ? 1 : 0
  name               = var.name
  assume_role_policy = data.aws_iam_policy_document.container_instance.json
  tags               = var.tags
}

resource "aws_iam_role_policy_attachment" "ecs_cluster_ec2_role" {
  count      = var.create_ec2_instance ? 1 : 0
  role       = aws_iam_role.cluster_instance[0].id
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}


## Managed Policy to allow ssm agent to communicate with SSM Manager
resource "aws_iam_role_policy_attachment" "ssm" {
  count      = var.create_ec2_instance && var.install_ssm_agent ? 1 : 0
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
  role       = aws_iam_role.cluster_instance[0].id
}


# Launch template
resource "aws_launch_template" "this" {
  count         = var.create_ec2_instance ? 1 : 0
  name_prefix   = var.name
  description   = "Launch template for ${var.name} ECS cluster"
  image_id      = var.image_id
  instance_type = var.instance_type
  key_name      = var.key_name

  iam_instance_profile {
    name = aws_iam_instance_profile.this[0].name
  }

  user_data = var.install_ssm_agent ? local.userdata_script : var.user_data

  monitoring {
    enabled = var.monitoring_enabled
  }

  network_interfaces {
    associate_public_ip_address = var.associate_public_ip_address
    delete_on_termination       = var.delete_on_termination
    security_groups             = try(aws_security_group.this.*.id, "")
    subnet_id                   = var.subnet_id
  }

  dynamic "block_device_mappings" {
    for_each = local.processed_block_device_mappings
    content {
      device_name  = block_device_mappings.value.device_name
      no_device    = lookup(block_device_mappings.value, "no_device", null)
      virtual_name = lookup(block_device_mappings.value, "virtual_name", null)

      dynamic "ebs" {
        for_each = block_device_mappings.value.ebs != null ? [block_device_mappings.value.ebs] : []
        content {
          encrypted             = try(ebs.value.encrypted, null)
          kms_key_id            = try(ebs.value.kms_key_id, null)
          delete_on_termination = try(ebs.value.delete_on_termination, null)
          iops                  = try(ebs.value.iops, null)
          throughput            = try(ebs.value.throughput, null)
          snapshot_id           = try(ebs.value.snapshot_id, null)
          volume_size           = try(ebs.value.volume_size, null)
          volume_type           = try(ebs.value.volume_type, null)
        }
      }
    }
  }

  metadata_options {
    http_endpoint               = lookup(var.metadata_options, "http_endpoint", var.metadata_options_defaults.http_endpoint)
    http_put_response_hop_limit = lookup(var.metadata_options, "http_put_response_hop_limit", null)
    http_tokens                 = lookup(var.metadata_options, "http_tokens", var.metadata_options_defaults.http_tokens)
    http_protocol_ipv6          = lookup(var.metadata_options, "value.http_protocol_ipv6", null)
    instance_metadata_tags      = lookup(var.metadata_options, "value.instance_metadata_tags", null)
  }

  lifecycle {
    create_before_destroy = true
  }

  tags = var.tags

  tag_specifications {
    resource_type = "instance"
    tags          = var.tags
  }

  tag_specifications {
    resource_type = "volume"
    tags          = var.tags
  }
}

# Auto-scaling Group
resource "aws_autoscaling_group" "container_instance" {
  count = var.create_ec2_instance ? 1 : 0
  name  = var.name

  launch_template {
    id      = aws_launch_template.this[0].id
    version = var.launch_template_version
  }

  dynamic "tag" {
    for_each = var.tags
    content {
      key                 = tag.key
      value               = tag.value
      propagate_at_launch = var.asg_tags_propagate_at_launch
    }
  }

  dynamic "tag" {
    for_each = var.enable_managed_scaling ? [1] : []
    content {
      key                 = var.ecs_managed_tag_key
      value               = var.ecs_managed_tag_value
      propagate_at_launch = var.ecs_managed_tag_propagate_at_launch
    }
  }

  availability_zones = var.availability_zones
  desired_capacity   = var.desired_capacity
  max_size           = var.max_size
  min_size           = var.min_size

  # Enable instance protection when managed scaling is enabled
  protect_from_scale_in = var.protect_from_scale_in_independent != null ? var.protect_from_scale_in_independent : var.enable_managed_scaling

  lifecycle {
    create_before_destroy = true
  }
}
