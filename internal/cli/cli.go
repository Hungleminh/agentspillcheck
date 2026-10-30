// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package cli defines the spillcheck command tree (PLAN 14).
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Hungleminh/agentspillcheck/internal/analyze"
	"github.com/Hungleminh/agentspillcheck/internal/api"
	"github.com/Hungleminh/agentspillcheck/internal/bait"
	"github.com/Hungleminh/agentspillcheck/internal/core/config"
	"github.com/Hungleminh/agentspillcheck/internal/core/i18n"
	coresession "github.com/Hungleminh/agentspillcheck/internal/core/session"
	"github.com/Hungleminh/agentspillcheck/internal/demo"
	spillmcp "github.com/Hungleminh/agentspillcheck/internal/mcp"
	"github.com/Hungleminh/agentspillcheck/internal/report"
	"github.com/Hungleminh/agentspillcheck/internal/store"
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

	sessionCmd := &cobra.Command{Use: "session", Short: "session"}
	newSessionCmd := &cobra.Command{
		Use:   "new",
		Short: "new",
		Run: func(cmd *cobra.Command, _ []string) {
			cfg := loadCfg(locale, portBase)
			mgr := coresession.NewManager(nil)
			tier, _ := cmd.Flags().GetString("tier")
			mode, _ := cmd.Flags().GetString("mode")
			label, _ := cmd.Flags().GetString("agent-label")
			info, err := mgr.Create(cfg.Locale, tier, mode, label)
			if err != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), err.Error())
				return
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "session_id: %s\npersona: %s\n", info.SessionID, info.PersonaSummary)
		},
	}
	for _, name := range []string{"list", "show", "expire"} {
		sub := name
		sessionCmd.AddCommand(&cobra.Command{
			Use:   sub + " [session_id]",
			Short: sub,
			Run: func(cmd *cobra.Command, args []string) {
				cfg := loadCfg(locale, portBase)
				base := fmt.Sprintf("http://127.0.0.1:%d", cfg.PortBase)
				path := "/api/v1/sessions"
				method := http.MethodGet
				if sub != "list" {
					if len(args) != 1 {
						_, _ = fmt.Fprintln(cmd.OutOrStdout(), "session_id required")
						return
					}
					path += "/" + args[0]
					if sub == "expire" {
						path += "/expire"
						method = http.MethodPost
					}
				}
				body, err := apiCall(cmd.Context(), method, base+path)
				if err != nil {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), err.Error())
					return
				}
				_, _ = fmt.Fprint(cmd.OutOrStdout(), body)
			},
		})
	}
	sessionCmd.AddCommand(newSessionCmd)
	newSessionCmd.Flags().String("tier", "standard", "quick|standard|deep")
	newSessionCmd.Flags().String("mode", "mcp", "mcp|browser|connector|manual")
	newSessionCmd.Flags().String("agent-label", "", "")
	root.AddCommand(sessionCmd)

	baitCmd := &cobra.Command{Use: "bait", Short: "bait"}
	previewCmd := &cobra.Command{
		Use:   "preview",
		Short: "preview",
		Run: func(cmd *cobra.Command, _ []string) {
			catalog := bait.Catalog()
			names := make([]string, 0, len(catalog))
			for name := range catalog {
				names = append(names, string(name))
			}
			sort.Strings(names)
			for _, name := range names {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s (%s)\n", name, catalog[bait.Subclass(name)])
			}
		},
	}
	previewCmd.Flags().String("tier", "standard", "quick|standard|deep")
	baitCmd.AddCommand(previewCmd)
	root.AddCommand(baitCmd)

	demoCmd := &cobra.Command{
		Use:   "demo",
		Short: "demo",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := loadCfg(locale, portBase)
			backend := spillmcp.NewBackend(cfg.Locale, 42)
			profile, _ := cmd.Flags().GetString("profile")
			if err := demo.Run(cmd.Context(), backend, demo.Profile(profile)); err != nil {
				return err
			}
			rep := analyze.Enrich(report.Build(backend.SessionID(), backend.Baits(), backend.Events()), backend.Events())
			outDir, _ := cmd.Flags().GetString("out")
			if outDir != "" {
				if err := writeReports(outDir, rep, cfg.Locale); err != nil {
					return err
				}
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "session: %s\ngrade: %s\nscore: %d\nfindings: %d\n",
				rep.SessionID, rep.Grade, rep.Score, len(rep.Findings))
			return nil
		},
	}
	demoCmd.Flags().String("profile", "leaky", "clean|leaky")
	demoCmd.Flags().String("out", "", "output directory for report files")
	root.AddCommand(demoCmd)

	reportCmd := &cobra.Command{
		Use:   "report [session_id]",
		Short: "report",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadCfg(locale, portBase)
			format, _ := cmd.Flags().GetString("format")
			explainWith, _ := cmd.Flags().GetString("explain-with")
			base, _ := cmd.Flags().GetString("base-url")
			if base == "" {
				base = fmt.Sprintf("http://127.0.0.1:%d", cfg.PortBase)
			}
			body, err := fetchReport(cmd.Context(), base, args[0], format)
			if err != nil {
				return err
			}
			if explainWith != "" {
				// Disclosure (ADR-0005): only the anonymised report
				// JSON above leaves this machine, sent to the named
				// provider with the user's key from the environment.
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "explain-with %s: sending only the anonymised report JSON (masked bait values) to the provider with your API key from SPILLCHECK_LLM_KEY/OPENAI_API_KEY\n", explainWith)
				text, err := explainReport(cmd.Context(), body)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "explanation: %s\n", text)
				return nil
			}
			_, _ = fmt.Fprint(cmd.OutOrStdout(), body)
			return nil
		},
	}
	purgeCmd := &cobra.Command{
		Use:   "purge",
		Short: "purge",
		Run: func(cmd *cobra.Command, _ []string) {
			cfg := loadCfg(locale, portBase)
			tr := trFor()
			older, _ := cmd.Flags().GetString("older-than")
			dur, err := parseOlderThan(older)
			if err != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), err.Error())
				return
			}
			_ = purgeOld(filepath.Join(cfg.DataDir, "sessions"), time.Now().Add(-dur))
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", tr.Tf("cli.purged", older))
		},
	}
	purgeCmd.Flags().String("older-than", "168h", "e.g. 168h or 7d")
	root.AddCommand(purgeCmd)

	root.AddCommand(&cobra.Command{
		Use:   "down",
		Short: "down",
		Run: func(cmd *cobra.Command, _ []string) {
			cfg := loadCfg(locale, portBase)
			tr := trFor()
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", stopInstance(cfg.DataDir, tr))
		},
	})

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "update",
		Run: func(cmd *cobra.Command, _ []string) {
			tr := trFor()
			latest, err := latestRelease(cmd.Context())
			if err != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), err.Error())
				return
			}
			if latest == version || latest == "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", tr.Tf("cli.up_to_date", version))
				return
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", tr.Tf("cli.update_available", latest, version))
		},
	}
	root.AddCommand(updateCmd)

	runCmd := &cobra.Command{
		Use:   "run [session_id]",
		Short: "run",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := loadCfg(locale, portBase)
			base := fmt.Sprintf("http://127.0.0.1:%d", cfg.PortBase)
			connectorName, _ := cmd.Flags().GetString("connector")
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "run session %s via connector %s against %s (manual/httpagent/claudecode execute through the API task loop)\n", args[0], connectorName, base)
		},
	}
	runCmd.Flags().String("connector", "manual", "manual|httpagent|claudecode")
	root.AddCommand(runCmd)

	reportCmd.Flags().String("format", "json", "html|json|md|all")
	reportCmd.Flags().String("explain-with", "", "provider name (off by default)")
	reportCmd.Flags().String("base-url", "", "running instance base URL")
	root.AddCommand(reportCmd)

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
			if err := os.MkdirAll(cfg.DataDir, 0o700); err == nil {
				_ = os.WriteFile(filepath.Join(cfg.DataDir, "spillcheck.pid"), []byte(fmt.Sprintf("%d", os.Getpid())), 0o600)
			}
			st, err := store.Open(filepath.Join(cfg.DataDir, "data.db"))
			if err != nil {
				return err
			}
			defer func() { _ = st.Close() }()
			mgr := coresession.NewManager(st)
			rootMux := http.NewServeMux()
			rootMux.Handle("/api/", api.NewHandler(mgr, version, ""))
			rootMux.Handle("/mcp", spillmcp.HTTPHandler(spillmcp.NewBackend(cfg.Locale, 42), version, ""))
			rootMux.Handle("/", ui.NewHandler(cfg))
			srv := &http.Server{Addr: addr, Handler: rootMux, ReadHeaderTimeout: 5 * time.Second}
			return srv.ListenAndServe()
		},
	}
	root.AddCommand(upCmd)
	return root
}

func fetchReport(ctx context.Context, base, sessionID, format string) (string, error) {
	if format == "all" {
		format = "json"
	}
	url := fmt.Sprintf("%s/api/v1/sessions/%s/report.%s", base, sessionID, format)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("report: instance returned %s", resp.Status)
	}
	return string(raw), nil
}

func explainReport(ctx context.Context, anonymisedJSON string) (string, error) {
	var rep report.Report
	if err := json.Unmarshal([]byte(anonymisedJSON), &rep); err != nil {
		return "", err
	}
	key := os.Getenv("SPILLCHECK_LLM_KEY")
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	client := &report.OpenAIClient{APIKey: key}
	return report.Explain(ctx, rep, client)
}

func writeReports(dir string, rep report.Report, locale string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "report.json"), raw, 0o600); err != nil {
		return err
	}
	html, err := report.RenderHTML(rep, locale)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "report.html"), []byte(html), 0o600); err != nil {
		return err
	}
	md, err := report.RenderMarkdown(rep, locale)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "report.md"), []byte(md), 0o600)
}

func parseOlderThan(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

func purgeOld(root string, cutoff time.Time) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		path := filepath.Join(root, e.Name())
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			_ = purgeOld(path, cutoff)
			_ = os.Remove(path) // only succeeds when empty
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(path)
		}
	}
	return nil
}

func stopInstance(dataDir string, tr *i18n.Translator) string {
	raw, err := os.ReadFile(filepath.Join(dataDir, "spillcheck.pid"))
	if err != nil {
		return tr.T("cli.not_running")
	}
	var pid int
	if _, err := fmt.Sscanf(string(raw), "%d", &pid); err != nil || pid <= 0 {
		return tr.T("cli.not_running")
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return tr.T("cli.not_running")
	}
	if err := proc.Signal(os.Interrupt); err != nil {
		return tr.T("cli.not_running")
	}
	_ = os.Remove(filepath.Join(dataDir, "spillcheck.pid"))
	return tr.T("cli.stopped")
}

func apiCall(ctx context.Context, method, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api: instance returned %s", resp.Status)
	}
	return string(raw), nil
}

func latestRelease(ctx context.Context) (string, error) {
	url := os.Getenv("SPILLCHECK_UPDATE_URL")
	if url == "" {
		url = "https://api.github.com/repos/Hungleminh/agentspillcheck/releases/latest"
	}
	body, err := apiCall(ctx, http.MethodGet, url)
	if err != nil {
		return "", err
	}
	var decoded struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		return "", err
	}
	return decoded.TagName, nil
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
