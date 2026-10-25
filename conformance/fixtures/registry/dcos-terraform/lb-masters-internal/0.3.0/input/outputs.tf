output "dns_name" {
  description = "DNS Name of the master load balancer"
  value       = module.masters-internal.dns_name
}

