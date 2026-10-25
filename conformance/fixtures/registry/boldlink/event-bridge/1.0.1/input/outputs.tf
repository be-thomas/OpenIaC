output "rule_id" {
  value       = aws_cloudwatch_event_rule.main.id
  description = "The name of the rule."
}

output "rule_arn" {
  value       = aws_cloudwatch_event_rule.main.arn
  description = "The Amazon Resource Name (ARN) of the rule."
}

output "rule_tags_all" {
  value       = aws_cloudwatch_event_rule.main.tags_all
  description = "A map of tags assigned to the resource, including those inherited from the provider default_tags configuration block."
}
