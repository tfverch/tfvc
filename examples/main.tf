terraform {
  required_version = "~> 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
  }
}

# module "consul" {
#   source  = "hashicorp/consul/aws"
#   version = "~> 0.7"
# }

module "consul_github_https_missing_ref" {
  source  = "github.com/hashicorp/terraform-aws-consul"
  version = "v0.7.3"
}

# module "consul_github_https_no_ref" {
#   source = "github.com/hashicorp/terraform-aws-consul"
# }

# module "consul_github_https" {
#   source  = "github.com/hashicorp/terraform-aws-consul?ref=v0.8.0"
#   version = "0.8.0"
# }

# module "consul_github_ssh" {
#   source  = "git@github.com:hashicorp/terraform-aws-consul?ref=0.1.0"
#   version = "~0.1.0"
# }

# module "example_git_ssh_branch" {
#   source = "git::ssh://git@github.com/keilerkonzept/terraform-module-versions?ref=master"
# }

# module "example_git_scp" {
#   source  = "git::git@github.com:keilerkonzept/terraform-module-versions?ref=0.12.0"
#   version = "~> 0.12"
# }

# module "example_with_prerelease_versions" {
#   source = "git@github.com:kubernetes/api.git?ref=v0.22.2"
# }

# module "local" {
#   source = "./local"
# }
