variable "database_name" {
  type = string

  description = <<EOS
Name of the database on to be used in the `DATABASE_URL`.
EOS
}

variable "db_instance" {
  type = object({
    address  = string
    password = string
    username = string
  })
  default = null

  description = <<EOS
Database instance to be used in the `DATABASE_URL`.
EOS
}

variable "default_tags" {
  type    = map(string)
  default = {}

  description = <<EOS
Map of tags assigned to all AWS resources created by this module.
EOS
}

variable "name_prefix" {
  type = string

  description = <<EOS
Name prefix for the SecretsManager. The full name will be $${var.name_prefix}.database_url.
EOS
}

variable "protocol" {
  type = string

  description = <<EOS
Protocol to be used in the `DATABASE_URL`.
EOS
}

variable "rds_cluster" {
  type = object({
    master_username = string
    master_password = string
    endpoint        = string
    reader_endpoint = string
  })
  default = null

  description = <<EOS
Database cluster to be used in the `DATABASE_URL`.
EOS
}

variable "secretsmanager_secret_tags" {
  type    = map(string)
  default = {}

  description = <<EOS
Map of tags assigned to the SecretsManager secret created by this module. Tags in this map will override tags in `var.default_tags`.
EOS
}
