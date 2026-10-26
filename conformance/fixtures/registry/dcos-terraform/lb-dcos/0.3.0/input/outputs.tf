output "masters_dns_name" {
  description = "DNS Name of the master load balancer"
  value       = module.dcos-lb-masters.dns_name
}

output "masters_internal_dns_name" {
  description = "DNS Name of the master load balancer"
  value       = module.dcos-lb-masters-internal.dns_name
}

output "public_agents_dns_name" {
  description = "DNS Name of the public agent load balancer"
  value       = module.dcos-lb-public-agents.dns_name
}

