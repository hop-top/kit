# Getting Started with {{app_name}}

## Quick Start

1. Install {{app_name}} (see [install](install.md))

2. Initialize a new project:
   ```sh
   {{app_name}} init
   ```

3. Run your first command:
   ```sh
   {{app_name}} help
   ```

## Basic Usage

```sh
# Show help
{{app_name}} --help

# Show version
{{app_name}} --version

# Run with verbose output
{{app_name}} --verbose <command>
```

## Configuration

{{app_name}} looks for configuration in:

1. `./{{app_name}}.yaml` (project-local)
2. `~/.config/{{app_name}}/config.yaml` (user)
3. Environment variables prefixed with
   `{{app_name}}_` (uppercase)

See [configuration](configuration.md) for details.

## Next Steps

- [Configuration Reference](configuration.md)
- [Commands Reference](commands.md)
- [Troubleshooting](troubleshooting.md)
