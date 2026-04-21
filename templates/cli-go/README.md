# {{app_name}}

{{description}}

## Install

```sh
go install {{module_prefix}}/{{app_name}}@latest
```

Or download a binary from
[Releases](https://github.com/{{author_name}}/{{app_name}}/releases).

## Usage

```sh
{{app_name}} --help
{{app_name}} --version
```

### Output formats

```sh
{{app_name}} --format json
{{app_name}} --format yaml
```

## Configuration

Config file: `~/.config/{{app_name}}/config.yaml`

Environment variables prefixed with `{{app_name_upper}}_` are
also recognized.

## Development

Prerequisites: Go 1.23+, [Task](https://taskfile.dev)

```sh
task setup    # download deps
task check    # lint + test
task build    # build binary
```

## License

See [LICENSE](LICENSE).

---
Maintained by {{author_name}}.
