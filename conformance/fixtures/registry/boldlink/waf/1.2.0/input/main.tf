resource "aws_wafv2_web_acl" "main" {
  name        = var.name
  description = var.description
  scope       = var.scope
  tags        = var.tags

  default_action {
    dynamic "allow" {
      for_each = var.default_action == "allow" ? [1] : []
      content {}
    }

    dynamic "block" {
      for_each = var.default_action == "block" ? [1] : []
      content {}
    }
  }

  dynamic "custom_response_body" {
    for_each = var.custom_response_bodies
    content {
      key          = custom_response_body.value.key
      content      = custom_response_body.value.content
      content_type = custom_response_body.value.content_type
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = var.cloudwatch_metrics_enabled
    metric_name                = var.metric_name
    sampled_requests_enabled   = var.sampled_requests_enabled
  }

  dynamic "rule" {
    for_each = var.rules
    content {
      name     = rule.value.name
      priority = rule.value.priority

      dynamic "action" {
        for_each = try([rule.value.action], [])
        content {

          dynamic "allow" {
            for_each = try([action.value.allow], [])
            content {
              dynamic "custom_request_handling" {
                for_each = try([allow.value.custom_request_handling], [])
                content {
                  dynamic "insert_header" {
                    for_each = try([custom_request_handling.value.insert_header], [])
                    content {
                      name  = try(insert_header.value.name, null)
                      value = try(insert_header.value.value, null)
                    }
                  }
                }
              }
            }
          }

          dynamic "block" {
            for_each = try([action.value.block], [])
            content {
              dynamic "custom_response" {
                for_each = try([block.value.custom_response], [])
                content {
                  custom_response_body_key = try(custom_response.value.custom_response_body_key, null)
                  response_code            = custom_response.value.response_code

                  dynamic "response_header" {
                    for_each = try(custom_response.value.response_headers, [])
                    content {
                      name  = try(response_header.value.name, null)
                      value = try(response_header.value.value, null)
                    }
                  }
                }
              }
            }
          }

          dynamic "captcha" {
            for_each = try([action.value.captcha], [])
            content {
              dynamic "custom_request_handling" {
                for_each = try([captcha.value.custom_request_handling], [])
                content {
                  dynamic "insert_header" {
                    for_each = try([custom_request_handling.value.insert_header], [])
                    content {
                      name  = try(insert_header.value.name, null)
                      value = try(insert_header.value.value, null)
                    }
                  }
                }
              }
            }
          }

          dynamic "challenge" {
            for_each = try([action.value.challenge], [])
            content {
              dynamic "custom_request_handling" {
                for_each = try([challenge.value.custom_request_handling], [])
                content {
                  dynamic "insert_header" {
                    for_each = try([custom_request_handling.value.insert_header], [])
                    content {
                      name  = try(insert_header.value.name, null)
                      value = try(insert_header.value.value, null)
                    }
                  }
                }
              }
            }
          }

          dynamic "count" {
            for_each = try([action.value.count], [])
            content {
              dynamic "custom_request_handling" {
                for_each = try([count.value.custom_request_handling], [])
                content {
                  dynamic "insert_header" {
                    for_each = try([custom_request_handling.value.insert_header], [])
                    content {
                      name  = try(insert_header.value.name, null)
                      value = try(insert_header.value.value, null)
                    }
                  }
                }
              }
            }
          }

        }
      }

      ## This block only applies to rules with `rule_group_reference_statement` or `managed_rule_group_statement` blocks
      dynamic "override_action" {
        for_each = try([rule.value.override_action], [])
        content {
          dynamic "count" {
            for_each = try([override_action.value.count], [])
            content {}
          }

          dynamic "none" {
            for_each = try([override_action.value.none], [])
            content {}
          }
        }
      }

      dynamic "rule_label" {
        for_each = try(rule.value.rule_label, [])
        content {
          name = rule_label.value
        }
      }

      dynamic "statement" {
        for_each = try([rule.value.statement], [])
        content {

          dynamic "geo_match_statement" {
            for_each = try([statement.value.geo_match_statement], [])
            content {
              country_codes = geo_match_statement.value.country_codes

              dynamic "forwarded_ip_config" {
                for_each = try([geo_match_statement.value.forwarded_ip_config], [])
                content {
                  fallback_behavior = forwarded_ip_config.value.fallback_behavior
                  header_name       = forwarded_ip_config.value.header_name
                }
              }
            }
          }

          dynamic "ip_set_reference_statement" {
            for_each = length(var.ip_set_v4) > 0 ? [var.ip_set_reference_statement] : []
            content {
              arn = aws_wafv2_ip_set.ipset_v4[0].arn

              dynamic "ip_set_forwarded_ip_config" {
                for_each = try([ip_set_reference_statement.value.ip_set_forwarded_ip_config], [])
                content {
                  fallback_behavior = ip_set_forwarded_ip_config.value.fallback_behavior
                  header_name       = ip_set_forwarded_ip_config.value.header_name
                  position          = ip_set_forwarded_ip_config.value.position
                }
              }
            }
          }

          dynamic "ip_set_reference_statement" {
            for_each = length(var.ip_set_v6) > 0 ? [var.ip_set_reference_statement] : []
            content {
              arn = aws_wafv2_ip_set.ipset_v6[0].arn

              dynamic "ip_set_forwarded_ip_config" {
                for_each = try([ip_set_reference_statement.value.ip_set_forwarded_ip_config], [])
                content {
                  fallback_behavior = ip_set_forwarded_ip_config.value.fallback_behavior
                  header_name       = ip_set_forwarded_ip_config.value.header_name
                  position          = ip_set_forwarded_ip_config.value.position
                }
              }
            }
          }

          dynamic "managed_rule_group_statement" {
            for_each = try([statement.value.managed_rule_group_statement], [])
            content {
              name        = managed_rule_group_statement.value.name
              vendor_name = managed_rule_group_statement.value.vendor_name
              version     = try(managed_rule_group_statement.value.version, null)

              # Add support for rule_action_override with fixed structure
              dynamic "rule_action_override" {
                for_each = try(managed_rule_group_statement.value.rule_action_override, [])
                content {
                  name = rule_action_override.value.name

                  # The action_to_use block is required
                  action_to_use {
                    # Only one of these can be specified
                    dynamic "count" {
                      for_each = try([rule_action_override.value.action_to_use.count], [])
                      content {}
                    }

                    dynamic "block" {
                      for_each = try([rule_action_override.value.action_to_use.block], [])
                      content {}
                    }

                    dynamic "captcha" {
                      for_each = try([rule_action_override.value.action_to_use.captcha], [])
                      content {}
                    }

                    dynamic "challenge" {
                      for_each = try([rule_action_override.value.action_to_use.challenge], [])
                      content {}
                    }
                  }
                }
              }

              # Removed the excluded_rule block that was causing errors

              # Scope down statement remains as it's different from excluded_rule
              dynamic "scope_down_statement" {
                for_each = try([managed_rule_group_statement.value.scope_down_statement], [])
                content {
                  # Byte match statement support for scope_down
                  dynamic "byte_match_statement" {
                    for_each = try([scope_down_statement.value.byte_match_statement], [])
                    content {
                      search_string         = byte_match_statement.value.search_string
                      positional_constraint = byte_match_statement.value.positional_constraint

                      # field_to_match
                      dynamic "field_to_match" {
                        for_each = try([byte_match_statement.value.field_to_match], [])
                        content {
                          # single_header field match
                          dynamic "single_header" {
                            for_each = try([field_to_match.value.single_header], [])
                            content {
                              name = single_header.value.name
                            }
                          }

                          # method field match
                          dynamic "method" {
                            for_each = try([field_to_match.value.method], [])
                            content {}
                          }

                          # uri_path field match
                          dynamic "uri_path" {
                            for_each = try([field_to_match.value.uri_path], [])
                            content {}
                          }

                          # query_string field match
                          dynamic "query_string" {
                            for_each = try([field_to_match.value.query_string], [])
                            content {}
                          }

                          # body field match
                          dynamic "body" {
                            for_each = try([field_to_match.value.body], [])
                            content {}
                          }

                          # all_query_arguments field match
                          dynamic "all_query_arguments" {
                            for_each = try([field_to_match.value.all_query_arguments], [])
                            content {}
                          }
                        }
                      }

                      # text_transformation
                      dynamic "text_transformation" {
                        for_each = try(byte_match_statement.value.text_transformation, [])
                        content {
                          priority = text_transformation.value.priority
                          type     = text_transformation.value.type
                        }
                      }
                    }
                  }
                }
              }
            }
          }

          # byte_match_statement support
          dynamic "byte_match_statement" {
            for_each = try([statement.value.byte_match_statement], [])
            content {
              search_string         = byte_match_statement.value.search_string
              positional_constraint = byte_match_statement.value.positional_constraint

              # field_to_match block
              dynamic "field_to_match" {
                for_each = try([byte_match_statement.value.field_to_match], [])
                content {
                  # single_header field match
                  dynamic "single_header" {
                    for_each = try([field_to_match.value.single_header], [])
                    content {
                      name = single_header.value.name
                    }
                  }

                  # method field match
                  dynamic "method" {
                    for_each = try([field_to_match.value.method], [])
                    content {}
                  }

                  # uri_path field match
                  dynamic "uri_path" {
                    for_each = try([field_to_match.value.uri_path], [])
                    content {}
                  }

                  # query_string field match
                  dynamic "query_string" {
                    for_each = try([field_to_match.value.query_string], [])
                    content {}
                  }

                  # body field match
                  dynamic "body" {
                    for_each = try([field_to_match.value.body], [])
                    content {}
                  }

                  # all_query_arguments field match
                  dynamic "all_query_arguments" {
                    for_each = try([field_to_match.value.all_query_arguments], [])
                    content {}
                  }
                }
              }

              # text_transformation block(s)
              dynamic "text_transformation" {
                for_each = try(byte_match_statement.value.text_transformation, [])
                content {
                  priority = text_transformation.value.priority
                  type     = text_transformation.value.type
                }
              }
            }
          }
        }
      }

      dynamic "visibility_config" {
        for_each = try([rule.value.visibility_config], [])
        content {
          cloudwatch_metrics_enabled = visibility_config.value.cloudwatch_metrics_enabled
          metric_name                = visibility_config.value.metric_name
          sampled_requests_enabled   = visibility_config.value.sampled_requests_enabled
        }
      }
    }
  }
}

resource "aws_wafv2_web_acl_association" "main" {
  count        = var.create_acl_association ? 1 : 0
  resource_arn = var.web_acl_resource_arn
  web_acl_arn  = aws_wafv2_web_acl.main.arn
  depends_on   = [aws_wafv2_web_acl.main]
}

resource "aws_wafv2_ip_set" "ipset_v4" {
  count              = length(var.ip_set_v4) > 0 ? length(var.ip_set_v4) : 0
  name               = var.ip_set_v4[count.index]["name"]
  description        = try(var.ip_set_v4[count.index]["description"], null)
  scope              = var.ip_set_v4[count.index]["scope"]
  ip_address_version = "IPV4"
  addresses          = try(var.ip_set_v4[count.index]["addresses"], null)
  tags               = var.tags
}

resource "aws_wafv2_ip_set" "ipset_v6" {
  count              = length(var.ip_set_v6) > 0 ? length(var.ip_set_v6) : 0
  name               = var.ip_set_v6[count.index]["name"]
  description        = try(var.ip_set_v6[count.index]["description"], null)
  scope              = var.ip_set_v6[count.index]["scope"]
  ip_address_version = "IPV6"
  addresses          = try(var.ip_set_v6[count.index]["addresses"], null)
  tags               = var.tags
}

# Simple logging configuration without dynamic blocks or experimental features
resource "aws_wafv2_web_acl_logging_configuration" "main" {
  count                   = var.enable_logging ? 1 : 0
  log_destination_configs = var.log_destination_configs
  resource_arn            = aws_wafv2_web_acl.main.arn

  # Only use redacted_fields if provided
  dynamic "redacted_fields" {
    for_each = var.redacted_fields != null ? var.redacted_fields : []
    content {
      single_header {
        name = lookup(redacted_fields.value, "single_header", null) != null ? redacted_fields.value.single_header.name : null
      }
    }
  }

  # Logging filter without experimental features
  dynamic "logging_filter" {
    for_each = var.logging_filter != null ? [var.logging_filter] : []
    content {
      default_behavior = logging_filter.value.default_behavior

      dynamic "filter" {
        for_each = lookup(logging_filter.value, "filters", [])
        content {
          behavior    = filter.value.behavior
          requirement = filter.value.requirement

          dynamic "condition" {
            for_each = lookup(filter.value, "conditions", [])
            content {
              # Handle action_condition if it exists
              dynamic "action_condition" {
                for_each = lookup(condition.value, "action_condition", null) != null ? [condition.value.action_condition] : []
                content {
                  action = action_condition.value
                }
              }

              # Handle label_name_condition if it exists
              dynamic "label_name_condition" {
                for_each = lookup(condition.value, "label_name_condition", null) != null ? [condition.value.label_name_condition] : []
                content {
                  label_name = label_name_condition.value
                }
              }
            }
          }
        }
      }
    }
  }

  depends_on = [aws_wafv2_web_acl.main]
}
