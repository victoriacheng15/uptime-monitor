terraform {
  required_version = ">= 1.6.0"

  backend "s3" {
    bucket = "uptime-monitor-tfstate"
    key    = "uptime-monitor/tofu.tfstate"
    region = "ca-central-1"
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = local.tags
  }
}

locals {
  tags = {
    Project = var.project_name
  }
}

module "s3" {
  source = "./modules/s3"

  history_bucket_name = var.history_bucket_name
  tfstate_bucket_name = var.tfstate_bucket_name
}

module "lambda" {
  source = "./modules/lambda"

  function_name       = var.lambda_function_name
  package_file        = "../bin/lambda.zip"
  history_bucket_arn  = module.s3.history_bucket_arn
  history_bucket_name = module.s3.history_bucket_name
}
