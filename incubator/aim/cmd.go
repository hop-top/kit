package aim

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"hop.top/kit/go/console/output"
)

type modelRow struct {
	Provider    string `table:"PROVIDER"     json:"provider"`
	ID          string `table:"ID"           json:"id"`
	Name        string `table:"NAME"         json:"name"`
	Input       string `table:"INPUT"        json:"input"`
	Output      string `table:"OUTPUT"       json:"output"`
	ToolCall    bool   `table:"TOOLCALL"     json:"tool_call"`
	Reasoning   bool   `table:"REASONING"    json:"reasoning"`
	OpenWeights bool   `table:"OPEN-WEIGHTS" json:"open_weights"`
}

type providerRow struct {
	ID         string `table:"ID"     json:"id"`
	Name       string `table:"NAME"   json:"name"`
	ModelCount int    `table:"MODELS" json:"model_count"`
}

// Cmd returns the root "models" cobra.Command with all subcommands.
func Cmd() *cobra.Command {
	var reg *Registry

	root := &cobra.Command{
		Use:   "models",
		Short: "Browse AI model catalog",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			var err error
			reg, err = NewRegistry()
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return listModels(cmd, reg, args)
		},
	}

	root.PersistentFlags().String("format", "table",
		"Output format (table, json)")

	list := &cobra.Command{
		Use:   "list [query]",
		Short: "List models (default)",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listModels(cmd, reg, args)
		},
	}
	addFilterFlags(list)

	show := &cobra.Command{
		Use:   "show <provider> <model>",
		Short: "Show model details",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showModel(cmd, reg, args[0], args[1])
		},
	}

	providers := &cobra.Command{
		Use:   "providers",
		Short: "List providers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return listProviders(cmd, reg)
		},
	}

	refresh := &cobra.Command{
		Use:   "refresh",
		Short: "Force-refresh model cache",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := reg.Refresh(cmd.Context()); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "cache refreshed")
			return nil
		},
	}

	query := &cobra.Command{
		Use:   "query <string>",
		Short: "Query models (alias for list)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listModels(cmd, reg, args)
		},
	}

	root.AddCommand(list, show, providers, refresh, query)
	return root
}

func addFilterFlags(cmd *cobra.Command) {
	cmd.Flags().String("provider", "", "Filter by provider ID")
	cmd.Flags().String("family", "", "Filter by model family")
	cmd.Flags().String("input", "", "Filter by input modality")
	cmd.Flags().String("output", "", "Filter by output modality")
	cmd.Flags().Bool("tool-call", false, "Require tool-call support")
	cmd.Flags().Bool("reasoning", false, "Require reasoning support")
	cmd.Flags().Bool("open-weights", false, "Require open-weights")
}

func listModels(cmd *cobra.Command, reg *Registry, args []string) error {
	q := strings.Join(args, " ")

	var f Filter
	if q != "" {
		var err error
		if f, err = ParseQuery(q); err != nil {
			return err
		}
	}

	if v, _ := cmd.Flags().GetString("provider"); v != "" {
		f.Provider = v
	}
	if v, _ := cmd.Flags().GetString("family"); v != "" {
		f.Family = v
	}
	if v, _ := cmd.Flags().GetString("input"); v != "" {
		f.Input = v
	}
	if v, _ := cmd.Flags().GetString("output"); v != "" {
		f.Output = v
	}
	if cmd.Flags().Changed("tool-call") {
		b, _ := cmd.Flags().GetBool("tool-call")
		f.ToolCall = &b
	}
	if cmd.Flags().Changed("reasoning") {
		b, _ := cmd.Flags().GetBool("reasoning")
		f.Reasoning = &b
	}
	if cmd.Flags().Changed("open-weights") {
		b, _ := cmd.Flags().GetBool("open-weights")
		f.OpenWeights = &b
	}

	models := reg.Models(f)
	if len(models) == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "no models found")
		return nil
	}

	format := resolveFormat(cmd)
	if format == "json" {
		return output.Render(cmd.OutOrStdout(), output.JSON, models)
	}

	rows := make([]modelRow, len(models))
	for i, m := range models {
		rows[i] = modelRow{
			Provider:    m.Provider,
			ID:          m.ID,
			Name:        m.Name,
			Input:       strings.Join(m.Input, ","),
			Output:      strings.Join(m.Output, ","),
			ToolCall:    m.ToolCall,
			Reasoning:   m.Reasoning,
			OpenWeights: m.OpenWeights,
		}
	}
	return output.Render(cmd.OutOrStdout(), output.Table, rows)
}

func showModel(cmd *cobra.Command, reg *Registry, prov, id string) error {
	m, ok := reg.Get(prov, id)
	if !ok {
		return fmt.Errorf("model %s/%s not found", prov, id)
	}

	format := resolveFormat(cmd)
	if format == "json" {
		return output.Render(cmd.OutOrStdout(), output.JSON, m)
	}

	w := cmd.OutOrStdout()
	kv := func(k, v string) { fmt.Fprintf(w, "%-14s %s\n", k+":", v) }
	kv("Provider", m.Provider)
	kv("ID", m.ID)
	kv("Name", m.Name)
	kv("Family", m.Family)
	kv("Input", strings.Join(m.Input, ", "))
	kv("Output", strings.Join(m.Output, ", "))
	kv("ToolCall", boolStr(m.ToolCall))
	kv("Reasoning", boolStr(m.Reasoning))
	kv("OpenWeights", boolStr(m.OpenWeights))
	kv("Context", fmt.Sprint(m.Context))
	kv("MaxOutput", fmt.Sprint(m.MaxOutput))
	kv("CostInput", fmt.Sprintf("$%.2f/M tokens", m.CostInput))
	kv("CostOutput", fmt.Sprintf("$%.2f/M tokens", m.CostOutput))
	return nil
}

func listProviders(cmd *cobra.Command, reg *Registry) error {
	provs := reg.Providers()
	if len(provs) == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "no providers found")
		return nil
	}

	format := resolveFormat(cmd)
	if format == "json" {
		return output.Render(cmd.OutOrStdout(), output.JSON, provs)
	}

	rows := make([]providerRow, len(provs))
	for i, p := range provs {
		rows[i] = providerRow{
			ID: p.ID, Name: p.Name, ModelCount: len(p.Models),
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].ID < rows[j].ID
	})
	return output.Render(cmd.OutOrStdout(), output.Table, rows)
}

func resolveFormat(cmd *cobra.Command) string {
	if v, _ := cmd.Flags().GetString("format"); v != "" {
		return v
	}
	return "table"
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
