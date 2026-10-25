module "private" {
  source                    = "./modules/private"
  create_private_repository = var.create_private_repository
  name                      = var.name
  image_tag_mutability      = var.image_tag_mutability
  encryption_configuration  = var.encryption_configuration
  aws_ecr_lifecycle_policy  = var.aws_ecr_lifecycle_policy
  aws_ecr_repository_policy = var.aws_ecr_repository_policy
  scan_on_push              = var.scan_on_push
  other_tags                = var.other_tags
}

module "public" {
  source                             = "./modules/public"
  create_public_repository           = var.create_public_repository
  repository_name                    = var.name
  catalog_data                       = var.catalog_data
  timeouts                           = var.timeouts
  create_ecrpublic_repository_policy = var.create_ecrpublic_repository_policy
  policy                             = var.policy
}
