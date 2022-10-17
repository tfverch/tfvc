# Getting Started

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

`tfvc` will scan the specified directory and report on the version configuration for providers and modules.

!!! info

    `tfvc` will return a non-zero exit status if if finds any issues, otherwise the exit status will be zero.

The following example will run `tfvc` against the current working directory (`.`).

```bash
tfvc .
```

## Docker usage

As an alternative to installing and running `tfvc` on your system, you can run it in a Docker container.

The following example will mount the current working directory (`$(pwd)`) in the  container and run `tfvc`.

```bash
docker run --rm -it -v "$(pwd):/src" tfverch/tfvc /src
```
