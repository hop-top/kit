# Python CLI API Reference

Package: `hop-top-kit` (`hop_top_kit.cli`)

## create_app

```python
def create_app(
    *,
    name: str,
    version: str,
    help: str,
) -> typer.Typer
```

Returns a Typer app pre-configured to the hop-top CLI
contract:

- `add_completion=False` — no `--install-completion`
- `no_args_is_help=True` — bare invocation shows help
- `-v, --version` prints `<name> <version>` and exits
- Root callback with `invoke_without_command=True`

## Command Groups

### GroupConfig

```python
@dataclass
class GroupConfig:
    id: str        # unique identifier (e.g. "management")
    title: str     # display title (e.g. "MANAGEMENT COMMANDS")
    hidden: bool   # True = excluded from default --help
```

### HelpConfig.groups

Groups are declared via `HelpConfig` and passed to
`create_app`:

```python
@dataclass
class HelpConfig:
    groups: list[GroupConfig] = field(default_factory=list)
```

Default groups when none specified:

| id | title | hidden |
|----|-------|--------|
| `commands` | COMMANDS | False |
| `management` | MANAGEMENT COMMANDS | True |

### set_command_group

```python
def set_command_group(name: str, group_id: str) -> None
```

Assigns a registered command to a named group by its
command name. Commands without assignment default to the
`commands` group.

```python
app = create_app(name="mytool", version="1.0.0", help="...")

@app.command()
def config():
    """Manage configuration."""
    ...

set_command_group("config", "management")
```

### --help-all

Registered as a root-level eager option. When passed, the
help formatter includes commands from all groups, including
hidden ones.

```
$ mytool --help          # shows COMMANDS only
$ mytool --help-all      # shows COMMANDS + MANAGEMENT
```

## Output

Package: `hop_top_kit.output`

Provides `render_table`, `render_json`, `render_yaml` for
consistent output formatting.
