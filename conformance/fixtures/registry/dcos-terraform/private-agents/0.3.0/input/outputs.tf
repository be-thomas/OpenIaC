output "instances" {
  description = "List of instance IDs"
  value       = module.dcos-private-agent-instances.instances
}

output "public_ips" {
  description = "List of public ip addresses created by this module"
  value       = module.dcos-private-agent-instances.public_ips
}

output "private_ips" {
  description = "List of private ip addresses created by this module"
  value       = module.dcos-private-agent-instances.private_ips
}

output "os_user" {
  description = "The OS user to be used"
  value       = module.dcos-private-agent-instances.os_user
}

