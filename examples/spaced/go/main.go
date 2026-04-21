// spaced is a satirical SpaceX CLI historian and parity test vehicle for hop.top/kit/cli.
//
// It exercises the full kit CLI contract: global flags, format flag, help output,
// version, comma-list flags, short flags, subcommand trees, and structured output.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"hop.top/kit/examples/spaced/go/cmd"
	"hop.top/kit/go/console/alias"
	"hop.top/kit/go/console/cli"
	"hop.top/kit/go/core/xdg"
	"hop.top/kit/go/runtime/bus"
)

const disclaimer = `Not affiliated with, endorsed by, or in any way authorized by SpaceX,
Elon Musk, DOGE, NASA, the FAA, or the Starman mannequin currently past Mars.
We would, however, accept a sponsorship (https://github.com/sponsors/hop-top).
Cash, Starlink credits, or a ride on the next Crew Dragon all acceptable.`

func main() {
	root := cli.New(cli.Config{
		Name:    "spaced",
		Version: "0.1.0",
		Short:   "satirical SpaceX CLI historian — every launch, every RUD, every daemon",
		Accent:  "#FF5733",
		Help:    cli.HelpConfig{Disclaimer: disclaimer},
	})

	b := bus.New()

	// Log launch and daemon events.
	b.SubscribeAsync("launch.#", func(_ context.Context, e bus.Event) {
		fmt.Printf("  [bus] %s → %v\n", e.Topic, e.Payload)
	})
	b.SubscribeAsync("daemon.#", func(_ context.Context, e bus.Event) {
		fmt.Printf("  [bus] %s → %v\n", e.Topic, e.Payload)
	})

	// User-facing commands (default COMMANDS group).
	root.Cmd.AddCommand(cmd.MissionCmd(root))
	root.Cmd.AddCommand(cmd.LaunchCmd(root, b))
	root.Cmd.AddCommand(cmd.AbortCmd(root))
	root.Cmd.AddCommand(cmd.TelemetryCmd(root))
	root.Cmd.AddCommand(cmd.CountdownCmd(root))
	root.Cmd.AddCommand(cmd.FleetCmd(root))
	root.Cmd.AddCommand(cmd.StarshipCmd(root))
	root.Cmd.AddCommand(cmd.ElonCmd(root))
	root.Cmd.AddCommand(cmd.IpoCmd(root))
	root.Cmd.AddCommand(cmd.CompetitorCmd(root))
	root.Cmd.AddCommand(cmd.DaemonCmd(root, b))
	root.Cmd.AddCommand(cmd.ServeCmd())
	root.Cmd.AddCommand(cmd.SyncCmd())
	root.Cmd.AddCommand(cmd.PeerCmd())

	// Alias store at XDG config path.
	cfgDir := xdg.MustEnsure(xdg.ConfigDir("spaced"))
	store := alias.NewStore(filepath.Join(cfgDir, "aliases.yaml"))
	_ = store.Load()

	aliasCmd := root.AliasCmd(store)
	aliasCmd.GroupID = "management"
	root.Cmd.AddCommand(aliasCmd)

	// Load persisted aliases into the command tree.
	if err := root.LoadAliasStore(store); err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	// Management commands (hidden by default, shown with --help-all).
	authCmd := cmd.AuthStatusCmd(root)
	authCmd.GroupID = "management"
	root.Cmd.AddCommand(authCmd)

	configCmd := cmd.ConfigShowCmd()
	configCmd.GroupID = "management"
	root.Cmd.AddCommand(configCmd)

	toolspecCmd := cmd.ToolspecCmd()
	toolspecCmd.GroupID = "management"
	root.Cmd.AddCommand(toolspecCmd)

	complianceCmd := cmd.ComplianceCmd()
	complianceCmd.GroupID = "management"
	root.Cmd.AddCommand(complianceCmd)

	ctx := context.Background()
	if err := root.Execute(ctx); err != nil {
		_ = b.Close(ctx)
		os.Exit(1)
	}
	_ = b.Close(ctx)
}
