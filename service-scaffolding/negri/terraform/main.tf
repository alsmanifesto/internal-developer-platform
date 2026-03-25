terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# LocalStack-compatible AWS provider configuration
provider "aws" {
  region                      = "us-east-1"
  access_key                  = "test"
  secret_key                  = "test"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  s3_use_path_style = true

  endpoints {
    s3 = "http://localhost:4566"
  }
}

resource "aws_s3_bucket" "negri" {
  bucket = "negri-bucket"

  tags = {
    Name        = "negri"
    ServiceType = "job"
    Workload    = "app"
    Stack       = "go"
    ManagedBy   = "scaffold"
  }
}

resource "aws_s3_bucket_versioning" "negri_versioning" {
  bucket = aws_s3_bucket.negri.id

  versioning_configuration {
    status = "Enabled"
  }
}

output "bucket_name" {
  value = aws_s3_bucket.negri.bucket
}
