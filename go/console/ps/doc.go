// Package ps provides a cross-tool process status convention for hop.top
// CLI tools.
//
// Every hop.top tool that manages asynchronous or long-running work
// exposes a `<tool> ps` subcommand. This package supplies the shared
// types, rendering, and command wiring so adopters get a consistent UX
// with minimal code.
//
// # Convention
//
// The ps subcommand lists active work items — tasks, jobs, sessions,
// evaluations — in a standard table format:
//
//	ID       STATUS    WORKER    SCOPE          DURATION  PROGRESS
//	abc-123  running   agent-1   build/deploy   5m        3/10 (30%)
//
// Standard columns: ID, Status (colored), Worker, Scope (truncated 40ch),
// Duration (since started), Progress (done/total with percentage).
// Optional columns Worktree and Track appear when any entry populates them.
//
// # Flags
//
//   - --json        output as JSON array
//   - --all / -a    include completed ("done") entries
//   - --quiet / -q  print IDs only, one per line
//   - --watch / -w  re-poll at interval (default 5s), clear + redraw
//
// # Provider interface
//
// Tools implement the [Provider] interface:
//
//	type Provider interface {
//	    List(ctx context.Context) ([]Entry, error)
//	}
//
// Then wire it in three lines:
//
//	root.AddCommand(ps.Command("mytool", myProvider, v))
//
// # Adopters
//
// The following hop.top tools implement ps:
//
//   - tlc   — task lifecycle: running agents, pending flows
//   - pod   — pod orchestrator: active containers, build jobs
//   - eva   — evaluator: running evals, scoring passes
//   - asm   — assembler: active assembly pipelines
package ps
