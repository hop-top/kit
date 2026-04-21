# {{app_name}}

{{description}}

## Install

```sh
pip install {{app_name}}
```

Or with [uv](https://docs.astral.sh/uv/):

```sh
uv tool install {{app_name}}
```

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

### Commands

```sh
{{app_name}} hello
{{app_name}} hello --name You
```

## Development

Prerequisites: Python 3.12+, [uv](https://docs.astral.sh/uv/),
[Task](https://taskfile.dev)

```sh
task setup    # sync deps
task check    # lint + typecheck + test
task format   # auto-format
```

## License

See [LICENSE](LICENSE).

---
Maintained by {{author_name}}.
