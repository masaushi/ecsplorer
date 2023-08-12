# ecsplorer
[![release](https://github.com/masaushi/ecsplorer/actions/workflows/release.yml/badge.svg)](https://github.com/masaushi/ecsplorer/actions/workflows/release.yml)
[![lint](https://github.com/masaushi/ecsplorer/actions/workflows/lint.yml/badge.svg)](https://github.com/masaushi/ecsplorer/actions/workflows/lint.yml)

ecsplorer is a tool designed for easy CLI operations with AWS ECS.

## Overview
This tool serves as a CLI utility to efficiently manage AWS ECS resources and services. It provides support for ECS operations through simple commands.

## Key Features
- Retrieve lists of ECS resources
- Exec into containers
- Various other ECS-related operations

## Installation
### Go version < 1.16
```sh
go get github.com/masaushi/ecsplorer
```

### Go 1.16+
```sh
go install github.com/masaushi/ecsplorer@latest
```

After installation, you can launch a terminal UI by executing the `ecsplorer` command.

### :warning: Note
If you intend to exec into containers, please ensure that the [session-manager-plugin](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html) is installed.

## License
Released under the MIT license.
