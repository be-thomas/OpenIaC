output "user" {
  description = "User"
  value       = data.template_file.aws_ami_user.rendered
}

output "aws_ami" {
  description = "AMI that will be used for the instances instead of the Mesosphere chosen default images. Custom AMIs must fulfill the Mesosphere DC/OS system-requirements: See https://docs.mesosphere.com/1.12/installing/production/system-requirements/"
  value       = data.template_file.aws_ami.rendered
}

