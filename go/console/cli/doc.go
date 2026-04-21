// Package cli provides a cobra+fang+viper root command factory for hop-top CLIs.
//
// # Factory
//
// [New] builds a fully-wired [Root] from a [Config] and optional functional
// options. Config fields control name, version, accent color, disabled
// flags, global flags, and help layout.
//
//	root := cli.New(cli.Config{
//	    Name:    "mytool",
//	    Version: "1.2.3",
//	    Short:   "A local-first CLI",
//	    Accent:  "#FF5F87",
//	})
//
// # Config
//
// Config drives what New registers:
//   - Name, Version, Short: identity and --version output
//   - Accent: lipgloss color used in Theme
//   - Disable: suppress built-in flags (Format, Quiet, NoColor, Hints)
//   - Globals: extra persistent flags registered on the root cobra.Command
//   - Help: section order, groups, disclaimer, alias display
//
// # Root
//
// Root is the returned value from New. Key fields:
//   - Cmd: the [cobra.Command] root
//   - Viper: bound [viper.Viper] instance
//   - Config: the resolved Config
//   - Theme: parity.Theme built from Accent
//   - Hints: output.Hints (respects --no-hints)
//   - Streams: output.Streams (stdout/stderr wrappers)
//   - Auth: JWT token from identity store (when WithIdentity used)
//   - Identity: *identity.Keypair (when WithIdentity used)
//   - Mesh: *peer.Mesh (when WithPeers used)
//
// # Options
//
//   - [WithAPI]: registers API client config flags and populates Root.API
//   - [WithIdentity]: loads or generates an Ed25519 keypair, signs a JWT
//   - [WithPeers]: starts mDNS discovery and trust mesh
//
// # Execution
//
// Execute wires context cancellation (SIGINT/SIGTERM), applies persistent
// pre-run logic (identity load, mesh start), and delegates to cobra:
//
//	root.Cmd.AddCommand(serveCmd(), syncCmd())
//	if err := root.Execute(context.Background()); err != nil {
//	    os.Exit(1)
//	}
package cli
