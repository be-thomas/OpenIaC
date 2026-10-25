variable "name" {
  type        = string
  description = "The name of a budget. Unique within accounts."
  default     = null
}

variable "name_prefix" {
  type        = string
  description = "The prefix of the name of a budget. Unique within accounts."
  default     = null
}

variable "account_id" {
  type        = number
  description = "The ID of the target account for budget. Will use current user's account_id by default if omitted"
  default     = null
}

variable "budget_type" {
  type        = string
  description = "Whether this budget tracks monetary cost or usage. Valid values include `COST`, `USAGE`, `SAVINGS_PLANS_UTILIZATION`, `RI_UTILIZATION`"
}

variable "limit_amount" {
  type        = number
  description = "The amount of cost or usage being measured for a budget."
  default     = null
}

variable "limit_unit" {
  type        = string
  description = "The unit of measurement used for the budget forecast, actual spend, or budget threshold, such as USD or GBP."
  default     = null
}

variable "time_period_end" {
  type        = string
  description = "The end of the time period covered by the budget. There are no restrictions on the end date. `Format: 2017-01-01_12:00`."
  default     = null
}

variable "time_period_start" {
  type        = string
  description = "The start of the time period covered by the budget. If you don't specify a start date, AWS defaults to the start of your chosen time period. The start date must come before the end date. `Format: 2017-01-01_12:00`."
  default     = null
}

variable "time_unit" {
  type        = string
  description = "The length of time until a budget resets the actual and forecasted spend. Valid values: MONTHLY, QUARTERLY, ANNUALLY, and DAILY."
}

variable "auto_adjust_data" {
  type        = any
  description = "The parameters that determine the budget amount for an auto-adjusting budget."
  default     = {}
}

variable "cost_filter" {
  type        = any
  description = "A list of CostFilter name/values pair to apply to budget. RI UtilizatioSn plans require a service cost filter to be set"
  default     = []
}

variable "planned_limit" {
  type        = any
  description = "Object containing Budget Notifications. Can be used multiple times to define more than one budget notification."
  default     = []
}

variable "cost_types" {
  type        = map(string)
  description = "Object containing CostTypes The types of cost included in a budget, such as tax and subscriptions. Cost types must be defined for RI budgets because the settings conflict with the defaults"
  default     = {}
}

variable "notification" {
  type        = any
  description = "Object containing Budget Notifications."
  default     = []
}
