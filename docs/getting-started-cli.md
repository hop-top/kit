# Getting Started: Build Your First hop-top CLI

Three language factories; identical user-facing behavior.
Pick your runtime, get the same contract: version, help,
global flags, themed output.

## Prerequisites

| Language   | Install                          | Min version |
|------------|----------------------------------|-------------|
| Go         | `go get hop.top/kit@latest`      | Go 1.26     |
| TypeScript | `pnpm add @hop/kit-ts`           | Node 20     |
| Python     | `pip install hop-kit`            | Python 3.11 |

## Go

### Minimal working example

```
package main

import (
    "context"
    "os"

    "hop.top/kit/go/console/cli"
)

func main() {
    root := cli.New(cli.Config{
        Name:    "mytool",
        Version: "0.1.0",
        Short:   "Does useful things",
    })
    if err := root.Execute(context.Background()); err != nil {
        os.Exit(1)
    }
}
```

This gives you: `-v`/`--version`, `-h`/`--help`, `--quiet`,
`--no-color`, `--format`, `--no-hints`. No help or completion
subcommands.

### Add subcommands

```
root.Cmd.AddCommand(serveCmd(), listCmd())
```

Where each function returns a `*cobra.Command`.

### Built-in global flags

Registered automatically by `cli.New`:

| Flag          | Viper key  | Default   |
|---------------|------------|-----------|
| `--quiet`     | `quiet`    | `false`   |
| `--no-color`  | `no-color` | `false`   |
| `--format`    | `format`   | `"table"` |
| `--no-hints`  | `no-hints` | `false`   |

Read them from `root.Viper`:

```
if root.Viper.GetBool("quiet") {
    // suppress non-essential output
}
fmt := root.Viper.GetString("format")
output.Render(os.Stdout, fmt, data)
```

### Custom accent color

```
root := cli.New(cli.Config{
    Name:    "mytool",
    Version: "0.1.0",
    Short:   "Does useful things",
    Accent:  "#E040FB",
})
```

Sets the theme command color. Default palette: Neon (grass
green `#7ED957`, neon pink `#FF00FF`).

### Themed output

Access `root.Theme` for semantic styles:

```
fmt.Println(root.Theme.Title.Render("Section Header"))
fmt.Println(root.Theme.Subtle.Render("muted text"))
fmt.Println(root.Theme.Bold.Render("emphasis"))
```

Theme fields: `Accent`, `Secondary`, `Muted`, `Error`,
`Success`, `Title` (style), `Subtle` (style), `Bold` (style).

### Structured output

Use `kit/output` for table/json/yaml rendering:

```
type Item struct {
    ID   string `table:"ID"   json:"id"`
    Name string `table:"Name" json:"name"`
}

output.Render(os.Stdout, root.Viper.GetString("format"), items)
```

Table rendering driven by `table` struct tag. JSON and YAML
pass through to standard encoders.

### Next-step hints

Register hints via `root.Hints`:

```
root.Hints.Add("launch", output.Hint{
    Message: "Run `mytool status` to check progress",
})
```

Hints auto-suppressed when: `--no-hints`, `--format json/yaml`,
non-TTY (piped), or `--quiet`.

## TypeScript

### Minimal working example

```
import { createCLI } from '@hop/kit-sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/cli'

const program = createCLI({
    name: 'mytool',
    version: '0.1.0',
    description: 'Does useful things',
})

program.parse()
```

Returns a Commander `Command` with `-v`/`--version`,
`-h`/`--help`, `--format`, `--quiet`, `--no-color` pre-wired.

### Add subcommands

```
program
    .command('serve')
    .description('Start the server')
    .action(() => {
        // handler
    })

program
    .command('list')
    .description('List all items')
    .action(() => {
        const opts = program.opts()
        // opts.format => "table" | "json" | "yaml"
        // opts.quiet  => boolean
    })

program.parse()
```

### Read global flags

```
const opts = program.opts()
if (!opts.quiet) {
    console.log('Processing...')
}
```

### Custom flags

Add flags to specific subcommands via Commander API:

```
program
    .command('deploy')
    .option('--env <name>', 'Target environment', 'staging')
    .action((cmdOpts) => {
        // cmdOpts.env => "staging"
    })
```

## Python

### Minimal working example

```
from hop_kit.cli import create_app

app = create_app(
    name="mytool",
    version="0.1.0",
    help="Does useful things",
)
```

Returns a `typer.Typer` with `-v`/`--version` (eager),
completion flags disabled, `no_args_is_help=True`.

### Add subcommands

```
@app.command()
def serve(port: int = 8080):
    """Start the server."""
    print(f"Listening on :{port}")

@app.command()
def status():
    """Show current state."""
    print("OK")

if __name__ == "__main__":
    app()
```

### Custom flags

Typer uses function signatures as flag definitions:

```
@app.command()
def deploy(
    env: str = typer.Option("staging", help="Target env"),
    dry_run: bool = typer.Option(False, "--dry-run"),
):
    """Deploy to target environment."""
    if dry_run:
        typer.echo(f"Would deploy to {env}")
```

## What You Get

All three factories produce identical behavior:

```
$ mytool --version
mytool 0.1.0

$ mytool --help
Does useful things

USAGE
  mytool [command] [flags]

FLAGS
  -v, --version       Print mytool version and exit
      --format <fmt>  Output format (table, json, yaml)
      --quiet         Suppress non-essential output
      --no-color      Disable ANSI colour
      --no-hints      Suppress next-step hints
  -h, --help          Display help
```

Users and scripts cannot tell which runtime produced the output.

## Next Steps

- [CLI Parity Guide](cli-parity-guide.md) — full contract spec
- [SetFlag / TextFlag API](setflag-textflag-api.md) —
  multi-value flag types
- [Spaced Showcase](spaced-showcase.md) — example app
  demonstrating all kit packages
