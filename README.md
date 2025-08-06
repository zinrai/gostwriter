# gostwriter

A ghostwriter for your gost commands. Generates [gost](https://github.com/ginuerzh/gost) proxy command-line arguments from YAML configuration templates.

## Overview

`gostwriter` is a simple CLI tool that generates gost command-line arguments from YAML configuration files. It allows you to manage multiple proxy configurations for different environments (staging, production, etc.) using templated YAML files.

## Requirements

[gost](https://github.com/ginuerzh/gost)

## Installation

```bash
$ go install github.com/zinrai/gostwriter@latest
```

## Usage

### Basic Usage

Generate gost command for production environment:

```bash
$ gostwriter --env=production
```

Specify custom config file:

```bash
$ gostwriter --env=staging --config=/path/to/config.yml
```

Use with gost directly

```bash
$ gost $(gostwriter --env=production)
```

### Integration with Makefile

```makefile
.PHONY: proxy-staging proxy-production

proxy-staging:
	gost $$(gostwriter --env=staging)

proxy-production:
	gost $$(gostwriter --env=production)
```

### Integration with Shell Scripts

```bash
#!/bin/bash
# start-proxy.sh

ENV=${1:-staging}
GOST_ARGS=$(gostwriter --env=$ENV)

echo "Starting gost proxy for $ENV environment..."
echo "Command: gost $GOST_ARGS"

exec gost $GOST_ARGS
```

## Configuration

Refer to the `gost-config.yml` file.

For detailed command line options for gost, refer to [gost](https://github.com/ginuerzh/gost) .

### Configuration Structure

- **`defaults`**: Global variables available to all environments
- **`environments`**: Environment-specific configurations
  - **`vars`**: Environment-specific variables (override defaults)
  - **`gost_command`**: Go template that generates gost command-line arguments

### Template Variables

Templates can access variables from both `defaults` and environment-specific `vars`:

- `{{.ports.consul}}`, `{{.ports.redis}}` - Nested values from defaults
- `{{.consul_endpoint}}`, `{{.log_level}}` - From environment vars
- `${CONSUL_USER}`, `${REDIS_PASS}` - Environment variables (expanded at runtime)

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) for details.
