# ecsplorer
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

## License
Released under the MIT license.
