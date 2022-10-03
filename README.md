# Terraform Version Check

[![release](https://img.shields.io/github/v/release/tfverch/tfvc?display_name=tag&color=blueviolet)](https://github.com/tfverch/tfvc/releases)
[![GoReportCard](https://goreportcard.com/badge/github.com/tfverch/tfvc)](https://goreportcard.com/report/github.com/tfverch/tfvc)
[![Go version](https://img.shields.io/github/go-mod/go-version/tfverch/tfvc.svg)](https://github.com/tfverch/tfvc)

Terraform version check (tfvc) is a tool for ensuring that your your Terraform code is always configured to use the latest versions of any referenced providers and modules.

**NOTE: This project is currently under heavy development and things WILL break (probably)**

## Installation

Install with Homebrew on MacOS or Linux.

```bash
brew install tfverch/tfvc/tfvc
```

Install with Go

```bash
go install github.com/tfverch/tfvc@latest
```

## Acknowledgements

This project started as a fork of the [github.com/keilerkonzept/terraform-module-versions](https://github.com/keilerkonzept/terraform-module-versions) project. However, given the changes that I needed to make in order to add the features that I wanted to I ended up migrating to this repo. Still, shout out to [keilerkonzept](https://github.com/keilerkonzept) for their work.
