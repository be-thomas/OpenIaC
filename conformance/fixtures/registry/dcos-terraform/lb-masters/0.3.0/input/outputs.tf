output "dns_name" {
  description = "DNS Name of the master load balancer"
  value       = module.masters.dns_name
}

