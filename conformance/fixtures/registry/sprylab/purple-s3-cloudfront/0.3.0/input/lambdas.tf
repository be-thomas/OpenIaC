module "lambdas" {
  source              = "./modules/lambdas"
  bucket_name         = var.bucket_name
  bucket_arn          = module.bucket.s3_bucket_arn
  tachyon_memory_size = var.tachyon_memory_size
}
