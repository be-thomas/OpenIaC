output "id" {
  description = "id of resource."
  value       = aws_budgets_budget.main.id
}

output "arn" {
  description = "The ARN of the budget."
  value       = aws_budgets_budget.main.arn
}
