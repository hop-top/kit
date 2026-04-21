package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"hop.top/kit/go/core/util"
	"hop.top/kit/go/runtime/peer"
	"hop.top/kit/go/storage/sqlstore"
)

// PeerConfig configures peer mesh management.
type PeerConfig struct {
	// Discovery overrides the default discoverer. Nil = static (no-op).
	Discovery peer.Discoverer
	// Service is the mDNS service name (default "_kit._tcp").
	Service string
	// DataDir overrides the peer data directory.
	// Default: $XDG_DATA_HOME/kit/peers/
	DataDir string
}

// WithPeers enables peer mesh management on the CLI root.
func WithPeers(cfg PeerConfig) func(*Root) {
	return func(r *Root) {
		r.peerCfg = &cfg
	}
}

// initPeers creates the registry, trust manager, mesh and attaches commands.
func (r *Root) initPeers() error {
	cfg := r.peerCfg
	if cfg == nil {
		return nil
	}

	if r.Identity == nil {
		return fmt.Errorf("peer: WithPeers requires WithIdentity")
	}

	dataDir := cfg.DataDir
	if dataDir == "" {
		xdg := os.Getenv("XDG_DATA_HOME")
		if xdg == "" {
			home, _ := os.UserHomeDir()
			xdg = filepath.Join(home, ".local", "share")
		}
		dataDir = filepath.Join(xdg, "kit", "peers")
	}

	store, err := sqlstore.Open(filepath.Join(dataDir, "peers.db"), sqlstore.Options{})
	if err != nil {
		return fmt.Errorf("peer store: %w", err)
	}

	registry := peer.NewRegistry(store)
	disc := cfg.Discovery
	if disc == nil {
		disc = &peer.StaticDiscoverer{}
	}

	tm := peer.NewTrustManager(registry, r.Identity)

	// Build self PeerInfo from identity if available.
	var self peer.PeerInfo
	if r.Identity != nil {
		pubPEM, _ := r.Identity.MarshalPublicKey()
		self = peer.PeerInfo{
			ID:        r.Identity.PublicKeyID(),
			PublicKey: pubPEM,
		}
	}

	r.PeerRegistry = registry
	r.PeerTrust = tm
	r.Mesh = peer.NewMesh(self, tm, disc)

	r.Cmd.AddCommand(peerCmd(r))
	return nil
}

func peerCmd(r *Root) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peer",
		Short: "Manage mesh peers",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		peerListCmd(r),
		peerTrustCmd(r),
		peerBlockCmd(r),
		peerRevokeCmd(r),
	)
	return cmd
}

func peerListCmd(r *Root) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show known peers",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			records, err := r.PeerRegistry.List()
			if err != nil {
				return err
			}
			if len(records) == 0 {
				fmt.Fprintln(r.Streams.Human, "No peers found.")
				return nil
			}
			w := tabwriter.NewWriter(r.Streams.Data, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tADDRS\tTRUST\tLAST SEEN")
			for _, rec := range records {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					rec.ID,
					rec.Name,
					strings.Join(rec.Addrs, ","),
					trustLabel(rec.Trust),
					util.RelativeTime(rec.LastSeen),
				)
			}
			return w.Flush()
		},
	}
}

func peerTrustCmd(r *Root) *cobra.Command {
	return &cobra.Command{
		Use:   "trust <id>",
		Short: "Explicitly trust a peer",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return r.PeerTrust.Trust(args[0])
		},
	}
}

func peerBlockCmd(r *Root) *cobra.Command {
	return &cobra.Command{
		Use:   "block <id>",
		Short: "Block a peer",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return r.PeerTrust.Block(args[0])
		},
	}
}

func peerRevokeCmd(r *Root) *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <id>",
		Short: "Revoke trust from a peer",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return r.PeerTrust.Revoke(args[0])
		},
	}
}

func trustLabel(t peer.TrustLevel) string {
	switch t {
	case peer.Trusted:
		return "trusted"
	case peer.Blocked:
		return "blocked"
	case peer.PendingTOFU:
		return "pending"
	default:
		return "unknown"
	}
}
