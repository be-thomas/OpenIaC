terraform {
  required_version = ">= 1.0"

  backend "s3" {
    bucket = "openiac-state"
    key    = "warned.tfstate"
    region = "us-east-1"
  }
}

# A moved block is valid Terraform 1.1+ but has no target equivalent, so the
# emitters keep it out of the generated code and report it as a warning. This
# fixture exists so the golden snapshots exercise the warning section.
moved {
  from = aws_s3_bucket.old
  to   = aws_s3_bucket.assets
}

resource "aws_s3_bucket" "assets" {
  bucket = "openiac-warned-assets"
}
