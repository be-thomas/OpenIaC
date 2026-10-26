output "subnets" {
  description = "Output all subnet information"
  value       = aws_subnet.public[*].id
}

output "route_tables" {
  description = "List of IDs of all route tables"
  value = concat(
    aws_route_table.public[*].id
  )
}