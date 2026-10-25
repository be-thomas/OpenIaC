
resource "aws_ecrpublic_repository" "this" {
  count           = var.create_public_repository ? 1 : 0
  repository_name = var.repository_name

  dynamic "catalog_data" {
    for_each = var.catalog_data
    content {
      about_text        = lookup(catalog_data.value, "about_text", null)
      architectures     = lookup(catalog_data.value, "architectures", [])
      description       = lookup(catalog_data.value, "description", null)
      logo_image_blob   = lookup(catalog_data.value, "logo_image_blob", null)
      operating_systems = lookup(catalog_data.value, "operating_systems", [])
      usage_text        = lookup(catalog_data.value, "usage_text", null)
    }
  }
  timeouts {
    delete = lookup(var.timeouts, "delete", "20m")
  }
}

resource "aws_ecrpublic_repository_policy" "this" {
  count           = var.create_ecrpublic_repository_policy && var.create_public_repository ? 1 : 0
  repository_name = aws_ecrpublic_repository.this[0].repository_name
  policy          = var.policy
}
