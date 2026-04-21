# Commands Reference: {{app_name}}

## Global Flags

```
--help, -h        Show help
--version         Show version
--verbose, -v     Enable verbose output
--format, -f FMT  Output format (text|json|yaml)
```

## Commands

### `init`

Initialize a new project.

```sh
{{app_name}} init [directory]
```

**Flags:**
- `--template` — Template to use (default: "default")

### `version`

Print version information.

```sh
{{app_name}} version
```

### `help`

Show help for any command.

```sh
{{app_name}} help [command]
```

## Adding Commands

See the architecture docs and language-specific guides
for how to add new commands to {{app_name}}.
