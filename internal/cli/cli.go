// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package cli defines the spillcheck command tree (PLAN 14).
package cli

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"github.com/Hungleminh/agentspillcheck/internal/core/config"
	"github.com/Hungleminh/agentspillcheck/internal/core/i18n"
	spillmcp "github.com/Hungleminh/agentspillcheck/internal/mcp"
	"github.com/Hungleminh/agentspillcheck/internal/ui"
)

// NewRootCommand builds the root command writing to out.
func NewRootCommand(version string, out io.Writer) *cobra.Command {
	var locale string
	var portBase int
	trFor := func() *i18n.Translator {
		loc := locale
		if loc == "" {
			loc = os.Getenv("SPILLCHECK_LOCALE")
		}
		if loc == "" {
			loc = os.Getenv("LANG")
		}
		tr, err := i18n.New(loc)
		if err != nil {
			tr, _ = i18n.New("en")
		}
		return tr
	}
	root := &cobra.Command{
		Use:   "spillcheck",
		Short: "AgentSpillCheck",
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.SetOut(out)
	defaultHelp := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		loc := locale
		if f := cmd.Flags().Lookup("locale"); f != nil && f.Value.String() != "" {
			loc = f.Value.String()
		}
		tr, err := i18n.New(loc)
		if err != nil {
			tr, _ = i18n.New("en")
		}
		root.Short = tr.T("root.short")
		root.Long = tr.T("root.long")
		defaultHelp(cmd, args)
	})
	// Localised help needs the flag value, so refresh usage strings on run.
	root.PersistentPreRun = func(_ *cobra.Command, _ []string) {
		tr := trFor()
		root.Short = tr.T("root.short")
		root.Long = tr.T("root.long")
	}
	root.PersistentFlags().StringVar(&locale, "locale", "", "vi|en")
	root.PersistentFlags().IntVar(&portBase, "port-base", 0, "7700")

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "version",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "spillcheck %s (%s)\n", trFor().T("version.label"), version)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", version)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "doctor",
		Short: "doctor",
		Run: func(cmd *cobra.Command, _ []string) {
			tr := trFor()
			cfg := loadCfg(locale, portBase)
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "%s\n", tr.T("doctor.title"))
			_, _ = fmt.Fprintf(w, "%s: %s %s\n", tr.T("doctor.go"), runtime.Version(), runtime.GOOS+"/"+runtime.GOARCH)
			status := tr.T("doctor.port_free")
			ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", cfg.PortBase))
			if err != nil {
				status = tr.T("doctor.port_busy") + " (" + err.Error() + ")"
			} else {
				_ = ln.Close()
			}
			_, _ = fmt.Fprintf(w, "127.0.0.1:%d %s\n", cfg.PortBase, status)
			dataStatus := tr.T("doctor.not_writable")
			if err := os.MkdirAll(cfg.DataDir, 0o700); err == nil {
				dataStatus = tr.T("doctor.writable")
			}
			_, _ = fmt.Fprintf(w, "data dir: %s %s\n", cfg.DataDir, dataStatus)
		},
	})

	mcpCmd := &cobra.Command{Use: "mcp", Short: "mcp"}
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "serve",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := loadCfg(locale, portBase)
			backend := spillmcp.NewBackend(cfg.Locale, 42)
			stdio, _ := cmd.Flags().GetBool("stdio")
			if stdio {
				return spillmcp.RunStdio(cmd.Context(), backend, version)
			}
			addr := fmt.Sprintf("127.0.0.1:%d", cfg.PortBase)
			srv := &http.Server{Addr: addr, Handler: spillmcp.HTTPHandler(backend, version, ""), ReadHeaderTimeout: 5 * time.Second}
			return srv.ListenAndServe()
		},
	}
	serveCmd.Flags().Bool("stdio", false, "stdio")
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "config",
		Run: func(cmd *cobra.Command, _ []string) {
			client, _ := cmd.Flags().GetString("client")
			_ = client // same stdio snippet shape for all supported clients in v0.1
			_, _ = fmt.Fprint(cmd.OutOrStdout(), "{\n  \"mcpServers\": {\n    \"spillcheck\": {\n      \"command\": \"spillcheck\",\n      \"args\": [\"mcp\", \"serve\", \"--stdio\"]\n    }\n  }\n}\n")
		},
	}
	configCmd.Flags().String("client", "claude-desktop", "claude-desktop|cursor|claude-code")
	mcpCmd.AddCommand(serveCmd, configCmd)
	root.AddCommand(mcpCmd)

	upCmd := &cobra.Command{
		Use:   "up",
		Short: "up",
		RunE: func(cmd *cobra.Command, _ []string) error {
			tr := trFor()
			cfg := loadCfg(locale, portBase)
			addr := fmt.Sprintf("127.0.0.1:%d", cfg.PortBase)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", tr.Tf("up.starting", cfg.PortBase))
			srv := &http.Server{Addr: addr, Handler: ui.NewHandler(cfg), ReadHeaderTimeout: 5 * time.Second}
			return srv.ListenAndServe()
		},
	}
	root.AddCommand(upCmd)
	return root
}

func loadCfg(locale string, portBase int) config.Config {
	cfg := config.Load(os.Getenv, os.Getenv("LANG"))
	if locale != "" {
		cfg.Locale = i18n.Detect(locale)
	}
	if portBase > 0 {
		cfg.PortBase = portBase
	}
	return cfg
}
