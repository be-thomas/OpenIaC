output "dns_name" {
  description = "DNS Name of the master load balancer"
  value       = length(var.instances) < 1 ? "" : join(",", aws_lb.loadbalancer.*.dns_name)
}

