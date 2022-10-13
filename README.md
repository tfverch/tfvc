# Terraform Version Check

[![release](https://img.shields.io/github/v/release/tfverch/tfvc?display_name=tag&color=blueviolet)](https://github.com/tfverch/tfvc/releases)
[![GoReportCard](https://goreportcard.com/badge/github.com/tfverch/tfvc)](https://goreportcard.com/report/github.com/tfverch/tfvc)
[![Go version](https://img.shields.io/github/go-mod/go-version/tfverch/tfvc.svg)](https://github.com/tfverch/tfvc)

Terraform version check (tfvc) is a reporting tool to identify available updates for providers and modules referenced in your Terraform code. It provides clear warning/failure
output and resolution guidance for any issues it detects.

**NOTE: This project is currently under heavy development and things WILL break (probably)**

## Example output

![Example output](example-output.png)

## Installation

Install with Homebrew on MacOS or Linux.

```bash
brew install tfverch/tfvc/tfvc
```

Install with Go

```bash
go install github.com/tfverch/tfvc@latest
```

## Usage

tfvc will scan the specified directories and report on the configuration of providers and module calls.

The exit status will be non-zero if tfvc finds problems, otherwise the exit status will be zero.

```bash
tfvc .
```

The following parameters are available.

| Parameter                   | Type                  | Description                                     |
| --------------------------- | --------------------- | ----------------------------------------------- |
| --include-passed, -a        | bool (default: false) | Include passed checks in console output         |
| --include-prerelease, -e    | bool (default: false) | Include prerelease versions in checks           |
| --ssh-private-key-path, -s, | string (default: "")  | Path to private key to use for SSH module calls |
| --ssh-private-key-pwd, -w   | string (default: "")  | Password for private key file if required       |

## Docker usage

As an alternative to installing and running tfvc on your system, you can run tfvc in a Docker container, for example:

```bash
docker run --rm -it -v "$(pwd):/src" tfverch/tfvc /src
```

## Motivation

While tools such as [dependabot](https://github.com/dependabot) and [renovate](https://docs.renovatebot.com/) provide fully automate dependency updates, I needed something with a lighter touch. tfvc aims to be a simple reporting tool that can be run either locally, or as part of a CI/CD pipeline, to give you feedback on any modules or providers which have updates available.

## Acknowledgements

This project started as a fork of the [github.com/keilerkonzept/terraform-module-versions](https://github.com/keilerkonzept/terraform-module-versions) project. However, given the changes that I needed to make to add the features that I wanted to, I ended up migrating to this repo. Still, shout out to [keilerkonzept](https://github.com/keilerkonzept) for their work.
