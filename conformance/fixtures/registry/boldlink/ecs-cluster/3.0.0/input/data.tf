data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "container_instance" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

locals {
  userdata_script = base64encode(join("\n", [
    templatefile("${path.module}/scripts/userdata.sh", {
      ssm_temp_dir              = var.ssm_agent_temp_directory
      ssm_agent_version_centos6 = var.ssm_agent_version_centos6
      ssm_urls                  = var.ssm_agent_s3_urls
    }),
    var.extra_script
  ]))
}
