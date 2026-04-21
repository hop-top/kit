# TypeScript CLI API Reference

Package: `@hop/kit-sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/cli`

## CLIConfig

```ts
interface CLIConfig {
  name: string;        // binary name (e.g. "mytool")
  version: string;     // semver (e.g. "1.2.3")
  description: string; // one-line help description
}
```

## createCLI

```ts
function createCLI(cfg: CLIConfig): Command
```

Returns a Commander `Command` pre-configured to the
hop-top CLI contract:

- No help/completion subcommands; `-h`/`--help` flag only
- `-v, --version` prints `<name> <version>` and exits
- Global options: `--format`, `--quiet`, `--no-color`
- `showHelpAfterError` enabled

## Command Groups

### groups Config

Groups are declared in the CLIConfig:

```ts
interface CLIConfig {
  name: string;
  version: string;
  description: string;
  groups?: GroupConfig[];
}

interface GroupConfig {
  id: string;      // unique identifier (e.g. "management")
  title: string;   // display title (e.g. "MANAGEMENT COMMANDS")
  hidden: boolean; // true = excluded from default --help
}
```

Default groups when none specified:

| id | title | hidden |
|----|-------|--------|
| `commands` | COMMANDS | false |
| `management` | MANAGEMENT COMMANDS | true |

### setCommandGroup

```ts
function setCommandGroup(cmd: Command, groupId: string): void
```

Assigns a subcommand to a named group. Commands without
a group assignment default to `commands`.

```ts
const program = createCLI({ name: 'mytool', ... });
const configCmd = program.command('config')
  .description('Manage configuration');
setCommandGroup(configCmd, 'management');
```

### --help-all

Registered as a root-level boolean option. When passed,
the help formatter includes commands from hidden groups.

```
$ mytool --help          # shows COMMANDS only
$ mytool --help-all      # shows COMMANDS + MANAGEMENT
```

## Output

Package: `@hop/kit-sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/output`

Provides `renderTable`, `renderJSON`, `renderYAML` for
consistent output formatting across tools.
