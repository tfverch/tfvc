terraform {
  required_version = "~> 1.0"

  required_providers {
    # AWS provider configured with outdated major version here and in lockfile - should fail
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
    # Google provider configured with correct constraint here but lockfile is out of date - should warn
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
  }
}

# Registry module configured with acceptable constraint - should pass
module "consul" {
  source  = "hashicorp/consul/aws"
  version = "~> 0.7"
}

# Git module configured with no version constraints - should fail
# module "consul_github_https_no_ref" {
#   source = "github.com/hashicorp/terraform-aws-consul"
# }

# Git module configured with outdated static version - should warn
# module "consul_github_https" {
#   source  = "github.com/hashicorp/terraform-aws-consul?ref=v0.8.0"
# }

# module "example_git_scp" {
#   source  = "git::git@github.com:keilerkonzept/terraform-module-versions?ref=0.12.0"
#   version = "~> 0.12"
# }

# SSH github module configured with ref but uses prerelease versions - should warn when using -e switch
# module "example_with_prerelease_versions" {
#   source = "git@github.com:kubernetes/api.git?ref=v0.22.2"
# }
