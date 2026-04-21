# Help Rendering

kit/cli customizes help output across all three languages
to enforce a consistent look and feel.

## Standard Help Layout

```
<short description>

Usage:
  <name> [command] [flags]

COMMANDS:
  run         Execute a task
  list        Show all items

FLAGS:
  -h, --help       Display help
  -v, --version    Print version and exit
      --format     Output format (table|json|yaml)
      --quiet      Suppress non-essential output
      --no-color   Disable ANSI colour
      --help-all   Show all command groups
```

## Command Groups

Groups partition commands into sections. Each group has
an ID, title, and visibility flag.

### Default Groups

| Group | Title | Hidden | Contains |
|-------|-------|--------|----------|
| `commands` | COMMANDS | no | user-facing commands |
| `management` | MANAGEMENT | yes | config, toolspec, doctor |

### Custom Groups

Developers define additional groups via config:

```
Go:     Help: cli.HelpConfig{Groups: []cli.GroupConfig{
            {ID: "extras", Title: "EXTRAS"},
            {ID: "debug", Title: "DEBUG", Hidden: true},
        }}
TS:     groups: [
            { id: 'extras', title: 'EXTRAS' },
            { id: 'debug', title: 'DEBUG', hidden: true },
        ]
Python: help_config=HelpConfig(groups=[
            GroupConfig(id='extras', title='EXTRAS'),
            GroupConfig(id='debug', title='DEBUG', hidden=True),
        ])
```

### Assigning Commands to Groups

```
Go:     cmd.GroupID = "management"
TS:     setCommandGroup(cmd, 'management')
Python: set_command_group('config', 'management')
```

Unassigned commands go to the first group (default:
`commands`).

## Help Modes

### `--help` (default)

Shows only visible groups. Hidden groups and their
commands are suppressed.

```
COMMANDS:
  deploy        Deploy the app
  mission       Query mission history

EXTRAS:
  bonus         Bonus feature

FLAGS:
  ...
```

### `--help-all`

Shows ALL groups including hidden ones.

```
COMMANDS:
  deploy        Deploy the app
  mission       Query mission history

MANAGEMENT:
  config        Show configuration
  toolspec      Validate toolspec

EXTRAS:
  bonus         Bonus feature

FLAGS:
  ...
```

### `--help-<id>` (per-group)

Shows ONLY the named group's commands + FLAGS.

```
$ spaced --help-management

MANAGEMENT:
  config        Show configuration
  toolspec      Validate toolspec

FLAGS:
  ...
```

### `help <id>` subcommand

Same as `--help-<id>` but as a subcommand:

```
$ spaced help management    # → same as --help-management
$ spaced help extras        # → same as --help-extras
$ spaced help all           # → same as --help-all
```

The `help` subcommand is hidden from `--help` output.

## Structural Rendering (No Regex)

All three languages render help structurally — no regex
post-processing on output text.

### Go (fang)

fang renders groups natively via cobra's `AddGroup` +
`GroupID`. kit registers groups in `New()`, controls
visibility via `cmd.Hidden` before `Execute()`. The
`--help-<id>` flags scan args pre-parse to filter.

### TypeScript (Commander)

The structural `formatHelp` override partitions
`visibleCommands` by group (via `WeakMap`). Reads
`--help-<id>` from `process.argv` to filter. The
`help [group]` subcommand injects the corresponding
argv flag.

### Python (Click/Typer)

`_format_commands_with_args` partitions commands using
the module-level `_command_groups` registry. Reads
`help_group` from Click context params to filter. The
`--help-<id>` flags are eager options; `help <group>`
is a Typer command.

## NO_COLOR Compliance

All paths (colored + no-color) use the same structural
renderer. When NO_COLOR is set or `--no-color` passed:
- Section headers rendered without ANSI
- Command/flag terms rendered plain
- Group titles rendered plain
- Section renaming (Options→FLAGS, Commands→COMMANDS)
  still applied

## ShowAliases

When `HelpConfig.ShowAliases = true`:
- Go: appends `(aliases: d, dp)` to command description
- TS: strips `|alias` from term, appends to description
- Python: appends to description (alias registry pending)

Aliases always work for dispatch regardless of setting.

## Command Term Format

All three languages render command terms matching Go/fang:

```
name <args> [--flags]    # command with args + options
name [command]           # command with subcommands
name                     # leaf command, no args
```
