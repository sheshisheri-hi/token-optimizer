package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sheshisheri-hi/token-optimizer/internal/core/capabilities"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/config"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/registry"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/manifest"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/measurement"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/tally"
	"github.com/sheshisheri-hi/token-optimizer/internal/hook"
	"github.com/sheshisheri-hi/token-optimizer/internal/app"
	"github.com/sheshisheri-hi/token-optimizer/internal/runtime"
	"github.com/sheshisheri-hi/token-optimizer/internal/version"
	"github.com/spf13/cobra"
)

var (
	Version = version.Version
	target  = "copilot"
)

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "slim",
		Short: "Copilot Slim — token optimization for GitHub Copilot",
		Long:  "Org toolkit to reduce Copilot token usage. See docs/PLAN.md and docs/commands.md.",
	}
	root.PersistentFlags().StringVar(&target, "target", "copilot", "IDE target: copilot, cursor, claude")

	root.AddCommand(cmdSetup())
	root.AddCommand(cmdOff())
	root.AddCommand(cmdStatus())
	root.AddCommand(cmdDoctor())
	root.AddCommand(cmdConfig())
	root.AddCommand(cmdReport())
	root.AddCommand(cmdLogs())
	root.AddCommand(cmdInternal())
	root.AddCommand(cmdHelpExtra())
	return root
}

func rt() (runtime.Runtime, error) {
	return app.RuntimeFor(target)
}

func printTallyFooter(rt runtime.Runtime) {
	reg, err := registry.Load()
	if err != nil {
		return
	}
	dataHome, err := rt.DataHome()
	if err != nil {
		return
	}
	man, _ := manifest.Load(dataHome)
	cfg, _ := config.Load(dataHome)
	fmt.Println()
	fmt.Printf("Copilot Slim %s (%s)\n", Version, rt.DisplayName())
	fmt.Println(tally.Compute(reg, man, cfg, rt.ID()).Footer())
	fmt.Println("Run: slim doctor | slim help paths")
}

func cmdSetup() *cobra.Command {
	var all bool
	var dryRun bool
	var modules []string
	c := &cobra.Command{
		Use:   "setup",
		Short: "Install or refresh Slim modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := rt()
			if err != nil {
				return err
			}
			res, err := rt.Setup(context.Background(), runtime.SetupOpts{
				Modules: modules,
				All:     all,
				DryRun:  dryRun,
			})
			if err != nil {
				if err == runtime.ErrNotImplemented {
					return fmt.Errorf("%s is not available yet — see docs/PLAN.md", target)
				}
				return err
			}
			for _, a := range res.Actions {
				fmt.Println(a)
			}
			if !dryRun {
				printTallyFooter(rt)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&all, "all", false, "install default modules")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "preview changes")
	c.Flags().StringSliceVar(&modules, "module", nil, "module to install (repeatable)")
	return c
}

func cmdOff() *cobra.Command {
	var all bool
	var modules []string
	c := &cobra.Command{
		Use:   "off",
		Short: "Disable modules or full opt-out",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := rt()
			if err != nil {
				return err
			}
			res, err := rt.Off(context.Background(), runtime.OffOpts{Modules: modules, All: all})
			if err != nil {
				return err
			}
			for _, a := range res.Actions {
				fmt.Println(a)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&all, "all", false, "disable all installed modules")
	c.Flags().StringSliceVar(&modules, "module", nil, "module to disable")
	return c
}

func cmdStatus() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show installed modules and feature counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := rt()
			if err != nil {
				return err
			}
			rep, err := rt.Status(context.Background())
			if err != nil {
				return err
			}
			printReport(rep)
			return nil
		},
	}
}

func cmdDoctor() *cobra.Command {
	var asJSON bool
	var probe bool
	c := &cobra.Command{
		Use:   "doctor",
		Short: "Health check per module and feature",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := rt()
			if err != nil {
				return err
			}
			if probe {
				dataHome, err := rt.DataHome()
				if err == nil {
					_ = capabilities.Seed(dataHome)
				}
			}
			rep, err := rt.Doctor(context.Background(), runtime.DoctorOpts{JSON: asJSON, Probe: probe})
			if err != nil {
				return err
			}
			if asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(rep)
			}
			fmt.Printf("Copilot Slim %s — Doctor\n", Version)
			fmt.Println(strings.Repeat("─", 40))
			fmt.Println(rep.Tally)
			fmt.Println("\nMODULES")
			for _, m := range rep.Modules {
				fmt.Printf("  %-14s %-8s %s\n", m.Name, m.Status, m.Detail)
			}
			fmt.Println("\nFEATURES")
			for _, f := range rep.Features {
				fmt.Printf("  %-20s %-8s %s\n", f.Name, f.Status, f.Detail)
			}
			for _, w := range rep.Warnings {
				fmt.Printf("\n! %s\n", w)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "JSON output")
	c.Flags().BoolVar(&probe, "probe", false, "re-seed capability probes")
	return c
}

func cmdReport() *cobra.Command {
	var days int
	c := &cobra.Command{
		Use:   "report",
		Short: "Usage and savings summary from local measurement DB",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := rt()
			if err != nil {
				return err
			}
			dataHome, err := rt.DataHome()
			if err != nil {
				return err
			}
			sum, err := measurement.Report(dataHome, days)
			if err != nil {
				return err
			}
			fmt.Println(measurement.FormatSummary(sum))
			fmt.Printf("DB: %s\n", measurement.DBPath(dataHome))
			return nil
		},
	}
	c.Flags().IntVar(&days, "days", 7, "rollup window in days")
	return c
}

func cmdConfig() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "View or change feature toggles",
	}
	var list bool
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List feature toggles",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := rt()
			if err != nil {
				return err
			}
			dataHome, err := rt.DataHome()
			if err != nil {
				return err
			}
			cfg, err := config.Load(dataHome)
			if err != nil {
				return err
			}
			for k, v := range cfg.Features {
				fmt.Printf("  %s: %v\n", k, v)
			}
			return nil
		},
	}
	setCmd := &cobra.Command{
		Use:   "set [feature] [on|off]",
		Short: "Enable or disable a feature",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := rt()
			if err != nil {
				return err
			}
			dataHome, err := rt.DataHome()
			if err != nil {
				return err
			}
			cfg, err := config.Load(dataHome)
			if err != nil {
				return err
			}
			on := strings.EqualFold(args[1], "on") || args[1] == "true"
			cfg.SetFeature(args[0], on)
			return config.Save(dataHome, cfg)
		},
	}
	_ = list
	configCmd.AddCommand(listCmd, setCmd)
	return configCmd
}

func cmdLogs() *cobra.Command {
	return &cobra.Command{
		Use:   "logs",
		Short: "Show log paths and recent hook activity",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := rt()
			if err != nil {
				return err
			}
			dataHome, err := rt.DataHome()
			if err != nil {
				return err
			}
			fmt.Println("Log paths:")
			fmt.Printf("  manifest: %s/manifest.json\n", dataHome)
			fmt.Printf("  config:   %s/config.yaml\n", dataHome)
			fmt.Printf("  hook.log: %s/logs/hook.log\n", dataHome)
			fmt.Printf("  sessions: %s/sessions/\n", dataHome)
			logPath := fmt.Sprintf("%s/logs/hook.log", dataHome)
			data, err := os.ReadFile(logPath)
			if err == nil && len(data) > 0 {
				lines := strings.Split(strings.TrimSpace(string(data)), "\n")
				start := 0
				if len(lines) > 20 {
					start = len(lines) - 20
				}
				fmt.Println("\nLast hook log lines:")
				for _, l := range lines[start:] {
					fmt.Println(" ", l)
				}
			}
			return nil
		},
	}
}

func cmdInternal() *cobra.Command {
	hookCmd := &cobra.Command{
		Use:    "internal",
		Hidden: true,
	}
	hookCmd.AddCommand(&cobra.Command{
		Use:   "hook [event]",
		Short: "Copilot hook entry (fail-open)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			rt, err := rt()
			if err != nil {
				return
			}
			dataHome, err := rt.DataHome()
			if err != nil {
				return
			}
			out, _ := hook.Handle(target, args[0], dataHome)
			if len(out) > 0 {
				_, _ = os.Stdout.Write(out)
			}
		},
	})
	return hookCmd
}

func cmdHelpExtra() *cobra.Command {
	return &cobra.Command{
		Use:   "paths",
		Short: "Print data paths and command map",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(`Commands:
  slim setup [--all] [--module NAME] [--dry-run]
  slim status | doctor [--json] [--probe] | logs | report [--days N]
  slim config list | config set FEATURE on|off
  slim off [--all] [--module NAME]

After setup (copilot):
  ~/.copilot/slim/           data home
  ~/.copilot/slim/bin/slim   installed CLI + hook bridge
  ~/.copilot/hooks/copilot-slim.json

See docs/commands.md and docs/PLAN.md`)
		},
	}
}

func printReport(rep *runtime.StatusReport) {
	fmt.Printf("Target:   %s\n", rep.Target)
	fmt.Printf("Data home: %s\n\n", rep.DataHome)
	fmt.Println(rep.Tally)
}
