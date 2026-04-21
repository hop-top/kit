# Go CLI API Reference

Package: `hop.top/kit/cli`

## Config

```go
type Config struct {
    Name    string // binary name (e.g. "mytool")
    Version string // semver (e.g. "1.2.3")
    Short   string // one-line description
    Accent  string // optional hex colour (e.g. "#E040FB")
}
```

## Root

```go
type Root struct {
    Cmd    *cobra.Command
    Viper  *viper.Viper
    Config Config
    Theme  Theme
    Hints  *output.HintSet
}
```

### New

```go
func New(cfg Config) *Root
```

Returns a Root pre-configured to the hop-top CLI contract:
- No help/completion subcommands
- Persistent flags: `--quiet`, `--no-color`, `--format`
- Version handled by fang (`-v`/`--version`)
- Styled help via fang colour scheme

### Execute

```go
func (r *Root) Execute(ctx context.Context) error
```

Runs the root command through fang. Handles version output,
styled help, error rendering, and man page generation.

## Command Groups

### GroupConfig

```go
type GroupConfig struct {
    ID     string // unique identifier (e.g. "management")
    Title  string // display title (e.g. "MANAGEMENT COMMANDS")
    Hidden bool   // true = excluded from default --help
}
```

### HelpConfig.Groups

Groups are declared on the HelpConfig and passed to the
root command setup:

```go
type HelpConfig struct {
    Groups []GroupConfig
}
```

Default groups when none specified:

| ID | Title | Hidden |
|----|-------|--------|
| `commands` | COMMANDS | false |
| `management` | MANAGEMENT COMMANDS | true |

### Assigning a Command to a Group

Use cobra's built-in `GroupID` field:

```go
cmd := &cobra.Command{
    Use:     "config",
    Short:   "Manage configuration",
    GroupID: "management",
}
root.Cmd.AddCommand(cmd)
```

Commands without a `GroupID` default to the `commands`
group.

### --help-all Flag

Registered as a persistent boolean flag on the root
command. When set, the help template includes commands
from all groups (including hidden ones).

```
$ mytool --help          # shows COMMANDS only
$ mytool --help-all      # shows COMMANDS + MANAGEMENT
```

## Theme

```go
type Theme struct {
    Accent    lipgloss.TerminalColor
    Dim       lipgloss.TerminalColor
    Success   lipgloss.TerminalColor
    Warning   lipgloss.TerminalColor
    Error     lipgloss.TerminalColor
    Command   lipgloss.TerminalColor
    Flag      lipgloss.TerminalColor
}
```

Built from CharmTone palette plus optional `Config.Accent`.
